package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xhrobj/gophermart/internal/model"
)

var (
	// ErrOrderNotRegistered возвращается, когда заказ не зарегистрирован в accrual.
	ErrOrderNotRegistered = errors.New("order not registered in accrual system")

	// ErrRateLimited возвращается, когда accrual ограничил частоту запросов.
	ErrRateLimited = errors.New("accrual rate limit exceeded")
)

// RateLimitError описывает ответ accrual 429 Too Many Requests.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%v: retry after %s", ErrRateLimited, e.RetryAfter)
}

func (e *RateLimitError) Unwrap() error {
	return ErrRateLimited
}

// Client описывает клиент для обращения к внешнему сервису начислений.
type Client interface {
	// FetchOrderAccrual запрашивает во внешнем сервисе результат начисления по номеру заказа.
	FetchOrderAccrual(ctx context.Context, orderNumber string) (model.AccrualResult, error)
}

// HTTPClient реализует Client поверх HTTP.
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создает HTTP-клиент внешнего сервиса начислений.
func NewClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type fetchOrderAccrualResponse struct {
	Order   string              `json:"order"`
	Status  model.AccrualStatus `json:"status"`
	Accrual *float64            `json:"accrual,omitempty"`
}

func (c *HTTPClient) FetchOrderAccrual(ctx context.Context, orderNumber string) (model.AccrualResult, error) {
	requestURL := c.baseURL + "/api/orders/" + url.PathEscape(orderNumber)

	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return model.AccrualResult{}, fmt.Errorf("build accrual request: %w", err)
	}

	rs, err := c.httpClient.Do(rq)
	if err != nil {
		return model.AccrualResult{}, fmt.Errorf("perform accrual request: %w", err)
	}
	defer func() {
		_ = rs.Body.Close()
	}()

	switch rs.StatusCode {
	case http.StatusOK:
		var payload fetchOrderAccrualResponse

		if err = json.NewDecoder(rs.Body).Decode(&payload); err != nil {
			return model.AccrualResult{}, fmt.Errorf("decode accrual response: %w", err)
		}

		result := model.AccrualResult{
			Order:  payload.Order,
			Status: payload.Status,
		}

		if result.Order == "" {
			result.Order = orderNumber
		}

		if payload.Accrual != nil {
			result.Accrual = amountToHundredths(*payload.Accrual)
		}

		return result, nil

	case http.StatusNoContent:
		return model.AccrualResult{}, ErrOrderNotRegistered

	case http.StatusTooManyRequests:
		retryAfter := parseRetryAfter(rs.Header.Get("Retry-After"), time.Now())
		return model.AccrualResult{}, &RateLimitError{RetryAfter: retryAfter}

	default:
		return model.AccrualResult{}, fmt.Errorf("unexpected response status: %d", rs.StatusCode)
	}
}

func amountToHundredths(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	seconds, err := strconv.Atoi(value)
	if err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	retryAt, err := http.ParseTime(value)
	if err == nil && retryAt.After(now) {
		return retryAt.Sub(now)
	}

	return 0
}
