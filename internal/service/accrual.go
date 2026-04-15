package service

import "context"

type AccrualService interface {
	ProcessPendingOrders(ctx context.Context) error
}
