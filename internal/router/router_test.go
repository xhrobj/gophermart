package router

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/auth"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/service"
	"go.uber.org/zap"
)

const (
	validOrderNumber       = "12345678903"
	invalidLuhnOrderNumber = "12345678904"
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

func (s *stubOrderService) UploadOrder(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
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

func newTestRouter(authService service.AuthService, orderService service.OrderService) http.Handler {
	return New(authService, orderService, &stubTokenManager{}, zap.NewNop())
}

func TestRegister_OK(t *testing.T) {
	authService := &stubAuthService{
		registerFunc: func(ctx context.Context, login, password string) (model.AuthResult, error) {
			require.Equal(t, "admin", login)
			require.Equal(t, "secret", password)

			return model.AuthResult{
				UserID: 42,
				Token:  "jwt-token",
			}, nil
		},
	}

	r := newTestRouter(authService, &stubOrderService{})

	rq := httptest.NewRequest(
		http.MethodPost,
		"/api/user/register",
		bytes.NewBufferString(`{"login":"admin","password":"secret"}`),
	)
	rq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusOK, rs.StatusCode)
	require.Equal(t, "Bearer jwt-token", rs.Header.Get("Authorization"))
}

func TestRegister_LoginAlreadyExists(t *testing.T) {
	authService := &stubAuthService{
		registerFunc: func(ctx context.Context, login, password string) (model.AuthResult, error) {
			return model.AuthResult{}, service.ErrLoginAlreadyExists
		},
	}

	r := newTestRouter(authService, &stubOrderService{})

	rq := httptest.NewRequest(
		http.MethodPost,
		"/api/user/register",
		bytes.NewBufferString(`{"login":"admin","password":"secret"}`),
	)
	rq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusConflict, rs.StatusCode)
}

func TestLogin_OK(t *testing.T) {
	authService := &stubAuthService{
		loginFunc: func(ctx context.Context, login, password string) (model.AuthResult, error) {
			require.Equal(t, "admin", login)
			require.Equal(t, "secret", password)

			return model.AuthResult{
				UserID: 42,
				Token:  "jwt-token",
			}, nil
		},
	}

	r := newTestRouter(authService, &stubOrderService{})

	rq := httptest.NewRequest(
		http.MethodPost,
		"/api/user/login",
		bytes.NewBufferString(`{"login":"admin","password":"secret"}`),
	)
	rq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusOK, rs.StatusCode)
	require.Equal(t, "Bearer jwt-token", rs.Header.Get("Authorization"))
}

func TestLogin_InvalidCredentials(t *testing.T) {
	authService := &stubAuthService{
		loginFunc: func(ctx context.Context, login, password string) (model.AuthResult, error) {
			return model.AuthResult{}, service.ErrInvalidCredentials
		},
	}

	r := newTestRouter(authService, &stubOrderService{})

	rq := httptest.NewRequest(
		http.MethodPost,
		"/api/user/login",
		bytes.NewBufferString(`{"login":"admin","password":"wrong"}`),
	)
	rq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
}

func TestUploadOrder_Unauthorized(t *testing.T) {
	authService := &stubAuthService{}
	orderService := &stubOrderService{}
	tokenManager := &stubTokenManager{}

	r := New(authService, orderService, tokenManager, zap.NewNop())

	rq := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(validOrderNumber))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
}

func TestUploadOrder_Accepted(t *testing.T) {
	authService := &stubAuthService{}
	orderService := &stubOrderService{
		uploadOrderFunc: func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
			require.Equal(t, int64(1), userID)
			require.Equal(t, "12345678903", orderNumber)

			return model.UploadOrderResult{
				Status: model.UploadOrderAccepted,
			}, nil
		},
	}

	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, "good-token", token)
			return 1, nil
		},
	}

	r := New(authService, orderService, tokenManager, zap.NewNop())

	rq := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(validOrderNumber))
	rq.Header.Set("Authorization", "Bearer good-token")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusAccepted, rs.StatusCode)
}

