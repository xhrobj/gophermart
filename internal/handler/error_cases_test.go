package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/service"
)

func serveHandler(t *testing.T, h http.Handler, rq *http.Request) *http.Response {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	return rs
}

func TestHandlers_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		rq      *http.Request
	}{
		{
			name:    "Register",
			handler: Register(&stubAuthService{}),
			rq:      httptest.NewRequest(http.MethodGet, "/api/user/register", nil),
		},
		{
			name:    "Login",
			handler: Login(&stubAuthService{}),
			rq:      httptest.NewRequest(http.MethodGet, "/api/user/login", nil),
		},
		{
			name:    "UploadOrder",
			handler: UploadOrder(&stubOrderService{}),
			rq:      httptest.NewRequest(http.MethodGet, "/api/user/orders", nil),
		},
		{
			name:    "GetOrders",
			handler: GetOrders(&stubOrderService{}),
			rq:      httptest.NewRequest(http.MethodPost, "/api/user/orders", nil),
		},
		{
			name:    "GetBalance",
			handler: GetBalance(&stubBalanceService{}),
			rq:      httptest.NewRequest(http.MethodPost, "/api/user/balance", nil),
		},
		{
			name:    "Withdraw",
			handler: Withdraw(&stubBalanceService{}),
			rq:      httptest.NewRequest(http.MethodGet, "/api/user/balance/withdraw", nil),
		},
		{
			name:    "GetWithdrawals",
			handler: GetWithdrawals(&stubBalanceService{}),
			rq:      httptest.NewRequest(http.MethodPost, "/api/user/withdrawals", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := serveHandler(t, tt.handler, tt.rq)

			require.Equal(t, http.StatusMethodNotAllowed, rs.StatusCode)
		})
	}
}

func TestProtectedHandlers_Unauthorized(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		rq      *http.Request
	}{
		{
			name:    "UploadOrder",
			handler: UploadOrder(&stubOrderService{}),
			rq:      httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(validLuhnOrderNumber)),
		},
		{
			name:    "GetOrders",
			handler: GetOrders(&stubOrderService{}),
			rq:      httptest.NewRequest(http.MethodGet, "/api/user/orders", nil),
		},
		{
			name:    "GetBalance",
			handler: GetBalance(&stubBalanceService{}),
			rq:      httptest.NewRequest(http.MethodGet, "/api/user/balance", nil),
		},
		{
			name:    "Withdraw",
			handler: Withdraw(&stubBalanceService{}),
			rq: httptest.NewRequest(
				http.MethodPost,
				"/api/user/balance/withdraw",
				strings.NewReader(`{"order":"`+validLuhnOrderNumber+`","sum":5.11}`),
			),
		},
		{
			name:    "GetWithdrawals",
			handler: GetWithdrawals(&stubBalanceService{}),
			rq:      httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := serveHandler(t, tt.handler, tt.rq)

			require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
		})
	}
}

func TestRegister_ErrorCases(t *testing.T) {
	internalErr := errors.New("register failed")

	tests := []struct {
		name            string
		body            string
		wantServiceCall bool
		serviceErr      error
		wantStatus      int
		wantContentType string
		wantBodyPart    string
	}{
		{
			name:       "BadJSON",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:            "InvalidAuthInput",
			body:            `{"login":"","password":"secret"}`,
			wantServiceCall: true,
			serviceErr:      service.ErrInvalidAuthInput,
			wantStatus:      http.StatusBadRequest,
		},
		{
			name:            "PasswordTooLong",
			body:            `{"login":"admin","password":"secret"}`,
			wantServiceCall: true,
			serviceErr:      service.ErrPasswordTooLong,
			wantStatus:      http.StatusBadRequest,
			wantContentType: "text/plain; charset=utf-8",
			wantBodyPart:    "Пароль слишком длинный",
		},
		{
			name:            "InternalServerError",
			body:            `{"login":"admin","password":"secret"}`,
			wantServiceCall: true,
			serviceErr:      internalErr,
			wantStatus:      http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &stubAuthService{}

			if tt.wantServiceCall {
				authService.registerFunc = func(ctx context.Context, login, password string) (model.AuthResult, error) {
					return model.AuthResult{}, tt.serviceErr
				}
			}

			rq := httptest.NewRequest(
				http.MethodPost,
				"/api/user/register",
				strings.NewReader(tt.body),
			)

			rs := serveHandler(t, Register(authService), rq)

			require.Equal(t, tt.wantStatus, rs.StatusCode)

			if tt.wantContentType != "" {
				require.Equal(t, tt.wantContentType, rs.Header.Get("Content-Type"))
			}

			if tt.wantBodyPart != "" {
				body, err := io.ReadAll(rs.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), tt.wantBodyPart)
			}
		})
	}
}

func TestLogin_ErrorCases(t *testing.T) {
	internalErr := errors.New("login failed")

	tests := []struct {
		name            string
		body            string
		wantServiceCall bool
		serviceErr      error
		wantStatus      int
	}{
		{
			name:       "BadJSON",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:            "InvalidAuthInput",
			body:            `{"login":"","password":"secret"}`,
			wantServiceCall: true,
			serviceErr:      service.ErrInvalidAuthInput,
			wantStatus:      http.StatusBadRequest,
		},
		{
			name:            "InternalServerError",
			body:            `{"login":"admin","password":"secret"}`,
			wantServiceCall: true,
			serviceErr:      internalErr,
			wantStatus:      http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := &stubAuthService{}

			if tt.wantServiceCall {
				authService.loginFunc = func(ctx context.Context, login, password string) (model.AuthResult, error) {
					return model.AuthResult{}, tt.serviceErr
				}
			}

			rq := httptest.NewRequest(
				http.MethodPost,
				"/api/user/login",
				strings.NewReader(tt.body),
			)

			rs := serveHandler(t, Login(authService), rq)

			require.Equal(t, tt.wantStatus, rs.StatusCode)
		})
	}
}

