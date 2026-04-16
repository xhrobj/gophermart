package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/xhrobj/gophermart/internal/config"
)

func main() {
	printBanner()

	cfg := config.GetConfig()

	log.Printf("(^.^)~ Gophermart will run on %s", cfg.RunAddress)

	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Fatal(err)
		}

		log.Println("database connected")
	} else {
		log.Println("database connection string is empty")
	}

	if cfg.AccrualSystemAddress == "" {
		log.Println("accrual system address is empty")
	} else {
		log.Printf("accrual system address is configured: %s", cfg.AccrualSystemAddress)
	}
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
