package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/xhrobj/gophermart/internal/accrual"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
	"go.uber.org/zap"
)

const (
	pendingOrdersSize = 100
	nextPollDelay     = time.Second * 30
)

// AccrualService описывает обработку заказов через внешний сервис начислений.
type AccrualService interface {
	// ProcessPendingOrders обрабатывает заказы, ожидающие проверки во внешнем сервисе начислений.
	ProcessPendingOrders(ctx context.Context) error
}

type accrualService struct {
	orderRepo     repository.OrderRepository
	accrualClient accrual.Client
	logger        *zap.Logger
}

func NewAccrualService(
	orderRepo repository.OrderRepository,
	accrualClient accrual.Client,
	logger *zap.Logger,
) AccrualService {
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

	// вызываем repository.ListPending
	//   -> ListPending возвращает заказы, ожидающие проверки во внешнем сервисе начислений

	orders, err := s.orderRepo.ListPending(ctx, pendingOrdersSize)
	if err != nil {
		return fmt.Errorf("list pending orders: %w", err)
	}

	// для каждого:

	for _, order := range orders {
		if err = ctx.Err(); err != nil {
			return err
		}

		// TODO: проверяем cooldown/rate limit

		// - вызываем внешний accrual

		result, fetchErr := s.accrualClient.FetchOrderAccrual(ctx, order.Number)
		// обычная ошибка -> log and continue
		if fetchErr != nil {
			if errors.Is(fetchErr, accrual.ErrOrderNotRegistered) {
				update := repository.OrderAccrualUpdate{
					Status:     model.OrderStatusNew,
					Accrual:    0,
					NextPollAt: time.Now().UTC().Add(nextPollDelay),
				}

				err = s.orderRepo.SetAccrualResult(ctx, order.Number, update)
				if err != nil {
					return fmt.Errorf("set accrual result for unregistered order %s: %w", order.Number, err)
				}

				continue
			}

			s.logger.Warn(
				"fetch accrual result failed",
				zap.String("order_number", order.Number),
				zap.Error(fetchErr),
			)
			continue
		}

		// - определяем новый статус
		//     - 204 -> NEW, next poll later
		//     - REGISTERED -> PROCESSING, next poll later
		//     - PROCESSING -> PROCESSING, next poll later
		//     - INVALID -> INVALID, final
		//     - PROCESSED -> PROCESSED + accrual, final
		// TODO: - 429 -> cooldown + stop pass

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
		}

		// - вызываем SetAccrualResult
		//     -> SetAccrualResult сохраняет результат проверки заказа во внешнем сервисе начислений

		err = s.orderRepo.SetAccrualResult(ctx, order.Number, update)
		if err != nil {
			return fmt.Errorf("set accrual result for order %s: %w", order.Number, err)
		}
	}

	return nil
}
