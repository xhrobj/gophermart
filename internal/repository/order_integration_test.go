//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func truncateOrders(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`TRUNCATE TABLE orders, users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func createTestUser(t *testing.T, repo *PostgresUserRepository, login string) int64 {
	t.Helper()

	user, err := repo.Create(context.Background(), login, "hashed-secret")
	require.NoError(t, err)

	return user.ID
}

func TestPostgresOrderRepository_Create_OK(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	userID := createTestUser(t, userRepo, "admin")

	got, err := orderRepo.Create(context.Background(), userID, "12345678903")
	require.NoError(t, err)

	require.NotZero(t, got.ID)
	require.Equal(t, "12345678903", got.Number)
	require.Equal(t, userID, got.UserID)
	require.Equal(t, "NEW", string(got.Status))
	require.Equal(t, int64(0), got.Accrual)
	require.False(t, got.UploadedAt.IsZero())
}

func TestPostgresOrderRepository_Create_OrderAlreadyExists(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	ctx := context.Background()

	firstUserID := createTestUser(t, userRepo, "first")
	secondUserID := createTestUser(t, userRepo, "second")

	_, err := orderRepo.Create(ctx, firstUserID, "12345678903")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, secondUserID, "12345678903")
	require.ErrorIs(t, err, ErrOrderAlreadyExists)
}

func TestPostgresOrderRepository_FindByNumber_OK(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	ctx := context.Background()
	userID := createTestUser(t, userRepo, "admin")

	created, err := orderRepo.Create(ctx, userID, "12345678903")
	require.NoError(t, err)

	got, err := orderRepo.FindByNumber(ctx, "12345678903")
	require.NoError(t, err)

	require.Equal(t, created.ID, got.ID)
	require.Equal(t, created.Number, got.Number)
	require.Equal(t, created.UserID, got.UserID)
	require.Equal(t, created.Status, got.Status)
	require.Equal(t, created.Accrual, got.Accrual)
	require.Equal(t, created.UploadedAt, got.UploadedAt)
}

func TestPostgresOrderRepository_FindByNumber_OrderNotFound(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	orderRepo := NewPostgresOrderRepository(db)

	_, err := orderRepo.FindByNumber(context.Background(), "missing")
	require.ErrorIs(t, err, ErrOrderNotFound)
}
