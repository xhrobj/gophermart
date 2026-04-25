package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/model"
)

func TestUploadOrder_Accepted(t *testing.T) {
	orderService := &stubOrderService{
		uploadOrderFunc: func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
			require.Equal(t, handlerCurrentUserID, userID)
			require.Equal(t, validLuhnOrderNumber, orderNumber)

			return model.UploadOrderResult{
				Status: model.UploadOrderAccepted,
			}, nil
		},
	}

	rq := newAuthorizedRequest(http.MethodPost, "/api/user/orders", validLuhnOrderNumber)

	rs := serveWithAuth(t, UploadOrder(orderService), rq)

	require.Equal(t, http.StatusAccepted, rs.StatusCode)
}

func TestUploadOrder_Duplicate(t *testing.T) {
	orderService := &stubOrderService{
		uploadOrderFunc: func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
			require.Equal(t, handlerCurrentUserID, userID)
			require.Equal(t, validLuhnOrderNumber, orderNumber)

			return model.UploadOrderResult{
				Status: model.UploadOrderDuplicate,
			}, nil
		},
	}

	rq := newAuthorizedRequest(http.MethodPost, "/api/user/orders", validLuhnOrderNumber)

	rs := serveWithAuth(t, UploadOrder(orderService), rq)

	require.Equal(t, http.StatusOK, rs.StatusCode)
}

func TestUploadOrder_Conflict(t *testing.T) {
	orderService := &stubOrderService{
		uploadOrderFunc: func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
			require.Equal(t, handlerCurrentUserID, userID)
			require.Equal(t, validLuhnOrderNumber, orderNumber)

			return model.UploadOrderResult{
				Status: model.UploadOrderConflict,
			}, nil
		},
	}

	rq := newAuthorizedRequest(http.MethodPost, "/api/user/orders", validLuhnOrderNumber)

	rs := serveWithAuth(t, UploadOrder(orderService), rq)

	require.Equal(t, http.StatusConflict, rs.StatusCode)
}

func TestGetOrders_NoContent(t *testing.T) {
	orderService := &stubOrderService{
		listOrdersFunc: func(ctx context.Context, userID int64) ([]model.Order, error) {
			require.Equal(t, handlerCurrentUserID, userID)

			return []model.Order{}, nil
		},
	}

	rq := newAuthorizedRequest(http.MethodGet, "/api/user/orders", "")

	rs := serveWithAuth(t, GetOrders(orderService), rq)

	require.Equal(t, http.StatusNoContent, rs.StatusCode)
}

func TestGetOrders_OK(t *testing.T) {
	loc := time.FixedZone("+0300", 3*60*60)

	orderService := &stubOrderService{
		listOrdersFunc: func(ctx context.Context, userID int64) ([]model.Order, error) {
			require.Equal(t, handlerCurrentUserID, userID)

			return []model.Order{
				{
					ID:         1,
					Number:     validLuhnOrderNumber,
					UserID:     handlerCurrentUserID,
					Status:     model.OrderStatusProcessed,
					Accrual:    50050,
					UploadedAt: time.Date(2026, 4, 18, 15, 15, 45, 0, loc),
				},
				{
					ID:         2,
					Number:     anotherValidLuhnOrderNumber,
					UserID:     handlerCurrentUserID,
					Status:     model.OrderStatusProcessing,
					Accrual:    0,
					UploadedAt: time.Date(2026, 4, 18, 15, 12, 1, 0, loc),
				},
			}, nil
		},
	}

	rq := newAuthorizedRequest(http.MethodGet, "/api/user/orders", "")

	rs := serveWithAuth(t, GetOrders(orderService), rq)

	require.Equal(t, http.StatusOK, rs.StatusCode)
	require.Equal(t, "application/json", rs.Header.Get("Content-Type"))

	var got []getOrdersResponseItem
	err := json.NewDecoder(rs.Body).Decode(&got)
	require.NoError(t, err)

	require.Len(t, got, 2)

	require.Equal(t, validLuhnOrderNumber, got[0].Number)
	require.Equal(t, "PROCESSED", got[0].Status)
	require.Equal(t, "2026-04-18T15:15:45+03:00", got[0].UploadedAt)
	require.NotNil(t, got[0].Accrual)
	require.InDelta(t, 500.5, *got[0].Accrual, 0.000001)

	require.Equal(t, anotherValidLuhnOrderNumber, got[1].Number)
	require.Equal(t, "PROCESSING", got[1].Status)
	require.Equal(t, "2026-04-18T15:12:01+03:00", got[1].UploadedAt)
	require.Nil(t, got[1].Accrual)
}
