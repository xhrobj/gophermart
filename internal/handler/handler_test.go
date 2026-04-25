package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/middleware"
	"github.com/xhrobj/gophermart/internal/model"
)

const (
	validLuhnOrderNumber        = "12345678903"
	anotherValidLuhnOrderNumber = "9278923470"
	handlerGoodToken            = "good-token"
	handlerCurrentUserID        = int64(1)
	handlerIssuedToken          = "jwt-token"
)

type stubAuthService struct {
	registerFunc func(ctx context.Context, login, password string) (model.AuthResult, error)
	loginFunc    func(ctx context.Context, login, password string) (model.AuthResult, error)
}

func (s *stubAuthService) Register(ctx context.Context, login, password string) (model.AuthResult, error) {
	if s.registerFunc == nil {
		panic("unexpected call to stubAuthService.Register")
	}

	return s.registerFunc(ctx, login, password)
}

func (s *stubAuthService) Login(ctx context.Context, login, password string) (model.AuthResult, error) {
	if s.loginFunc == nil {
		panic("unexpected call to stubAuthService.Login")
	}

	return s.loginFunc(ctx, login, password)
}

type stubOrderService struct {
	uploadOrderFunc func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error)
	listOrdersFunc  func(ctx context.Context, userID int64) ([]model.Order, error)
}

func (s *stubOrderService) UploadOrder(
	ctx context.Context,
	userID int64,
	orderNumber string,
) (model.UploadOrderResult, error) {
	if s.uploadOrderFunc == nil {
		panic("unexpected call to stubOrderService.UploadOrder")
	}

	return s.uploadOrderFunc(ctx, userID, orderNumber)
}

func (s *stubOrderService) ListOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	if s.listOrdersFunc == nil {
		panic("unexpected call to stubOrderService.ListOrders")
	}

	return s.listOrdersFunc(ctx, userID)
}

type stubBalanceService struct {
	getBalanceFunc      func(ctx context.Context, userID int64) (model.Balance, error)
	withdrawFunc        func(ctx context.Context, userID int64, orderNumber string, sum int64) error
	listWithdrawalsFunc func(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

func (s *stubBalanceService) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	if s.getBalanceFunc == nil {
		panic("unexpected call to stubBalanceService.GetBalance")
	}

	return s.getBalanceFunc(ctx, userID)
}

func (s *stubBalanceService) Withdraw(
	ctx context.Context,
	userID int64,
	orderNumber string,
	sum int64,
) error {
	if s.withdrawFunc == nil {
		panic("unexpected call to stubBalanceService.Withdraw")
	}

	return s.withdrawFunc(ctx, userID, orderNumber, sum)
}

func (s *stubBalanceService) ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	if s.listWithdrawalsFunc == nil {
		panic("unexpected call to stubBalanceService.ListWithdrawals")
	}

	return s.listWithdrawalsFunc(ctx, userID)
}

type stubTokenManager struct {
	generateFunc func(userID int64) (string, error)
	parseFunc    func(token string) (int64, error)
}

func (s *stubTokenManager) Generate(userID int64) (string, error) {
	if s.generateFunc == nil {
		panic("unexpected call to stubTokenManager.Generate")
	}

	return s.generateFunc(userID)
}

func (s *stubTokenManager) Parse(token string) (int64, error) {
	if s.parseFunc == nil {
		panic("unexpected call to stubTokenManager.Parse")
	}

	return s.parseFunc(token)
}

func newAuthorizedRequest(method, target, body string) *http.Request {
	rq := httptest.NewRequest(method, target, strings.NewReader(body))
	rq.Header.Set("Authorization", "Bearer "+handlerGoodToken)

	return rq
}

func serveWithAuth(t *testing.T, h http.Handler, rq *http.Request) *http.Response {
	t.Helper()

	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, handlerGoodToken, token)

			return handlerCurrentUserID, nil
		},
	}

	rec := httptest.NewRecorder()
	middleware.WithAuth(tokenManager)(h).ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	return rs
}