func TestUploadOrder_Duplicate(t *testing.T) {
	authService := &stubAuthService{}
	orderService := &stubOrderService{
		uploadOrderFunc: func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
			require.Equal(t, int64(1), userID)
			require.Equal(t, "12345678903", orderNumber)

			return model.UploadOrderResult{
				Status: model.UploadOrderDuplicate,
			}, nil
		},
	}

	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, "good-token", token)
			return 1, nil
		},
	}

	r := New(authService, orderService, tokenManager, zap.NewNop())

	rq := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(validOrderNumber))
	rq.Header.Set("Authorization", "Bearer good-token")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusOK, rs.StatusCode)
}

func TestUploadOrder_Conflict(t *testing.T) {
	authService := &stubAuthService{}
	orderService := &stubOrderService{
		uploadOrderFunc: func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
			require.Equal(t, int64(1), userID)
			require.Equal(t, validOrderNumber, orderNumber)

			return model.UploadOrderResult{
				Status: model.UploadOrderConflict,
			}, nil
		},
	}

	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, "good-token", token)
			return 1, nil
		},
	}

	r := New(authService, orderService, tokenManager, zap.NewNop())

	rq := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(validOrderNumber))
	rq.Header.Set("Authorization", "Bearer good-token")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusConflict, rs.StatusCode)
}

func TestUploadOrder_InvalidOrderInput(t *testing.T) {
	authService := &stubAuthService{}
	orderService := &stubOrderService{
		uploadOrderFunc: func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
			require.Equal(t, int64(1), userID)
			require.Equal(t, "   ", orderNumber)

			return model.UploadOrderResult{}, service.ErrInvalidOrderInput
		},
	}

	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, "good-token", token)
			return 1, nil
		},
	}

	r := New(authService, orderService, tokenManager, zap.NewNop())

	rq := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("   "))
	rq.Header.Set("Authorization", "Bearer good-token")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusBadRequest, rs.StatusCode)
}

func TestUploadOrder_InvalidOrderNumber(t *testing.T) {
	authService := &stubAuthService{}
	orderService := &stubOrderService{
		uploadOrderFunc: func(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
			require.Equal(t, int64(1), userID)
			require.Equal(t, invalidLuhnOrderNumber, orderNumber)

			return model.UploadOrderResult{}, service.ErrInvalidOrderNumber
		},
	}

	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, "good-token", token)
			return 1, nil
		},
	}

	r := New(authService, orderService, tokenManager, zap.NewNop())

	rq := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(invalidLuhnOrderNumber))
	rq.Header.Set("Authorization", "Bearer good-token")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnprocessableEntity, rs.StatusCode)
}

func TestGetOrders_Unauthorized(t *testing.T) {
	authService := &stubAuthService{}
	tokenManager := &stubTokenManager{}

	r := New(authService, &stubOrderService{}, tokenManager, zap.NewNop())

	rq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
}

func TestGetOrders_OK(t *testing.T) {
	authService := &stubAuthService{}

	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, "good-token", token)
			return 1, nil
		},
	}

	r := New(authService, &stubOrderService{}, tokenManager, zap.NewNop())

	rq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rq.Header.Set("Authorization", "Bearer good-token")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	body, err := io.ReadAll(rs.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rs.StatusCode)
	require.Equal(t, "text/plain; charset=utf-8", rs.Header.Get("Content-Type"))
	require.Equal(t, "userID=1", string(body))
}

func TestGetOrders_InvalidToken(t *testing.T) {
	authService := &stubAuthService{}

	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, "bad-token", token)
			return 0, auth.ErrInvalidToken
		},
	}

	r := New(authService, &stubOrderService{}, tokenManager, zap.NewNop())

	rq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rq.Header.Set("Authorization", "Bearer bad-token")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
}
