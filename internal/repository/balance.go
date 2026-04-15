package repository

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

// BalanceRepository описывает операции хранения баланса и списаний.
type BalanceRepository interface {
	// GetBalance возвращает текущий баланс и сумму всех списаний пользователя.
	GetBalance(ctx context.Context, userID int64) (model.Balance, error)

	// Withdraw создает операцию списания баллов.
	//
	// Сумма списания передается в копейках.
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error

	// ListWithdrawals возвращает список списаний пользователя.
	ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}
