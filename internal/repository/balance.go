package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/xhrobj/gophermart/internal/model"
)

// ErrInsufficientFunds означает, что на бонусном счете недостаточно средств для списания.
var ErrInsufficientFunds = errors.New("insufficient funds")

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

func (r *PostgresBalanceRepository) Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin withdraw transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const lockOrdersQuery = `
		SELECT id
		FROM orders
		WHERE user_id = $1
		  AND status = 'PROCESSED'
		FOR UPDATE
	`

	rows, err := tx.QueryContext(ctx, lockOrdersQuery, userID)
	if err != nil {
		return fmt.Errorf("lock processed orders: %w", err)
	}

	for rows.Next() {
		var orderID int64

		if err = rows.Scan(&orderID); err != nil {
			_ = rows.Close()

			return fmt.Errorf("scan locked order row: %w", err)
		}
	}

	if err = rows.Err(); err != nil {
		_ = rows.Close()

		return fmt.Errorf("iterate locked order rows: %w", err)
	}

	if err = rows.Close(); err != nil {
		return fmt.Errorf("close locked order rows: %w", err)
	}

	const balanceQuery = `
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
		SELECT accruals.total - withdrawals_total.total AS current
		FROM accruals, withdrawals_total
	`

	var currentBalance int64

	err = tx.QueryRowContext(ctx, balanceQuery, userID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("get current balance in transaction: %w", err)
	}

	if currentBalance < sum {
		return ErrInsufficientFunds
	}

	const insertWithdrawalQuery = `
		INSERT INTO withdrawals (user_id, order_number, amount)
		VALUES ($1, $2, $3)
	`

	_, err = tx.ExecContext(ctx, insertWithdrawalQuery, userID, orderNumber, sum)
	if err != nil {
		return fmt.Errorf("insert withdrawal: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit withdraw transaction: %w", err)
	}

	return nil
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
