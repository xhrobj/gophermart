package service

import (
	"context"
	"fmt"

	"github.com/xhrobj/gophermart/internal/accrual"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
	"go.uber.org/zap"
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

	orders, err := s.orderRepo.ListPending(ctx, 0)
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

		status := model.OrderStatusNew
		accrual := int64(0)

		switch result.Status {
		case model.AccrualStatusRegistered:
			status = model.OrderStatusProcessing

		case model.AccrualStatusProcessing:
			status = model.OrderStatusProcessing

		case model.AccrualStatusInvalid:
			status = model.OrderStatusInvalid

		case model.AccrualStatusProcessed:
			status = model.OrderStatusProcessed
			accrual = result.Accrual
		}

		// - вызываем SetAccrualResult
		//     -> SetAccrualResult сохраняет результат проверки заказа во внешнем сервисе начислений

		err = s.orderRepo.SetAccrualResult(ctx, order.Number, status, accrual)
		if err != nil {
			return fmt.Errorf("set accrual result for order %s: %w", order.Number, err)
		}
	}

	return nil
}
