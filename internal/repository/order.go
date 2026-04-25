package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/xhrobj/gophermart/internal/model"
)

var (
	// ErrOrderNotFound означает, что заказ не найден.
	ErrOrderNotFound = errors.New("order not found")

	// ErrOrderAlreadyExists означает, что заказ с таким номером уже существует.
	ErrOrderAlreadyExists = errors.New("order already exists")
)

// OrderAccrualUpdate описывает данные для обновления результата проверки заказа в accrual.
type OrderAccrualUpdate struct {
	Status     model.OrderStatus
	Accrual    int64
	NextPollAt time.Time
}

// OrderRepository описывает операции хранения заказов.
type OrderRepository interface {
	// Create создает новый заказ пользователя.
	Create(ctx context.Context, userID int64, orderNumber string) (model.Order, error)

	// FindByNumber возвращает заказ по номеру.
	FindByNumber(ctx context.Context, orderNumber string) (model.Order, error)

	// ListByUserID возвращает список заказов пользователя.
	ListByUserID(ctx context.Context, userID int64) ([]model.Order, error)

	// ListPending возвращает заказы, ожидающие проверки во внешнем сервисе начислений.
	ListPending(ctx context.Context, limit int) ([]model.Order, error)

	// SetAccrualResult сохраняет результат проверки заказа во внешнем сервисе начислений.
	SetAccrualResult(ctx context.Context, orderNumber string, update OrderAccrualUpdate) error
}

// PostgresOrderRepository реализует OrderRepository поверх PostgreSQL.
type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository создает репозиторий заказов на базе PostgreSQL.
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

// Create создает новый заказ пользователя.
//
// Если заказ с таким номером уже существует, метод возвращает ErrOrderAlreadyExists.
func (r *PostgresOrderRepository) Create(ctx context.Context, userID int64, orderNumber string) (model.Order, error) {
	const query = `
		INSERT INTO orders (number, user_id)
		VALUES ($1, $2)
		RETURNING id, number, user_id, status, accrual, uploaded_at
	`

	var order model.Order

	err := r.db.QueryRowContext(ctx, query, orderNumber, userID).Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return model.Order{}, ErrOrderAlreadyExists
		}

		return model.Order{}, fmt.Errorf("create order: %w", err)
	}

	return order, nil
}

// FindByNumber возвращает заказ по его номеру.
//
// Если заказ не найден, метод возвращает ErrOrderNotFound.
func (r *PostgresOrderRepository) FindByNumber(ctx context.Context, orderNumber string) (model.Order, error) {
	const query = `
		SELECT id, number, user_id, status, accrual, uploaded_at
		FROM orders
		WHERE number = $1
	`

	var order model.Order

	err := r.db.QueryRowContext(ctx, query, orderNumber).Scan(
		&order.ID,
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Order{}, ErrOrderNotFound
		}

		return model.Order{}, fmt.Errorf("find order by number: %w", err)
	}

	return order, nil
}

// ListByUserID возвращает список заказов пользователя
//
// в обратном хронологическом порядке.
func (r *PostgresOrderRepository) ListByUserID(ctx context.Context, userID int64) ([]model.Order, error) {
	const query = `
		SELECT
			id,
			number,
			user_id,
			status,
			accrual,
			uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list orders by user id: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	orders := make([]model.Order, 0)

	for rows.Next() {
		var order model.Order

		err = rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan order row: %w", err)
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order rows: %w", err)
	}

	return orders, nil
}

// выбирает NEW и PROCESSING только с next_poll_at <= now
func (r *PostgresOrderRepository) ListPending(ctx context.Context, limit int) ([]model.Order, error) {
	const query = `
		SELECT
			id,
			number,
			user_id,
			status,
			accrual,
			uploaded_at
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING')
		  AND next_poll_at <= now()
		ORDER BY next_poll_at ASC, uploaded_at ASC, id ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending orders: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	orders := make([]model.Order, 0, limit)

	for rows.Next() {
		var order model.Order

		err = rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan pending order row: %w", err)
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending order rows: %w", err)
	}

	return orders, nil
}

// SetAccrualResult сохраняет результат проверки заказа во внешнем сервисе начислений.
//
// Метод обновляет статус заказа, сумму начислений и время следующего опроса.
// Если заказ с указанным номером не найден, метод возвращает ErrOrderNotFound.
func (r *PostgresOrderRepository) SetAccrualResult(ctx context.Context, orderNumber string, update OrderAccrualUpdate) error {
	const query = `
		UPDATE orders
		SET
			status = $2,
			accrual = $3,
			next_poll_at = $4
		WHERE number = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		orderNumber,
		update.Status,
		update.Accrual,
		update.NextPollAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("set accrual result: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("set accrual result rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrOrderNotFound
	}

	return nil
}
