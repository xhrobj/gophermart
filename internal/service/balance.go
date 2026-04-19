package service

import (
	"context"
	"fmt"

	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
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

type balanceService struct {
	balanceRepo repository.BalanceRepository
}

// NewBalanceService создаёт сервис работы с бонусным счётом пользователя.
func NewBalanceService(balanceRepo repository.BalanceRepository) BalanceService {
	return &balanceService{
		balanceRepo: balanceRepo,
	}
}

func (s *balanceService) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	balance, err := s.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		return model.Balance{}, fmt.Errorf("get balance: %w", err)
	}

	return balance, nil
}

func (s *balanceService) Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error {
	return s.balanceRepo.Withdraw(ctx, userID, orderNumber, sum)
}

func (s *balanceService) ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	withdrawals, err := s.balanceRepo.ListWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list withdrawals: %w", err)
	}

	return withdrawals, nil
}
