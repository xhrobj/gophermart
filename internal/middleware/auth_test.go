package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	authTestToken  = "good-token"
	authTestUserID = int64(42)
)

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

func TestWithAuth_OK(t *testing.T) {
	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, authTestToken, token)

			return authTestUserID, nil
		},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		require.True(t, ok)
		require.Equal(t, authTestUserID, userID)

		w.WriteHeader(http.StatusTeapot)
	})

	rq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rq.Header.Set("Authorization", "Bearer "+authTestToken)

	rec := httptest.NewRecorder()
	WithAuth(tokenManager)(next).ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusTeapot, rs.StatusCode)
}

func TestWithAuth_MissingAuthorizationHeader(t *testing.T) {
	tokenManager := &stubTokenManager{}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	})

	rq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)

	rec := httptest.NewRecorder()
	WithAuth(tokenManager)(next).ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
}

func TestWithAuth_InvalidAuthorizationPrefix(t *testing.T) {
	tokenManager := &stubTokenManager{}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	})

	rq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rq.Header.Set("Authorization", "Token "+authTestToken)

	rec := httptest.NewRecorder()
	WithAuth(tokenManager)(next).ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
}

func TestWithAuth_EmptyToken(t *testing.T) {
	tokenManager := &stubTokenManager{}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	})

	rq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rq.Header.Set("Authorization", "Bearer ")

	rec := httptest.NewRecorder()
	WithAuth(tokenManager)(next).ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
}

func TestWithAuth_ParseError(t *testing.T) {
	parseErr := errors.New("parse token")

	tokenManager := &stubTokenManager{
		parseFunc: func(token string) (int64, error) {
			require.Equal(t, authTestToken, token)

			return 0, parseErr
		},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	})

	rq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rq.Header.Set("Authorization", "Bearer "+authTestToken)

	rec := httptest.NewRecorder()
	WithAuth(tokenManager)(next).ServeHTTP(rec, rq)

	rs := rec.Result()
	t.Cleanup(func() {
		require.NoError(t, rs.Body.Close())
	})

	require.Equal(t, http.StatusUnauthorized, rs.StatusCode)
}

func TestUserIDFromContext_NotFound(t *testing.T) {
	rq := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)

	userID, ok := UserIDFromContext(rq.Context())

	require.False(t, ok)
	require.Zero(t, userID)
}
