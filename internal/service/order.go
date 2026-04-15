package service

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

// OrderService описывает операции с заказами пользователя.
type OrderService interface {
	// UploadOrder загружает номер заказа пользователя в систему.
	//
	// Для успешно разобранной бизнес-операции возвращает один из статусов: ACCEPTED, DUPLICATE или CONFLICT.
	// Некорректный номер заказа возвращается как ошибка.
	UploadOrder(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error)

	// ListOrders возвращает список заказов пользователя.
	ListOrders(ctx context.Context, userID int64) ([]model.Order, error)
}
