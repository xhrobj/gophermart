package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/accrual"
	"github.com/xhrobj/gophermart/internal/model"
	"go.uber.org/zap"
)

type stubAccrualClient struct {
	fetchOrderAccrualFunc func(ctx context.Context, orderNumber string) (model.AccrualResult, error)
}

func (s *stubAccrualClient) FetchOrderAccrual(
	ctx context.Context,
	orderNumber string,
) (model.AccrualResult, error) {
	if s.fetchOrderAccrualFunc == nil {
		panic("unexpected call to stubAccrualClient.FetchOrderAccrual")
	}

	return s.fetchOrderAccrualFunc(ctx, orderNumber)
}

func TestAccrualService_ProcessPendingOrders_RateLimitSkipsNextPass(t *testing.T) {
	orderNumber := "12345678903"

	listPendingCalls := 0
	fetchCalls := 0

	orderRepo := &stubOrderRepo{
		listPendingFunc: func(ctx context.Context, limit int) ([]model.Order, error) {
			listPendingCalls++

			require.Equal(t, pendingOrdersSize, limit)

			return []model.Order{
				{
					Number: orderNumber,
					Status: model.OrderStatusNew,
				},
			}, nil
		},
	}

	accrualClient := &stubAccrualClient{
		fetchOrderAccrualFunc: func(ctx context.Context, gotOrderNumber string) (model.AccrualResult, error) {
			fetchCalls++

			require.Equal(t, orderNumber, gotOrderNumber)

			return model.AccrualResult{}, &accrual.RateLimitError{
				RetryAfter: time.Minute,
			}
		},
	}

	svc := NewAccrualService(orderRepo, accrualClient, zap.NewNop())

	err := svc.ProcessPendingOrders(context.Background())
	require.NoError(t, err)

	err = svc.ProcessPendingOrders(context.Background())
	require.NoError(t, err)

	require.Equal(t, 1, listPendingCalls)
	require.Equal(t, 1, fetchCalls)
}
