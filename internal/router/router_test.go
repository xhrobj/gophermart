package router

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
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

	r := New(authService, zap.NewNop())

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

	r := New(authService, zap.NewNop())

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

	r := New(authService, zap.NewNop())

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

	r := New(authService, zap.NewNop())

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
