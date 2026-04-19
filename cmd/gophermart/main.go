package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/xhrobj/gophermart/internal/auth"
	"github.com/xhrobj/gophermart/internal/config"
	"github.com/xhrobj/gophermart/internal/database"
	"github.com/xhrobj/gophermart/internal/logger"
	"github.com/xhrobj/gophermart/internal/migration"
	"github.com/xhrobj/gophermart/internal/repository"
	"github.com/xhrobj/gophermart/internal/router"
	"github.com/xhrobj/gophermart/internal/service"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	lg, err := logger.New()
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer func() {
		_ = lg.Sync()
	}()

	cfg := config.GetConfig()

	db, err := database.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			lg.Error("close PostgreSQL connection", zap.Error(err))
		}
	}()

	lg.Info("PostgreSQL connection is ready")

	if err := migration.Run(cfg.DatabaseDSN); err != nil {
		return fmt.Errorf("run database migrations: %w", err)
	}

	lg.Info("database migrations are up to date")

	if cfg.AccrualSystemAddress == "" {
		lg.Info("accrual system address is empty")
	} else {
		lg.Info("accrual system address is configured",
			zap.String("address", cfg.AccrualSystemAddress),
		)
	}

	echoBanner()

	userRepo := repository.NewPostgresUserRepository(db)
	passwordManager := auth.NewSHA256PasswordManager()
	tokenManager := auth.NewJWTTokenManager(cfg.JWTSecret, 24*time.Hour)
	authService := service.NewAuthService(userRepo, passwordManager, tokenManager)

	orderRepo := repository.NewPostgresOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)

	balanceRepo := repository.NewPostgresBalanceRepository(db)
	balanceService := service.NewBalanceService(balanceRepo)

	appRouter := router.New(authService, orderService, balanceService, tokenManager, lg)

	lg.Info("(^.^)~ Gophermart is starting HTTP server",
		zap.String("address", cfg.RunAddress),
	)

	server := &http.Server{
		Addr:              cfg.RunAddress,
		Handler:           appRouter,
		ReadHeaderTimeout: time.Second * 5,
	}

	return server.ListenAndServe()
}

func echoBanner() {
	const banner = `
  ________              .__                                        __   
 /  _____/  ____ ______ |  |__   ___________  _____ _____ ________/  |_ 
/   \  ___ /  _ \\____ \|  |  \_/ __ \_  __ \/     \\__  \\_  __ \   __\
\    \_\  (  <_> )  |_> >   Y  \  ___/|  | \/  Y Y  \/ __ \|  | \/|  |
 \______  /\____/|   __/|___|  /\___  >__|  |__|_|  (____  /__|   |__|
        \/       |__|        \/     \/            \/     \/
	`
	fmt.Println(banner)
}
