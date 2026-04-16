package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/xhrobj/gophermart/internal/config"
	"github.com/xhrobj/gophermart/internal/database"
	"github.com/xhrobj/gophermart/internal/handler"
	"github.com/xhrobj/gophermart/internal/logger"
	"github.com/xhrobj/gophermart/internal/middleware"
	"github.com/xhrobj/gophermart/internal/migration"
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

	printBanner()

	return runRouter(cfg.RunAddress, db, lg)
}

func runRouter(runAddress string, db *sql.DB, lg *zap.Logger) error {
	router := http.NewServeMux()
	router.HandleFunc("/ping", handler.DBPing(db, lg))

	server := &http.Server{
		Addr:              runAddress,
		Handler:           middleware.WithLogging(lg)(router),
		ReadHeaderTimeout: time.Second * 5,
	}

	lg.Info("(^.^)~ Gophermart is starting HTTP server",
		zap.String("address", runAddress),
	)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("run HTTP server: %w", err)
	}

	return nil
}

func printBanner() {
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
