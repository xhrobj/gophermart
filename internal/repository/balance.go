package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/xhrobj/gophermart/internal/model"
)

// BalanceRepository описывает операции хранения баланса и списаний.
type BalanceRepository interface {
	// GetBalance возвращает текущий баланс и сумму всех списаний пользователя.
	GetBalance(ctx context.Context, userID int64) (model.Balance, error)

	// Withdraw создает операцию списания баллов.
	//
	// Сумма списания передается в копейках.
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error

	// ListWithdrawals возвращает список списаний пользователя.
	ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}

// PostgresBalanceRepository реализует работу с балансом в PostgreSQL.
type PostgresBalanceRepository struct {
	db *sql.DB
}

// NewPostgresBalanceRepository создаёт репозиторий баланса на PostgreSQL.
func NewPostgresBalanceRepository(db *sql.DB) *PostgresBalanceRepository {
	return &PostgresBalanceRepository{
		db: db,
	}
}

func (r *PostgresBalanceRepository) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	const query = `
		WITH accruals AS (
			SELECT COALESCE(SUM(accrual), 0) AS total
			FROM orders
			WHERE user_id = $1
			AND status = 'PROCESSED'
		),
		withdrawals_total AS (
			SELECT COALESCE(SUM(amount), 0) AS total
			FROM withdrawals
			WHERE user_id = $1
		)
		SELECT
			accruals.total - withdrawals_total.total AS current,
			withdrawals_total.total AS withdrawn
		FROM accruals, withdrawals_total
		`

	var balance model.Balance

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&balance.Current,
		&balance.Withdrawn,
	)
	if err != nil {
		return model.Balance{}, fmt.Errorf("get balance: %w", err)
	}

	return balance, nil
}

func (r *PostgresBalanceRepository) Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error {
	panic("¯＼_(ツ)_/¯")
}

func (r *PostgresBalanceRepository) ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	panic("¯＼_(ツ)_/¯")
}
