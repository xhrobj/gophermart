package model

import "time"

// OrderStatus описывает статус обработки заказа в системе Gophermart.
type OrderStatus string

const (
	// OrderStatusNew - промежуточный статус:
	// заказ загружен, но еще не отправлен в обработку во внешний сервис.
	OrderStatusNew OrderStatus = "NEW"

	// OrderStatusProcessing - промежуточный статус:
	// заказ отправлен во внешний сервис и находится в обработке.
	OrderStatusProcessing OrderStatus = "PROCESSING"

	// OrderStatusInvalid - финальный статус: внешний сервис отказал в расчете начисления.
	OrderStatusInvalid OrderStatus = "INVALID"

	// OrderStatusProcessed - финальный статус:
	// расчет внешним сервисом завершен и результат начисления получен.
	OrderStatusProcessed OrderStatus = "PROCESSED"
)

type Order struct {
	ID         int64
	Number     string
	UserID     int64
	Status     OrderStatus
	Accrual    int64
	UploadedAt time.Time
}

// UploadOrderStatus описывает бизнес-результат попытки загрузить заказ.
type UploadOrderStatus string

const (
	// Новый заказ, HTTP 202
	UploadOrderAccepted UploadOrderStatus = "ACCEPTED"

	// Уже загружен этим пользователем, HTTP 200
	UploadOrderDuplicate UploadOrderStatus = "DUPLICATE"

	// Уже загружен другим пользователем, HTTP 409
	UploadOrderConflict UploadOrderStatus = "CONFLICT"
)

// UploadOrderResult содержит результат попытки загрузить заказ.
type UploadOrderResult struct {
	Status UploadOrderStatus
	Order  Order
}
