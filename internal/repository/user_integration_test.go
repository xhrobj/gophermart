//go:build integration

package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
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

func truncateUsers(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func TestPostgresUserRepository_Create_OK(t *testing.T) {
	db := openTestDB(t)
	truncateUsers(t, db)

	repo := NewPostgresUserRepository(db)

	got, err := repo.Create(context.Background(), "admin", "hashed-secret")
	require.NoError(t, err)

	require.NotZero(t, got.ID)
	require.Equal(t, "admin", got.Login)
	require.Equal(t, "hashed-secret", got.PasswordHash)
	require.False(t, got.CreatedAt.IsZero())
}

func TestPostgresUserRepository_Create_UserAlreadyExists(t *testing.T) {
	db := openTestDB(t)
	truncateUsers(t, db)

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, "admin", "hashed-secret")
	require.NoError(t, err)

	_, err = repo.Create(ctx, "admin", "another-hash")
	require.ErrorIs(t, err, ErrUserAlreadyExists)
}

func TestPostgresUserRepository_FindByLogin_OK(t *testing.T) {
	db := openTestDB(t)
	truncateUsers(t, db)

	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, "admin", "hashed-secret")
	require.NoError(t, err)

	got, err := repo.FindByLogin(ctx, "admin")
	require.NoError(t, err)

	require.Equal(t, created.ID, got.ID)
	require.Equal(t, "admin", got.Login)
	require.Equal(t, "hashed-secret", got.PasswordHash)
	require.Equal(t, created.CreatedAt, got.CreatedAt)
}

func TestPostgresUserRepository_FindByLogin_UserNotFound(t *testing.T) {
	db := openTestDB(t)
	truncateUsers(t, db)

	repo := NewPostgresUserRepository(db)

	_, err := repo.FindByLogin(context.Background(), "unknown")
	require.ErrorIs(t, err, ErrUserNotFound)
}
