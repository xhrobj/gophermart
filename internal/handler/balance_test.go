package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/service"
)

func TestGetBalance_OK(t *testing.T) {
	balanceService := &stubBalanceService{
		getBalanceFunc: func(ctx context.Context, userID int64) (model.Balance, error) {
			require.Equal(t, handlerCurrentUserID, userID)

			return model.Balance{
				Current:   50050,
				Withdrawn: 4200,
			}, nil
		},
	}

	rq := newAuthorizedRequest(http.MethodGet, "/api/user/balance", "")

	rs := serveWithAuth(t, GetBalance(balanceService), rq)

	require.Equal(t, http.StatusOK, rs.StatusCode)
	require.Equal(t, "application/json", rs.Header.Get("Content-Type"))

	var got getBalanceResponse
	err := json.NewDecoder(rs.Body).Decode(&got)
	require.NoError(t, err)

	require.InDelta(t, 500.5, got.Current, 0.000001)
	require.InDelta(t, 42.0, got.Withdrawn, 0.000001)
}

func TestWithdraw_OK(t *testing.T) {
	balanceService := &stubBalanceService{
		withdrawFunc: func(ctx context.Context, userID int64, orderNumber string, sum int64) error {
			require.Equal(t, handlerCurrentUserID, userID)
			require.Equal(t, validLuhnOrderNumber, orderNumber)
			require.Equal(t, int64(511), sum)

			return nil
		},
	}

	rq := newAuthorizedRequest(
		http.MethodPost,
		"/api/user/balance/withdraw",
		`{"order":"`+validLuhnOrderNumber+`","sum":5.11}`,
	)
	rq.Header.Set("Content-Type", "application/json")

	rs := serveWithAuth(t, Withdraw(balanceService), rq)

	require.Equal(t, http.StatusOK, rs.StatusCode)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	balanceService := &stubBalanceService{
		withdrawFunc: func(ctx context.Context, userID int64, orderNumber string, sum int64) error {
			require.Equal(t, handlerCurrentUserID, userID)
			require.Equal(t, validLuhnOrderNumber, orderNumber)
			require.Equal(t, int64(511), sum)

			return service.ErrInsufficientFunds
		},
	}

	rq := newAuthorizedRequest(
		http.MethodPost,
		"/api/user/balance/withdraw",
		`{"order":"`+validLuhnOrderNumber+`","sum":5.11}`,
	)
	rq.Header.Set("Content-Type", "application/json")

	rs := serveWithAuth(t, Withdraw(balanceService), rq)

	require.Equal(t, http.StatusPaymentRequired, rs.StatusCode)
}

func TestGetWithdrawals_NoContent(t *testing.T) {
	balanceService := &stubBalanceService{
		listWithdrawalsFunc: func(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
			require.Equal(t, handlerCurrentUserID, userID)

			return []model.Withdrawal{}, nil
		},
	}

	rq := newAuthorizedRequest(http.MethodGet, "/api/user/withdrawals", "")

	rs := serveWithAuth(t, GetWithdrawals(balanceService), rq)

	body, err := io.ReadAll(rs.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusNoContent, rs.StatusCode)
	require.Empty(t, body)
}

func TestGetWithdrawals_OK(t *testing.T) {
	loc := time.FixedZone("+0300", 3*60*60)

	balanceService := &stubBalanceService{
		listWithdrawalsFunc: func(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
			require.Equal(t, handlerCurrentUserID, userID)

			return []model.Withdrawal{
				{
					ID:          1,
					UserID:      handlerCurrentUserID,
					OrderNumber: validLuhnOrderNumber,
					Sum:         51100,
					ProcessedAt: time.Date(2026, 4, 18, 15, 15, 45, 0, loc),
				},
				{
					ID:          2,
					UserID:      handlerCurrentUserID,
					OrderNumber: anotherValidLuhnOrderNumber,
					Sum:         7500,
					ProcessedAt: time.Date(2026, 4, 17, 16, 9, 57, 0, loc),
				},
			}, nil
		},
	}

	rq := newAuthorizedRequest(http.MethodGet, "/api/user/withdrawals", "")

	rs := serveWithAuth(t, GetWithdrawals(balanceService), rq)

	require.Equal(t, http.StatusOK, rs.StatusCode)
	require.Equal(t, "application/json", rs.Header.Get("Content-Type"))

	var got []getWithdrawalsResponseItem
	err := json.NewDecoder(rs.Body).Decode(&got)
	require.NoError(t, err)

	require.Len(t, got, 2)

	require.Equal(t, validLuhnOrderNumber, got[0].Order)
	require.InDelta(t, 511.0, got[0].Sum, 0.000001)
	require.Equal(t, "2026-04-18T15:15:45+03:00", got[0].ProcessedAt)

	require.Equal(t, anotherValidLuhnOrderNumber, got[1].Order)
	require.InDelta(t, 75.0, got[1].Sum, 0.000001)
	require.Equal(t, "2026-04-17T16:09:57+03:00", got[1].ProcessedAt)
}
