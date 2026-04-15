package service

import "context"

// AccrualService описывает обработку заказов через внешний сервис начислений.
type AccrualService interface {
	// ProcessPendingOrders обрабатывает заказы, ожидающие проверки во внешнем сервисе начислений.
	ProcessPendingOrders(ctx context.Context) error
}
