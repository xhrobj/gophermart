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
	const query = `
		SELECT id, user_id, order_number, amount, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list withdrawals by user id: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	withdrawals := make([]model.Withdrawal, 0)
	for rows.Next() {
		var withdrawal model.Withdrawal
		err = rows.Scan(
			&withdrawal.ID,
			&withdrawal.UserID,
			&withdrawal.OrderNumber,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan withdrawal row: %w", err)
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate withdrawal rows: %w", err)
	}

	return withdrawals, nil
}
