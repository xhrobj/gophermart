package service

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

// BalanceService описывает операции с бонусным счетом пользователя.
type BalanceService interface {
	// GetBalance возвращает текущее состояние бонусного счета пользователя.
	GetBalance(ctx context.Context, userID int64) (model.Balance, error)

	// Withdraw списывает баллы с бонусного счета пользователя в счет оплаты заказа с переданным номером.
	//
	// Сумма списания передается в копейках.
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error

	// ListWithdrawals возвращает список списаний пользователя.
	ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}
