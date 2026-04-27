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

// Order описывает заказ, загруженный пользователем в систему Gophermart.
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
	// UploadOrderAccepted означает, что заказ новый и принят в обработку.
	UploadOrderAccepted UploadOrderStatus = "ACCEPTED"

	// UploadOrderDuplicate означает, что этот же пользователь уже загружал заказ.
	UploadOrderDuplicate UploadOrderStatus = "DUPLICATE"

	// UploadOrderConflict означает, что заказ уже был загружен другим пользователем.
	UploadOrderConflict UploadOrderStatus = "CONFLICT"
)

// UploadOrderResult содержит результат попытки загрузить заказ.
type UploadOrderResult struct {
	Status UploadOrderStatus
	Order  Order
}
