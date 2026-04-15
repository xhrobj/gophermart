package service

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

// OrderService описывает операции с заказами пользователя.
type OrderService interface {
	// UploadOrder загружает номер заказа пользователя в систему.
	//
	// NOTE: UploadOrderStatus возвращаем для успешно разобранной бизнес-операции -> 202, 200, 409
	// для неверного формата номера заказа (422) <- вернем ошибку
	UploadOrder(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error)

	// ListOrders возвращает список заказов пользователя.
	ListOrders(ctx context.Context, userID int64) ([]model.Order, error)
}
