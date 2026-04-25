package config

import (
	"flag"
	"os"
)

// Config содержит конфигурацию сервиса Gophermart.
type Config struct {
	// RunAddress содержит адрес и порт запуска HTTP-сервера.
	RunAddress string

	// DatabaseDSN содержит строку подключения к PostgreSQL.
	DatabaseDSN string

	// AccrualSystemAddress содержит адрес внешней системы расчёта начислений.
	AccrualSystemAddress string

	// JWTSecret содержит секрет для подписи JWT-токенов.
	JWTSecret string
}

// GetConfig возвращает конфигурацию сервиса Gophermart.
//
// Значения параметров могут быть заданы через:
//   - флаги: -a -d -r
//   - переменные окружения: RUN_ADDRESS, DATABASE_URI, ACCRUAL_SYSTEM_ADDRESS
//
// Приоритет источников:
//   - для RunAddress, DatabaseDSN и AccrualSystemAddress: flag > env > default
//   - для JWTSecret: env > default
func GetConfig() Config {
	cfg := Config{
		RunAddress:           "localhost:8080",
		DatabaseDSN:          "",
		AccrualSystemAddress: "",
		JWTSecret:            "dev-secret",
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "address and port to run server")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database connection string")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address")

	flag.Parse()

	definedFlags := make(map[string]bool)

	flag.Visit(func(f *flag.Flag) {
		definedFlags[f.Name] = true
	})

	if !definedFlags["a"] {
		if runAddress, ok := os.LookupEnv("RUN_ADDRESS"); ok {
			cfg.RunAddress = runAddress
		}
	}

	if !definedFlags["d"] {
		if databaseURI, ok := os.LookupEnv("DATABASE_URI"); ok {
			cfg.DatabaseDSN = databaseURI
		}
	}

	if !definedFlags["r"] {
		if accrualSystemAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
			cfg.AccrualSystemAddress = accrualSystemAddress
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret != "" {
		cfg.JWTSecret = jwtSecret
	}

	return cfg
}
