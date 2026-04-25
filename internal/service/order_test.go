package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
)

const (
	validOrderNumber        = "12345678903"
	anotherValidOrderNumber = "49927398716"
	invalidLuhnOrderNumber  = "12345678904"

	currentUserID = int64(42)
	otherUserID   = int64(69)
)

type stubOrderRepo struct {
	createFunc           func(ctx context.Context, userID int64, orderNumber string) (model.Order, error)
	findByNumberFunc     func(ctx context.Context, orderNumber string) (model.Order, error)
	listByUserIDFunc     func(ctx context.Context, userID int64) ([]model.Order, error)
	listPendingFunc      func(ctx context.Context, limit int) ([]model.Order, error)
	setAccrualResultFunc func(ctx context.Context, orderNumber string, update repository.OrderAccrualUpdate) error
}

func (s *stubOrderRepo) Create(ctx context.Context, userID int64, orderNumber string) (model.Order, error) {
	if s.createFunc == nil {
		panic("unexpected call to stubOrderRepo.Create")
	}

	return s.createFunc(ctx, userID, orderNumber)
}

func (s *stubOrderRepo) FindByNumber(ctx context.Context, orderNumber string) (model.Order, error) {
	if s.findByNumberFunc == nil {
		panic("unexpected call to stubOrderRepo.FindByNumber")
	}

	return s.findByNumberFunc(ctx, orderNumber)
}

func (s *stubOrderRepo) ListByUserID(ctx context.Context, userID int64) ([]model.Order, error) {
	if s.listByUserIDFunc == nil {
		panic("unexpected call to stubOrderRepo.ListByUserID")
	}

	return s.listByUserIDFunc(ctx, userID)
}

func (s *stubOrderRepo) ListPending(ctx context.Context, limit int) ([]model.Order, error) {
	if s.listPendingFunc == nil {
		panic("unexpected call to stubOrderRepo.ListPending")
	}

	return s.listPendingFunc(ctx, limit)
}

func (s *stubOrderRepo) SetAccrualResult(ctx context.Context, orderNumber string, update repository.OrderAccrualUpdate) error {
	if s.setAccrualResultFunc == nil {
		panic("unexpected call to stubOrderRepo.SetAccrualResult")
	}

	return s.setAccrualResultFunc(ctx, orderNumber, update)
}

func TestOrderService_UploadOrder_Accepted(t *testing.T) {
	t.Parallel()

	orderRepo := &stubOrderRepo{
		findByNumberFunc: func(ctx context.Context, orderNumber string) (model.Order, error) {
			require.Equal(t, validOrderNumber, orderNumber)
			return model.Order{}, repository.ErrOrderNotFound
		},
		createFunc: func(ctx context.Context, userID int64, orderNumber string) (model.Order, error) {
			require.Equal(t, currentUserID, userID)
			require.Equal(t, validOrderNumber, orderNumber)

			return model.Order{
				ID:      1,
				Number:  orderNumber,
				UserID:  userID,
				Status:  model.OrderStatusNew,
				Accrual: 0,
			}, nil
		},
	}

	svc := NewOrderService(orderRepo)

	got, err := svc.UploadOrder(context.Background(), currentUserID, " "+validOrderNumber+" ")
	require.NoError(t, err)
	require.Equal(t, model.UploadOrderResult{
		Status: model.UploadOrderAccepted,
		Order: model.Order{
			ID:      1,
			Number:  validOrderNumber,
			UserID:  currentUserID,
			Status:  model.OrderStatusNew,
			Accrual: 0,
		},
	}, got)
}

func TestOrderService_UploadOrder_InvalidOrderInput(t *testing.T) {
	t.Parallel()

	orderRepo := &stubOrderRepo{}
	svc := NewOrderService(orderRepo)

	_, err := svc.UploadOrder(context.Background(), currentUserID, "   ")
	require.ErrorIs(t, err, ErrInvalidOrderInput)
}

func TestOrderService_UploadOrder_InvalidOrderNumber_NonDigits(t *testing.T) {
	t.Parallel()

	orderRepo := &stubOrderRepo{}
	svc := NewOrderService(orderRepo)

	_, err := svc.UploadOrder(context.Background(), currentUserID, "12ab34")
	require.ErrorIs(t, err, ErrInvalidOrderNumber)
}

func TestOrderService_UploadOrder_InvalidOrderNumber_Luhn(t *testing.T) {
	t.Parallel()

	orderRepo := &stubOrderRepo{}
	svc := NewOrderService(orderRepo)

	_, err := svc.UploadOrder(context.Background(), currentUserID, invalidLuhnOrderNumber)
	require.ErrorIs(t, err, ErrInvalidOrderNumber)
}

