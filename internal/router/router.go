package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xhrobj/gophermart/internal/auth"
	"github.com/xhrobj/gophermart/internal/handler"
	"github.com/xhrobj/gophermart/internal/middleware"
	"github.com/xhrobj/gophermart/internal/service"
	"go.uber.org/zap"
)

// New собирает HTTP-роутер приложения.
func New(
	authService service.AuthService,
	orderService service.OrderService,
	tokenManager auth.TokenManager,
	lg *zap.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.WithLogging(lg))

	r.Route("/api/user", func(r chi.Router) {
		// Публичные маршруты
		r.Post("/register", handler.Register(authService))
		r.Post("/login", handler.Login(authService))

		// Защищенные маршруты
		r.Group(func(r chi.Router) {
			r.Use(middleware.WithAuth(tokenManager))
			r.Post("/orders", handler.UploadOrder(orderService))
			r.Get("/orders", handler.GetOrders(orderService))
		})
	})

	return r
}
