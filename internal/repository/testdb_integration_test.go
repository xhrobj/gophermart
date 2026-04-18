//go:build integration

package repository

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// 1. подготовить тестовую бд:
	// docker exec -it gophermart-postgres psql -U gophermart -d postgres -c "CREATE DATABASE gophermart_test;"
	// docker exec -i gophermart-postgres psql -U gophermart -d gophermart_test < migrations/001_init.up.sql
	//
	// 2. положить dsn в окружение:
	// export TEST_DATABASE_DSN="postgres://gophermart:secret@localhost:5432/gophermart_test?sslmode=disable"

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	err = db.Ping()
	require.NoError(t, err)

	return db
}
