package main

import (
	"context"
	"fmt"
	"log"

	"github.com/xhrobj/gophermart/internal/config"
	"github.com/xhrobj/gophermart/internal/database"
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

	if cfg.AccrualSystemAddress == "" {
		log.Println("accrual system address is empty")
	} else {
		log.Printf("accrual system address is configured: %s", cfg.AccrualSystemAddress)
	}

	printBanner()

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
