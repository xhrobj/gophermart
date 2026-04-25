//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/model"
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

func updateOrderForListTest(
	t *testing.T,
	db *sql.DB,
	number string,
	status model.OrderStatus,
	accrual int64,
	uploadedAt time.Time,
) {
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

func updateOrderForPendingTest(
	t *testing.T,
	db *sql.DB,
	number string,
	status model.OrderStatus,
	accrual int64,
	uploadedAt time.Time,
	nextPollAt time.Time,
) {
	t.Helper()

	_, err := db.Exec(
		`UPDATE orders
		 SET status = $1, accrual = $2, uploaded_at = $3, next_poll_at = $4
		 WHERE number = $5`,
		status,
		accrual,
		uploadedAt,
		nextPollAt,
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
	require.Equal(t, model.OrderStatusNew, got.Status)
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
	require.Equal(t, created.UploadedAt.UTC(), got.UploadedAt.UTC())
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

	updateOrderForListTest(t, db, "109", model.OrderStatusNew, 0, olderTime)
	updateOrderForListTest(t, db, "117", model.OrderStatusProcessed, 50050, newerTime)
	updateOrderForListTest(t, db, "125", model.OrderStatusInvalid, 0, newerTime)

	got, err := orderRepo.ListByUserID(ctx, firstUserID)
	require.NoError(t, err)

	require.Len(t, got, 2)

	require.Equal(t, "117", got[0].Number)
	require.Equal(t, firstUserID, got[0].UserID)
	require.Equal(t, model.OrderStatusProcessed, got[0].Status)
	require.Equal(t, int64(50050), got[0].Accrual)
	require.Equal(t, newerTime, got[0].UploadedAt.UTC())

	require.Equal(t, "109", got[1].Number)
	require.Equal(t, firstUserID, got[1].UserID)
	require.Equal(t, model.OrderStatusNew, got[1].Status)
	require.Equal(t, int64(0), got[1].Accrual)
	require.Equal(t, olderTime, got[1].UploadedAt.UTC())
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

func TestPostgresOrderRepository_ListPending_OK(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, userRepo, "admin")

	_, err := orderRepo.Create(ctx, userID, "109")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, userID, "117")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, userID, "125")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, userID, "133")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, userID, "141")
	require.NoError(t, err)

	now := time.Now().UTC()
	olderUploadedAt := time.Date(2026, 4, 18, 20, 0, 0, 0, time.UTC)
	newerUploadedAt := time.Date(2026, 4, 19, 20, 0, 0, 0, time.UTC)

	updateOrderForPendingTest(t, db, "109", model.OrderStatusNew, 0, olderUploadedAt, now.Add(-2*time.Hour))
	updateOrderForPendingTest(t, db, "117", model.OrderStatusProcessing, 0, newerUploadedAt, now.Add(-time.Hour))
	updateOrderForPendingTest(t, db, "125", model.OrderStatusProcessed, 50050, newerUploadedAt, now.Add(-3*time.Hour))
	updateOrderForPendingTest(t, db, "133", model.OrderStatusNew, 0, newerUploadedAt, now.Add(time.Hour))
	updateOrderForPendingTest(t, db, "141", model.OrderStatusInvalid, 0, newerUploadedAt, now.Add(-4*time.Hour))

	got, err := orderRepo.ListPending(ctx, 2)

	require.NoError(t, err)
	require.Len(t, got, 2)

	require.Equal(t, "109", got[0].Number)
	require.Equal(t, userID, got[0].UserID)
	require.Equal(t, model.OrderStatusNew, got[0].Status)
	require.Equal(t, int64(0), got[0].Accrual)
	require.Equal(t, olderUploadedAt, got[0].UploadedAt.UTC())

	require.Equal(t, "117", got[1].Number)
	require.Equal(t, userID, got[1].UserID)
	require.Equal(t, model.OrderStatusProcessing, got[1].Status)
	require.Equal(t, int64(0), got[1].Accrual)
	require.Equal(t, newerUploadedAt, got[1].UploadedAt.UTC())
}

func TestPostgresOrderRepository_ListPending_Empty(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, userRepo, "admin")

	_, err := orderRepo.Create(ctx, userID, "109")
	require.NoError(t, err)

	now := time.Now().UTC()
	uploadedAt := time.Date(2026, 4, 19, 20, 0, 0, 0, time.UTC)

	updateOrderForPendingTest(t, db, "109", model.OrderStatusProcessed, 50050, uploadedAt, now.Add(-time.Hour))

	got, err := orderRepo.ListPending(ctx, 10)

	require.NoError(t, err)
	require.Empty(t, got)
}

func TestPostgresOrderRepository_SetAccrualResult_OK(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, userRepo, "admin")

	_, err := orderRepo.Create(ctx, userID, "12345678903")
	require.NoError(t, err)

	nextPollAt := time.Date(2026, 4, 19, 20, 0, 0, 0, time.UTC)

	err = orderRepo.SetAccrualResult(ctx, "12345678903", OrderAccrualUpdate{
		Status:     model.OrderStatusProcessed,
		Accrual:    50050,
		NextPollAt: nextPollAt,
	})
	require.NoError(t, err)

	got, err := orderRepo.FindByNumber(ctx, "12345678903")
	require.NoError(t, err)

	require.Equal(t, "12345678903", got.Number)
	require.Equal(t, userID, got.UserID)
	require.Equal(t, model.OrderStatusProcessed, got.Status)
	require.Equal(t, int64(50050), got.Accrual)

	var gotNextPollAt time.Time
	err = db.QueryRow(
		`SELECT next_poll_at FROM orders WHERE number = $1`,
		"12345678903",
	).Scan(&gotNextPollAt)
	require.NoError(t, err)
	require.Equal(t, nextPollAt, gotNextPollAt.UTC())
}

func TestPostgresOrderRepository_SetAccrualResult_OrderNotFound(t *testing.T) {
	db := openTestDB(t)
	truncateOrders(t, db)

	orderRepo := NewPostgresOrderRepository(db)

	err := orderRepo.SetAccrualResult(context.Background(), "missing", OrderAccrualUpdate{
		Status:     model.OrderStatusProcessing,
		Accrual:    0,
		NextPollAt: time.Date(2026, 4, 19, 20, 0, 0, 0, time.UTC),
	})

	require.ErrorIs(t, err, ErrOrderNotFound)
}
