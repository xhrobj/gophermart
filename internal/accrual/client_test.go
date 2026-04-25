package accrual

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHTTPClient_FetchOrderAccrual_RateLimited(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/orders/12345678903", r.URL.Path)

		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.FetchOrderAccrual(context.Background(), "12345678903")

	require.ErrorIs(t, err, ErrRateLimited)

	var rateLimitErr *RateLimitError
	require.True(t, errors.As(err, &rateLimitErr))
	require.Equal(t, time.Minute, rateLimitErr.RetryAfter)
}
