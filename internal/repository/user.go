package repository

import (
	"context"
	"errors"

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