func TestUploadOrder_ErrorCases(t *testing.T) {
	internalErr := errors.New("upload order failed")

	tests := []struct {
		name         string
		serviceErr   error
		resultStatus model.UploadOrderStatus
		wantStatus   int
	}{
		{
			name:       "InvalidInput",
			serviceErr: service.ErrInvalidOrderInput,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "InvalidOrderNumber",
			serviceErr: service.ErrInvalidOrderNumber,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "InternalServerError",
			serviceErr: internalErr,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:         "UnexpectedUploadStatus",
			resultStatus: model.UploadOrderStatus("UNKNOWN"),
			wantStatus:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderService := &stubOrderService{
				uploadOrderFunc: func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
					require.Equal(t, handlerCurrentUserID, userID)
					require.Equal(t, validLuhnOrderNumber, orderNumber)

					if tt.serviceErr != nil {
						return model.UploadOrderResult{}, tt.serviceErr
					}

					return model.UploadOrderResult{
						Status: tt.resultStatus,
					}, nil
				},
			}

			rq := newAuthorizedRequest(http.MethodPost, "/api/user/orders", validLuhnOrderNumber)

			rs := serveWithAuth(t, UploadOrder(orderService), rq)

			require.Equal(t, tt.wantStatus, rs.StatusCode)
		})
	}
}

func TestGetOrders_ServiceError(t *testing.T) {
	expectedErr := errors.New("list orders failed")

	orderService := &stubOrderService{
		listOrdersFunc: func(ctx context.Context, userID int64) ([]model.Order, error) {
			require.Equal(t, handlerCurrentUserID, userID)

			return nil, expectedErr
		},
	}

	rq := newAuthorizedRequest(http.MethodGet, "/api/user/orders", "")

	rs := serveWithAuth(t, GetOrders(orderService), rq)

	require.Equal(t, http.StatusInternalServerError, rs.StatusCode)
}

func TestGetBalance_ServiceError(t *testing.T) {
	expectedErr := errors.New("get balance failed")

	balanceService := &stubBalanceService{
		getBalanceFunc: func(ctx context.Context, userID int64) (model.Balance, error) {
			require.Equal(t, handlerCurrentUserID, userID)

			return model.Balance{}, expectedErr
		},
	}

	rq := newAuthorizedRequest(http.MethodGet, "/api/user/balance", "")

	rs := serveWithAuth(t, GetBalance(balanceService), rq)

	require.Equal(t, http.StatusInternalServerError, rs.StatusCode)
}

func TestWithdraw_ErrorCases(t *testing.T) {
	internalErr := errors.New("withdraw failed")

	tests := []struct {
		name            string
		body            string
		wantServiceCall bool
		serviceErr      error
		wantSum         int64
		wantStatus      int
	}{
		{
			name:       "BadJSON",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:            "InvalidWithdrawOrderNumber",
			body:            `{"order":"` + validLuhnOrderNumber + `","sum":5.11}`,
			wantServiceCall: true,
			serviceErr:      service.ErrInvalidWithdrawOrderNumber,
			wantSum:         511,
			wantStatus:      http.StatusUnprocessableEntity,
		},
		{
			name:            "InvalidWithdrawSum",
			body:            `{"order":"` + validLuhnOrderNumber + `","sum":-5.11}`,
			wantServiceCall: true,
			serviceErr:      service.ErrInvalidWithdrawSum,
			wantSum:         -511,
			wantStatus:      http.StatusBadRequest,
		},
		{
			name:            "InternalServerError",
			body:            `{"order":"` + validLuhnOrderNumber + `","sum":5.11}`,
			wantServiceCall: true,
			serviceErr:      internalErr,
			wantSum:         511,
			wantStatus:      http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balanceService := &stubBalanceService{}

			if tt.wantServiceCall {
				balanceService.withdrawFunc = func(ctx context.Context, userID int64, orderNumber string, sum int64) error {
					require.Equal(t, handlerCurrentUserID, userID)
					require.Equal(t, validLuhnOrderNumber, orderNumber)
					require.Equal(t, tt.wantSum, sum)

					return tt.serviceErr
				}
			}

			rq := newAuthorizedRequest(
				http.MethodPost,
				"/api/user/balance/withdraw",
				tt.body,
			)
			rq.Header.Set("Content-Type", "application/json")

			rs := serveWithAuth(t, Withdraw(balanceService), rq)

			require.Equal(t, tt.wantStatus, rs.StatusCode)
		})
	}
}

func TestGetWithdrawals_ServiceError(t *testing.T) {
	expectedErr := errors.New("list withdrawals failed")

	balanceService := &stubBalanceService{
		listWithdrawalsFunc: func(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
			require.Equal(t, handlerCurrentUserID, userID)

			return nil, expectedErr
		},
	}

	rq := newAuthorizedRequest(http.MethodGet, "/api/user/withdrawals", "")

	rs := serveWithAuth(t, GetWithdrawals(balanceService), rq)

	require.Equal(t, http.StatusInternalServerError, rs.StatusCode)
}
