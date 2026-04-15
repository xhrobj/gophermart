package repository

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

// OrderRepository описывает операции хранения заказов.
type OrderRepository interface {
	// Create создает новый заказ пользователя.
	Create(ctx context.Context, userID int64, orderNumber string) (model.Order, error)

	// FindByNumber возвращает заказ по номеру.
	FindByNumber(ctx context.Context, orderNumber string) (model.Order, error)

	// ListByUserID возвращает список заказов пользователя.
	ListByUserID(ctx context.Context, userID int64) ([]model.Order, error)

	// ListPending возвращает заказы, ожидающие проверки во внешнем сервисе начислений.
	ListPending(ctx context.Context, limit int) ([]model.Order, error)

	// SetAccrualResult сохраняет результат проверки заказа во внешнем сервисе начислений.
	SetAccrualResult(ctx context.Context, orderNumber string, status model.OrderStatus, accrual int64) error
}
