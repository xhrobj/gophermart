//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

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

func updateOrderForListTest(t *testing.T, db *sql.DB, number string, status string, accrual int64, uploadedAt time.Time) {
	t.Helper()

	_, err := db.Exec(
		`UPDATE orders SET status = $1, accrual = $2, uploaded_at = $3 WHERE number = $4`,
		status,
		accrual,
		uploadedAt,
		number,
	)
	require.NoError(t, err)
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

func TestPostgresOrderRepository_ListByUserID_OK(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	ctx := context.Background()

	firstUserID := createTestUser(t, userRepo, "first-user")
	secondUserID := createTestUser(t, userRepo, "second-user")

	_, err := orderRepo.Create(ctx, firstUserID, "109")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, firstUserID, "117")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, secondUserID, "125")
	require.NoError(t, err)

	olderTime := time.Date(2026, 4, 18, 20, 0, 0, 0, time.UTC)
	newerTime := time.Date(2026, 4, 19, 20, 0, 0, 0, time.UTC)

	updateOrderForListTest(t, db, "109", "NEW", 0, olderTime)
	updateOrderForListTest(t, db, "117", "PROCESSED", 50050, newerTime)
	updateOrderForListTest(t, db, "125", "INVALID", 0, newerTime)

	got, err := orderRepo.ListByUserID(ctx, firstUserID)
	require.NoError(t, err)

	require.Len(t, got, 2)

	require.Equal(t, "117", got[0].Number)
	require.Equal(t, firstUserID, got[0].UserID)
	require.Equal(t, "PROCESSED", string(got[0].Status))
	require.Equal(t, int64(50050), got[0].Accrual)
	require.Equal(t, newerTime, got[0].UploadedAt)

	require.Equal(t, "109", got[1].Number)
	require.Equal(t, firstUserID, got[1].UserID)
	require.Equal(t, "NEW", string(got[1].Status))
	require.Equal(t, int64(0), got[1].Accrual)
	require.Equal(t, olderTime, got[1].UploadedAt)
}

func TestPostgresOrderRepository_ListByUserID_Empty(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, userRepo, "admin")

	got, err := orderRepo.ListByUserID(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, got)
}
