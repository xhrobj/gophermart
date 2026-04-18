package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/xhrobj/gophermart/internal/auth"
)

type userIDContextKey struct{}

// WithAuth проверяет bearer-токен в запросе, извлекает из него userID
// и сохраняет идентификатор пользователя в контексте запроса.
func WithAuth(tokenManager auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID, err := tokenManager.Parse(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey{}, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext возвращает userID из контекста запроса, если он был сохранен middleware авторизации.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(int64)
	return userID, ok
}
