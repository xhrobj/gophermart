package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
)

type stubBalanceRepo struct {
	getBalanceFunc      func(ctx context.Context, userID int64) (model.Balance, error)
	withdrawFunc        func(ctx context.Context, userID int64, orderNumber string, sum int64) error
	listWithdrawalsFunc func(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

func (s *stubBalanceRepo) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	if s.getBalanceFunc == nil {
		panic("unexpected call to stubBalanceRepo.GetBalance")
	}
	return s.getBalanceFunc(ctx, userID)
}

func (s *stubBalanceRepo) Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error {
	if s.withdrawFunc == nil {
		panic("unexpected call to stubBalanceRepo.Withdraw")
	}
	return s.withdrawFunc(ctx, userID, orderNumber, sum)
}

func (s *stubBalanceRepo) ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	if s.listWithdrawalsFunc == nil {
		panic("unexpected call to stubBalanceRepo.ListWithdrawals")
	}
	return s.listWithdrawalsFunc(ctx, userID)
}

func TestBalanceService_GetBalance_OK(t *testing.T) {
	t.Parallel()

	want := model.Balance{
		Current:   50050,
		Withdrawn: 4200,
	}

	repo := &stubBalanceRepo{
		getBalanceFunc: func(ctx context.Context, userID int64) (model.Balance, error) {
			require.Equal(t, int64(42), userID)
			return want, nil
		},
	}

	svc := NewBalanceService(repo)

	got, err := svc.GetBalance(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestBalanceService_GetBalance_RepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("db failed")

	repo := &stubBalanceRepo{
		getBalanceFunc: func(ctx context.Context, userID int64) (model.Balance, error) {
			require.Equal(t, int64(42), userID)
			return model.Balance{}, expectedErr
		},
	}

	svc := NewBalanceService(repo)

	_, err := svc.GetBalance(context.Background(), 42)
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "get balance")
}

func TestBalanceService_Withdraw_OK(t *testing.T) {
	t.Parallel()

	repo := &stubBalanceRepo{
		withdrawFunc: func(ctx context.Context, userID int64, orderNumber string, sum int64) error {
			require.Equal(t, int64(42), userID)
			require.Equal(t, "2377225624", orderNumber)
			require.Equal(t, int64(511), sum)

			return nil
		},
	}

	svc := NewBalanceService(repo)

	err := svc.Withdraw(context.Background(), 42, "2377225624", 511)

	require.NoError(t, err)
}

func TestBalanceService_Withdraw_InvalidOrderNumber(t *testing.T) {
	t.Parallel()

	repo := &stubBalanceRepo{}
	svc := NewBalanceService(repo)

	err := svc.Withdraw(context.Background(), 42, "2377225625", 511)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidWithdrawOrderNumber)
}

func TestBalanceService_Withdraw_InvalidSum(t *testing.T) {
	t.Parallel()

	repo := &stubBalanceRepo{}
	svc := NewBalanceService(repo)

	err := svc.Withdraw(context.Background(), 42, "2377225624", 0)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidWithdrawSum)
}

func TestBalanceService_Withdraw_InsufficientFunds(t *testing.T) {
	t.Parallel()

	repo := &stubBalanceRepo{
		withdrawFunc: func(ctx context.Context, userID int64, orderNumber string, sum int64) error {
			require.Equal(t, int64(42), userID)
			require.Equal(t, "2377225624", orderNumber)
			require.Equal(t, int64(511), sum)

			return repository.ErrInsufficientFunds
		},
	}

	svc := NewBalanceService(repo)

	err := svc.Withdraw(context.Background(), 42, "2377225624", 511)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInsufficientFunds)
}

func TestBalanceService_ListWithdrawals_OK(t *testing.T) {
	t.Parallel()

	want := []model.Withdrawal{
		{
			ID:          1,
			UserID:      42,
			OrderNumber: "12345678903",
			Sum:         51100,
			ProcessedAt: time.Date(2026, 4, 19, 20, 0, 0, 0, time.UTC),
		},
	}

	repo := &stubBalanceRepo{
		listWithdrawalsFunc: func(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
			require.Equal(t, int64(42), userID)
			return want, nil
		},
	}

	svc := NewBalanceService(repo)

	got, err := svc.ListWithdrawals(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestBalanceService_ListWithdrawals_RepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("db failed")

	repo := &stubBalanceRepo{
		listWithdrawalsFunc: func(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
			return nil, expectedErr
		},
	}

	svc := NewBalanceService(repo)

	_, err := svc.ListWithdrawals(context.Background(), 42)
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "list withdrawals")
}
