package router

import (
	"net/http"

	"github.com/xhrobj/gophermart/internal/auth"
	"github.com/xhrobj/gophermart/internal/handler"
	"github.com/xhrobj/gophermart/internal/middleware"
	"github.com/xhrobj/gophermart/internal/service"
	"go.uber.org/zap"
)

// New собирает HTTP-роутер приложения.
func New(authService service.AuthService, tokenManager auth.TokenManager, lg *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	// публичные маршруты iam
	mux.HandleFunc("/api/user/register", handler.Register(authService))
	mux.HandleFunc("/api/user/login", handler.Login(authService))

	// защищенный маршрут для авторизованных пользователей
	mux.Handle("/api/user/orders", middleware.Auth(tokenManager)(handler.GetOrders()))

	return middleware.WithLogging(lg)(mux)
}
