package accrual

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/model"
)

const accrualTestOrderNumber = "12345678903"

func TestHTTPClient_FetchOrderAccrual_OKWithAccrual(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/orders/"+accrualTestOrderNumber, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"order":"12345678903","status":"PROCESSED","accrual":500.5}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	got, err := client.FetchOrderAccrual(context.Background(), accrualTestOrderNumber)

	require.NoError(t, err)
	require.Equal(t, model.AccrualResult{
		Order:   accrualTestOrderNumber,
		Status:  model.AccrualStatusProcessed,
		Accrual: 50050,
	}, got)
}

func TestHTTPClient_FetchOrderAccrual_OKWithoutAccrual(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/orders/"+accrualTestOrderNumber, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"order":"12345678903","status":"PROCESSING"}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	got, err := client.FetchOrderAccrual(context.Background(), accrualTestOrderNumber)

	require.NoError(t, err)
	require.Equal(t, model.AccrualResult{
		Order:  accrualTestOrderNumber,
		Status: model.AccrualStatusProcessing,
	}, got)
}

func TestHTTPClient_FetchOrderAccrual_OKWithoutOrderInResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/orders/"+accrualTestOrderNumber, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"status":"REGISTERED"}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	got, err := client.FetchOrderAccrual(context.Background(), accrualTestOrderNumber)

	require.NoError(t, err)
	require.Equal(t, model.AccrualResult{
		Order:  accrualTestOrderNumber,
		Status: model.AccrualStatusRegistered,
	}, got)
}

func TestHTTPClient_FetchOrderAccrual_NoContent(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/orders/"+accrualTestOrderNumber, r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.FetchOrderAccrual(context.Background(), accrualTestOrderNumber)

	require.ErrorIs(t, err, ErrOrderNotRegistered)
}

func TestHTTPClient_FetchOrderAccrual_RateLimited(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/orders/"+accrualTestOrderNumber, r.URL.Path)

		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.FetchOrderAccrual(context.Background(), accrualTestOrderNumber)

	require.ErrorIs(t, err, ErrRateLimited)

	var rateLimitErr *RateLimitError
	require.True(t, errors.As(err, &rateLimitErr))
	require.Equal(t, time.Minute, rateLimitErr.RetryAfter)
}

func TestHTTPClient_FetchOrderAccrual_UnexpectedStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/orders/"+accrualTestOrderNumber, r.URL.Path)

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.FetchOrderAccrual(context.Background(), accrualTestOrderNumber)

	require.ErrorContains(t, err, "unexpected response status: 500")
}

func TestHTTPClient_FetchOrderAccrual_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/orders/"+accrualTestOrderNumber, r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.FetchOrderAccrual(context.Background(), accrualTestOrderNumber)

	require.ErrorContains(t, err, "decode accrual response")
}

func TestHTTPClient_FetchOrderAccrual_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClient("http://127.0.0.1")

	_, err := client.FetchOrderAccrual(ctx, accrualTestOrderNumber)

	require.ErrorContains(t, err, "perform accrual request")
}

func TestAmountToHundredths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		amount float64
		want   int64
	}{
		{
			name:   "integer",
			amount: 500,
			want:   50000,
		},
		{
			name:   "fractional",
			amount: 500.5,
			want:   50050,
		},
		{
			name:   "rounded",
			amount: 42.129,
			want:   4213,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, amountToHundredths(tt.amount))
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	future := now.Add(2 * time.Minute).Format(http.TimeFormat)
	past := now.Add(-2 * time.Minute).Format(http.TimeFormat)

	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{
			name:  "empty",
			value: "",
			want:  0,
		},
		{
			name:  "seconds",
			value: "60",
			want:  time.Minute,
		},
		{
			name:  "future http date",
			value: future,
			want:  2 * time.Minute,
		},
		{
			name:  "past http date",
			value: past,
			want:  0,
		},
		{
			name:  "invalid",
			value: "nope",
			want:  0,
		},
		{
			name:  "negative seconds",
			value: "-1",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, parseRetryAfter(tt.value, now))
		})
	}
}
