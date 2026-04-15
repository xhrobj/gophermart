package service

import "context"

// AccrualService описывает обработку заказов через внешний сервис начислений.
type AccrualService interface {
	ProcessPendingOrders(ctx context.Context) error
}
