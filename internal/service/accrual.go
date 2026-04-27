package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xhrobj/gophermart/internal/accrual"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
	"go.uber.org/zap"
)

const (
	pendingOrdersSize  = 100
	accrualWorkerCount = 3
	nextPollDelay      = time.Second * 30
	defaultRetryAfter  = time.Second * 60
)

// PendingOrdersProcessor описывает обработку заказов через внешний сервис начислений.
type PendingOrdersProcessor interface {
	// ProcessPendingOrders обрабатывает заказы, ожидающие проверки во внешнем сервисе начислений.
	ProcessPendingOrders(ctx context.Context) error
}

type accrualService struct {
	orderRepo     repository.OrderRepository
	accrualClient accrual.OrderAccrualFetcher
	logger        *zap.Logger

	rateLimitMu    sync.Mutex
	rateLimitUntil time.Time
}

// NewAccrualService создаёт сервис обработки заказов через внешний сервис начислений.
func NewAccrualService(
	orderRepo repository.OrderRepository,
	accrualClient accrual.OrderAccrualFetcher,
	logger *zap.Logger,
) PendingOrdersProcessor {
	return &accrualService{
		orderRepo:     orderRepo,
		accrualClient: accrualClient,
		logger:        logger,
	}
}

func (s *accrualService) ProcessPendingOrders(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if delay := s.rateLimitDelay(time.Now().UTC()); delay > 0 {
		s.logger.Info(
			"(-_-)Zzz skip accrual polling because service is rate limited",
			zap.Duration("retry_after", delay),
		)
		return nil
	}

	orders, err := s.orderRepo.ListPending(ctx, pendingOrdersSize)
	if err != nil {
		return fmt.Errorf("list pending orders: %w", err)
	}

	if len(orders) == 0 {
		return nil
	}

	jobs := make(chan model.Order)
	errs := make(chan error, len(orders))
	stop := newOnceSignal()

	var wg sync.WaitGroup

	workerCount := min(len(orders), accrualWorkerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			s.runAccrualWorker(ctx, jobs, errs, stop)
		}()
	}

feedLoop:
	for _, order := range orders {
		select {
		case <-ctx.Done():
			break feedLoop
		case <-stop.Done():
			break feedLoop
		case jobs <- order:
		}
	}

	close(jobs)
	wg.Wait()
	close(errs)

	if err := ctx.Err(); err != nil {
		return err
	}

	return firstError(errs)
}

func (s *accrualService) runAccrualWorker(
	ctx context.Context,
	jobs <-chan model.Order,
	errs chan<- error,
	stop *onceSignal,
) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-stop.Done():
			return

		case order, ok := <-jobs:
			if !ok {
				return
			}

			if delay := s.rateLimitDelay(time.Now().UTC()); delay > 0 {
				s.logger.Info(
					"(-_-)Zzz stop accrual polling pass because service is rate limited",
					zap.Duration("retry_after", delay),
				)
				stop.Signal()
				return
			}

			if err := s.processPendingOrder(ctx, order); err != nil {
				if errors.Is(err, accrual.ErrRateLimited) {
					stop.Signal()
					return
				}

				errs <- err
			}
		}
	}
}

func (s *accrualService) processPendingOrder(ctx context.Context, order model.Order) error {
	result, fetchErr := s.accrualClient.FetchOrderAccrual(ctx, order.Number)
	if fetchErr != nil {
		var rateLimitErr *accrual.RateLimitError
		if errors.As(fetchErr, &rateLimitErr) {
			retryAfter := rateLimitErr.RetryAfter
			if retryAfter <= 0 {
				retryAfter = defaultRetryAfter
			}

			s.setRateLimit(time.Now().UTC(), retryAfter)

			s.logger.Warn(
				"(-_-)Zzz accrual rate limit received; pause polling",
				zap.String("order_number", order.Number),
				zap.Duration("retry_after", retryAfter),
			)

			return fetchErr
		}

		if errors.Is(fetchErr, accrual.ErrOrderNotRegistered) {
			update := repository.OrderAccrualUpdate{
				Status:     model.OrderStatusNew,
				Accrual:    0,
				NextPollAt: time.Now().UTC().Add(nextPollDelay),
			}

			err := s.orderRepo.SetAccrualResult(ctx, order.Number, update)
			if err != nil {
				return fmt.Errorf("set accrual result for unregistered order %s: %w", order.Number, err)
			}

			return nil
		}

		s.logger.Warn(
			"fetch accrual result failed",
			zap.String("order_number", order.Number),
			zap.Error(fetchErr),
		)

		return nil
	}

	update := repository.OrderAccrualUpdate{
		Status:     model.OrderStatusNew,
		Accrual:    0,
		NextPollAt: time.Now().UTC().Add(nextPollDelay),
	}

	switch result.Status {
	case model.AccrualStatusRegistered, model.AccrualStatusProcessing:
		update.Status = model.OrderStatusProcessing

	case model.AccrualStatusInvalid:
		update.Status = model.OrderStatusInvalid

	case model.AccrualStatusProcessed:
		update.Status = model.OrderStatusProcessed
		update.Accrual = result.Accrual

	default:
		s.logger.Warn(
			"unknown accrual status",
			zap.String("order_number", order.Number),
			zap.String("status", string(result.Status)),
		)
		return nil
	}

	err := s.orderRepo.SetAccrualResult(ctx, order.Number, update)
	if err != nil {
		return fmt.Errorf("set accrual result for order %s: %w", order.Number, err)
	}

	return nil
}

func (s *accrualService) rateLimitDelay(now time.Time) time.Duration {
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()

	if s.rateLimitUntil.After(now) {
		return s.rateLimitUntil.Sub(now)
	}

	return 0
}

func (s *accrualService) setRateLimit(now time.Time, retryAfter time.Duration) {
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()

	until := now.Add(retryAfter)
	if until.After(s.rateLimitUntil) {
		s.rateLimitUntil = until
	}
}

func firstError(errs <-chan error) error {
	var first error

	for err := range errs {
		if err == nil {
			continue
		}

		if first == nil {
			first = err
		}
	}

	return first
}

type onceSignal struct {
	once sync.Once
	ch   chan struct{}
}

func newOnceSignal() *onceSignal {
	return &onceSignal{ch: make(chan struct{})}
}

func (s *onceSignal) Done() <-chan struct{} {
	return s.ch
}

func (s *onceSignal) Signal() {
	s.once.Do(func() {
		close(s.ch)
	})
}
