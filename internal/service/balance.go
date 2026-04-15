package service

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID int64) (model.Balance, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error
	ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}
