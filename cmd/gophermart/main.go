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
	"github.com/xhrobj/gophermart/internal/migration"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	cfg := config.GetConfig()

	log.Printf("(^.^)~ Gophermart will run on %s", cfg.RunAddress)

	db, err := database.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close PostgreSQL connection: %v", err)
		}
	}()

	log.Println("PostgreSQL connection is ready")

	if err := migration.Run(cfg.DatabaseDSN); err != nil {
		return fmt.Errorf("run database migrations: %w", err)
	}

	log.Println("database migrations are up to date")

	if cfg.AccrualSystemAddress == "" {
		log.Println("accrual system address is empty")
	} else {
		log.Printf("accrual system address is configured: %s", cfg.AccrualSystemAddress)
	}

	printBanner()

	return runRouter(cfg.RunAddress, db)
}

func runRouter(runAddress string, db *sql.DB) error {
	router := http.NewServeMux()
	router.HandleFunc("/ping", handler.DBPing(db))

	server := &http.Server{
		Addr:              runAddress,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 5,
	}

	log.Printf("(^.^)~ Gophermart is starting HTTP server on %s", runAddress)

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
