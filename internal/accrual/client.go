package accrual

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

// Client описывает клиент для обращения к внешнему сервису начислений.
type Client interface {
	// FetchOrderAccrual запрашивает во внешнем сервисе результат начисления по номеру заказа.
	FetchOrderAccrual(ctx context.Context, orderNumber string) (model.AccrualResult, error)
}
