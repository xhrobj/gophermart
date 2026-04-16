package main

import (
	"fmt"
	"log"

	"github.com/xhrobj/gophermart/internal/config"
)

func main() {
	printBanner()

	cfg := config.GetConfig()

	log.Printf("(^.^)~ Gophermart will run on %s", cfg.RunAddress)

	if cfg.DatabaseDSN == "" {
		log.Println("PostgreSQL connection string is empty")
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
