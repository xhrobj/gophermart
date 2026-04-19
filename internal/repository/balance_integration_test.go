//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func truncateBalanceData(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`TRUNCATE TABLE withdrawals, orders, users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func insertWithdrawalForBalanceTest(
	t *testing.T,
	db *sql.DB,
	userID int64,
	orderNumber string,
	amount int64,
	processedAt time.Time,
) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO withdrawals (user_id, order_number, amount, processed_at) VALUES ($1, $2, $3, $4)`,
		userID,
		orderNumber,
		amount,
		processedAt,
	)
	require.NoError(t, err)
}

func TestPostgresBalanceRepository_GetBalance_Empty(t *testing.T) {
	db := openTestDB(t)
	truncateBalanceData(t, db)

	userRepo := NewPostgresUserRepository(db)
	balanceRepo := NewPostgresBalanceRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, userRepo, "admin")

	got, err := balanceRepo.GetBalance(ctx, userID)
	require.NoError(t, err)

	require.Equal(t, int64(0), got.Current)
	require.Equal(t, int64(0), got.Withdrawn)
}

func TestPostgresBalanceRepository_GetBalance_OK(t *testing.T) {
	db := openTestDB(t)
	truncateBalanceData(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	balanceRepo := NewPostgresBalanceRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, userRepo, "admin")

	_, err := orderRepo.Create(ctx, userID, "109")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, userID, "117")
	require.NoError(t, err)

	updateOrderForListTest(
		t,
		db,
		"109",
		"PROCESSED",
		10050,
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
	)
	updateOrderForListTest(
		t,
		db,
		"117",
		"PROCESSED",
		20025,
		time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC),
	)

	insertWithdrawalForBalanceTest(
		t,
		db,
		userID,
		"12345678903",
		511,
		time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
	)
	insertWithdrawalForBalanceTest(
		t,
		db,
		userID,
		"79927398713",
		1000,
		time.Date(2026, 4, 19, 13, 0, 0, 0, time.UTC),
	)

	got, err := balanceRepo.GetBalance(ctx, userID)
	require.NoError(t, err)

	require.Equal(t, int64(28564), got.Current)
	require.Equal(t, int64(1511), got.Withdrawn)
}

func TestPostgresBalanceRepository_GetBalance_IgnoresNonProcessedOrdersAndOtherUsers(t *testing.T) {
	db := openTestDB(t)
	truncateBalanceData(t, db)

	userRepo := NewPostgresUserRepository(db)
	orderRepo := NewPostgresOrderRepository(db)
	balanceRepo := NewPostgresBalanceRepository(db)
	ctx := context.Background()

	firstUserID := createTestUser(t, userRepo, "first-user")
	secondUserID := createTestUser(t, userRepo, "second-user")

	_, err := orderRepo.Create(ctx, firstUserID, "109")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, firstUserID, "117")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, firstUserID, "125")
	require.NoError(t, err)

	_, err = orderRepo.Create(ctx, secondUserID, "133")
	require.NoError(t, err)

	updateOrderForListTest(
		t,
		db,
		"109",
		"PROCESSED",
		10050,
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
	)
	updateOrderForListTest(
		t,
		db,
		"117",
		"NEW",
		99999,
		time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC),
	)
	updateOrderForListTest(
		t,
		db,
		"125",
		"INVALID",
		77777,
		time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
	)
	updateOrderForListTest(
		t,
		db,
		"133",
		"PROCESSED",
		55555,
		time.Date(2026, 4, 19, 13, 0, 0, 0, time.UTC),
	)

	insertWithdrawalForBalanceTest(
		t,
		db,
		firstUserID,
		"12345678903",
		511,
		time.Date(2026, 4, 19, 14, 0, 0, 0, time.UTC),
	)
	insertWithdrawalForBalanceTest(
		t,
		db,
		secondUserID,
		"79927398713",
		1000,
		time.Date(2026, 4, 19, 15, 0, 0, 0, time.UTC),
	)

	got, err := balanceRepo.GetBalance(ctx, firstUserID)
	require.NoError(t, err)

	require.Equal(t, int64(9299), got.Current)
	require.Equal(t, int64(511), got.Withdrawn)
}
