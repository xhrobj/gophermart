package model

import "time"

// Withdrawal описывает операцию списания баллов с бонусного счета пользователя.
type Withdrawal struct {
	ID          int64
	UserID      int64
	OrderNumber string

	// Sum содержит сумму списания в копейках.
	Sum int64

	ProcessedAt time.Time
}
