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

func newTestRouter(authService service.AuthService) http.Handler {
	return New(authService, &stubTokenManager{}, zap.NewNop())
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

	r := newTestRouter(authService)

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

	r := newTestRouter(authService)

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

	r := newTestRouter(authService)

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

	r := newTestRouter(authService)

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

func TestGetOrders_Unauthorized(t *testing.T) {
	authService := &stubAuthService{}
	tokenManager := &stubTokenManager{}

	r := New(authService, tokenManager, zap.NewNop())

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

	r := New(authService, tokenManager, zap.NewNop())

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

	r := New(authService, tokenManager, zap.NewNop())

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