func TestOrderService_UploadOrder_Duplicate(t *testing.T) {
	t.Parallel()

	orderRepo := &stubOrderRepo{
		findByNumberFunc: func(ctx context.Context, orderNumber string) (model.Order, error) {
			return model.Order{
				ID:      1,
				Number:  validOrderNumber,
				UserID:  currentUserID,
				Status:  model.OrderStatusNew,
				Accrual: 0,
			}, nil
		},
	}

	svc := NewOrderService(orderRepo)

	got, err := svc.UploadOrder(context.Background(), currentUserID, validOrderNumber)
	require.NoError(t, err)
	require.Equal(t, model.UploadOrderDuplicate, got.Status)
	require.Equal(t, currentUserID, got.Order.UserID)
}

func TestOrderService_UploadOrder_Conflict(t *testing.T) {
	t.Parallel()

	orderRepo := &stubOrderRepo{
		findByNumberFunc: func(ctx context.Context, orderNumber string) (model.Order, error) {
			return model.Order{
				ID:      1,
				Number:  validOrderNumber,
				UserID:  otherUserID,
				Status:  model.OrderStatusNew,
				Accrual: 0,
			}, nil
		},
	}

	svc := NewOrderService(orderRepo)

	got, err := svc.UploadOrder(context.Background(), currentUserID, validOrderNumber)
	require.NoError(t, err)
	require.Equal(t, model.UploadOrderConflict, got.Status)
	require.Equal(t, otherUserID, got.Order.UserID)
}

func TestOrderService_UploadOrder_CreateRace_Duplicate(t *testing.T) {
	t.Parallel()

	findCalls := 0

	orderRepo := &stubOrderRepo{
		findByNumberFunc: func(ctx context.Context, orderNumber string) (model.Order, error) {
			findCalls++

			if findCalls == 1 {
				return model.Order{}, repository.ErrOrderNotFound
			}

			return model.Order{
				ID:      1,
				Number:  validOrderNumber,
				UserID:  currentUserID,
				Status:  model.OrderStatusNew,
				Accrual: 0,
			}, nil
		},
		createFunc: func(ctx context.Context, userID int64, orderNumber string) (model.Order, error) {
			return model.Order{}, repository.ErrOrderAlreadyExists
		},
	}

	svc := NewOrderService(orderRepo)

	got, err := svc.UploadOrder(context.Background(), currentUserID, validOrderNumber)
	require.NoError(t, err)
	require.Equal(t, model.UploadOrderDuplicate, got.Status)
}

func TestOrderService_UploadOrder_CreateRace_Conflict(t *testing.T) {
	t.Parallel()

	findCalls := 0

	orderRepo := &stubOrderRepo{
		findByNumberFunc: func(ctx context.Context, orderNumber string) (model.Order, error) {
			findCalls++

			if findCalls == 1 {
				return model.Order{}, repository.ErrOrderNotFound
			}

			return model.Order{
				ID:      1,
				Number:  validOrderNumber,
				UserID:  otherUserID,
				Status:  model.OrderStatusNew,
				Accrual: 0,
			}, nil
		},
		createFunc: func(ctx context.Context, userID int64, orderNumber string) (model.Order, error) {
			return model.Order{}, repository.ErrOrderAlreadyExists
		},
	}

	svc := NewOrderService(orderRepo)

	got, err := svc.UploadOrder(context.Background(), currentUserID, validOrderNumber)
	require.NoError(t, err)
	require.Equal(t, model.UploadOrderConflict, got.Status)
}

func TestOrderService_ListOrders_OK(t *testing.T) {
	t.Parallel()

	want := []model.Order{
		{
			ID:      1,
			Number:  validOrderNumber,
			UserID:  currentUserID,
			Status:  model.OrderStatusNew,
			Accrual: 0,
		},
		{
			ID:      2,
			Number:  anotherValidOrderNumber,
			UserID:  currentUserID,
			Status:  model.OrderStatusProcessed,
			Accrual: 500,
		},
	}

	orderRepo := &stubOrderRepo{
		listByUserIDFunc: func(ctx context.Context, userID int64) ([]model.Order, error) {
			require.Equal(t, int64(currentUserID), userID)
			return want, nil
		},
	}

	svc := NewOrderService(orderRepo)

	got, err := svc.ListOrders(context.Background(), currentUserID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestOrderService_ListOrders_RepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("db failed")

	orderRepo := &stubOrderRepo{
		listByUserIDFunc: func(ctx context.Context, userID int64) ([]model.Order, error) {
			return nil, expectedErr
		},
	}

	svc := NewOrderService(orderRepo)

	_, err := svc.ListOrders(context.Background(), currentUserID)
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "list orders by user id")
}
