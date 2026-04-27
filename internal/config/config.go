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

// Parse возвращает конфигурацию сервиса Gophermart.
//
// Значения параметров могут быть заданы через:
//   - флаги: -a -d -r
//   - переменные окружения: RUN_ADDRESS, DATABASE_URI, ACCRUAL_SYSTEM_ADDRESS
//
// Приоритет источников:
//   - для RunAddress, DatabaseDSN и AccrualSystemAddress: flag > env > default
//   - для JWTSecret: env > default
func Parse(args []string) (Config, error) {
	cfg := Config{
		RunAddress:           "localhost:8080",
		DatabaseDSN:          "",
		AccrualSystemAddress: "",
		JWTSecret:            "dev-secret",
	}

	if runAddress, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		cfg.RunAddress = runAddress
	}

	if databaseURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		cfg.DatabaseDSN = databaseURI
	}

	if accrualSystemAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		cfg.AccrualSystemAddress = accrualSystemAddress
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWTSecret = jwtSecret
	}

	flags := flag.NewFlagSet("gophermart", flag.ContinueOnError)

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "address and port to run server")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database connection string")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address")

	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
