package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/xhrobj/gophermart/internal/model"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserRepository описывает операции хранения пользователей.
type UserRepository interface {
	// Create создает нового пользователя с уже подготовленным хешем пароля.
	Create(ctx context.Context, login, passwordHash string) (model.User, error)

	// FindByLogin возвращает пользователя по логину.
	FindByLogin(ctx context.Context, login string) (model.User, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, login, passwordHash string) (model.User, error) {
	const query = "INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, login, password_hash, created_at"

	var user model.User

	err := r.db.QueryRowContext(ctx, query, login, passwordHash).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return model.User{}, ErrUserAlreadyExists
		}

		return model.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) FindByLogin(ctx context.Context, login string) (model.User, error) {
	const query = "SELECT id, login, password_hash, created_at FROM users WHERE login = $1"

	var user model.User

	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}

		return model.User{}, fmt.Errorf("find user by login: %w", err)
	}

	return user, nil
}
