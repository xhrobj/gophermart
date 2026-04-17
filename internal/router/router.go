package router

import (
	"net/http"

	"github.com/xhrobj/gophermart/internal/handler"
	"github.com/xhrobj/gophermart/internal/middleware"
	"github.com/xhrobj/gophermart/internal/service"
	"go.uber.org/zap"
)

// New собирает HTTP-роутер приложения.
func New(authService service.AuthService, lg *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/user/register", handler.Register(authService))
	mux.HandleFunc("/api/user/login", handler.Login(authService))

	return middleware.WithLogging(lg)(mux)
}
