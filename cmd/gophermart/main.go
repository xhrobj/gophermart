package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xhrobj/gophermart/internal/accrual"
	"github.com/xhrobj/gophermart/internal/auth"
	"github.com/xhrobj/gophermart/internal/config"
	"github.com/xhrobj/gophermart/internal/database"
	"github.com/xhrobj/gophermart/internal/logger"
	"github.com/xhrobj/gophermart/internal/migration"
	"github.com/xhrobj/gophermart/internal/repository"
	"github.com/xhrobj/gophermart/internal/router"
	"github.com/xhrobj/gophermart/internal/service"
	"github.com/xhrobj/gophermart/internal/worker"
	"go.uber.org/zap"
)

const shutdownTimeout = time.Second * 10

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithCancel(signalCtx)
	defer cancel()

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

	orderRepo := repository.NewPostgresOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)

	var accrualWorker *worker.AccrualWorker
	var workersWG sync.WaitGroup

	defer func() {
		cancel()
		workersWG.Wait()
	}()

	if cfg.AccrualSystemAddress == "" {
		lg.Info("accrual system address is empty")
	} else {
		lg.Info("accrual system address is configured",
			zap.String("address", cfg.AccrualSystemAddress),
		)

		accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
		accrualService := service.NewAccrualService(orderRepo, accrualClient, lg)
		accrualWorker = worker.NewAccrualWorker(accrualService, lg)
	}

	if accrualWorker != nil {
		workersWG.Add(1)

		go func() {
			defer workersWG.Done()
			accrualWorker.Run(ctx)
		}()
	}

	echoBanner()

	userRepo := repository.NewPostgresUserRepository(db)
	passwordManager := auth.NewBcryptPasswordManager()
	tokenManager := auth.NewJWTTokenManager(cfg.JWTSecret, 24*time.Hour)
	authService := service.NewAuthService(userRepo, passwordManager, tokenManager)

	balanceRepo := repository.NewPostgresBalanceRepository(db)
	balanceService := service.NewBalanceService(balanceRepo)

	appRouter := router.New(authService, orderService, balanceService, tokenManager, lg)

	lg.Info("(^.^)~ Gophermart is starting HTTP server ...",
		zap.String("address", cfg.RunAddress),
	)

	server := &http.Server{
		Addr:              cfg.RunAddress,
		Handler:           appRouter,
		ReadHeaderTimeout: time.Second * 5,
	}

	serverErrCh := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}

		serverErrCh <- nil
	}()

	select {
	case <-ctx.Done():
		stop()

		lg.Info("(^.^)~ Gophermart is shutting down ...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		if err := <-serverErrCh; err != nil {
			return fmt.Errorf("run HTTP server: %w", err)
		}

		lg.Info("HTTP server stopped")
		return nil

	case err := <-serverErrCh:
		if err != nil {
			return fmt.Errorf("run HTTP server: %w", err)
		}

		return nil
	}
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
