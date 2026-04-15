package model

import "time"

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         int64
	Number     string
	UserID     int64
	Status     OrderStatus
	Accrual    int64
	UploadedAt time.Time
}

type UploadOrderStatus string

const (
	// Новый заказ, HTTP 202
	UploadOrderAccepted UploadOrderStatus = "ACCEPTED"

	// Уже загружен этим пользователем, HTTP 200
	UploadOrderDuplicate UploadOrderStatus = "DUPLICATE"

	// Уже загружен другим пользователем, HTTP 409
	UploadOrderConflict UploadOrderStatus = "CONFLICT"
)

type UploadOrderResult struct {
	Status UploadOrderStatus
	Order  Order
}
