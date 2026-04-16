package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open открывает подключение к PostgreSQL и проверяет доступность базы данных.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("database connection string is empty")
	}

	const driverName = "pgx"

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping database: %w; close database: %w", err, closeErr)
		}

		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
