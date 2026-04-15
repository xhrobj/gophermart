package service

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

type OrderService interface {
	UploadOrder(ctx context.Context, userID int64, number string) (model.UploadOrderResult, error)
	ListOrders(ctx context.Context, userID int64) ([]model.Order, error)
}
