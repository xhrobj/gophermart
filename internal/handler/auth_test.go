package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/service"
)

func TestRegister_OK(t *testing.T) {
	authService := &stubAuthService{
		registerFunc: func(ctx context.Context, login, password string) (model.AuthResult, error) {
			require.Equal(t, "admin", login)
			require.Equal(t, "secret", password)

			return model.AuthResult{
				UserID: handlerCurrentUserID,
				Token:  handlerIssuedToken,
			}, nil
		},
	}

	rq := httptest.NewRequest(
		http.MethodPost,
		"/api/user/register",
		bytes.NewBufferString(`{"login":"admin","password":"secret"}`),
	)
	rq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	Register(authService).ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusOK, rs.StatusCode)
	require.Equal(t, "Bearer "+handlerIssuedToken, rs.Header.Get("Authorization"))
}

func TestRegister_LoginAlreadyExists(t *testing.T) {
	authService := &stubAuthService{
		registerFunc: func(ctx context.Context, login, password string) (model.AuthResult, error) {
			require.Equal(t, "admin", login)
			require.Equal(t, "secret", password)

			return model.AuthResult{}, service.ErrLoginAlreadyExists
		},
	}

	rq := httptest.NewRequest(
		http.MethodPost,
		"/api/user/register",
		bytes.NewBufferString(`{"login":"admin","password":"secret"}`),
	)
	rq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	Register(authService).ServeHTTP(rec, rq)

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
				UserID: handlerCurrentUserID,
				Token:  handlerIssuedToken,
			}, nil
		},
	}

	rq := httptest.NewRequest(
		http.MethodPost,
		"/api/user/login",
		bytes.NewBufferString(`{"login":"admin","password":"secret"}`),
	)
	rq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	Login(authService).ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusOK, rs.StatusCode)
	require.Equal(t, "Bearer "+handlerIssuedToken, rs.Header.Get("Authorization"))
}

func TestLogin_InvalidCredentials(t *testing.T) {
	authService := &stubAuthService{
		loginFunc: func(ctx context.Context, login, password string) (model.AuthResult, error) {
			require.Equal(t, "admin", login)
			require.Equal(t, "wrong", password)

			return model.AuthResult{}, service.ErrInvalidCredentials
		},
	}

	rq := httptest.NewRequest(
		http.MethodPost,
		"/api/user/login",
		bytes.NewBufferString(`{"login":"admin","password":"wrong"}`),
	)
	rq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	Login(authService).ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
}
