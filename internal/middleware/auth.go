package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/xhrobj/gophermart/internal/auth"
)

const userIDContextKey = "user_id"

// проверяем bearer токен в запросе и кладем userId из него в контекст
func Auth(tokenManager auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. читаем хэдер запроса
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// 2. рассчитываем найти Bearer +jwt
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

			// 3. вытаскиваем из него userId
			userID, err := tokenManager.Parse(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// 4. кладём userId в конеткст запроса
			ctx := context.WithValue(r.Context(), userIDContextKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// достаём и возвращаем userId из контекста запроса, если он там есть
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}
