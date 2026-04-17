package service

import (
	"context"
	"errors"

	"github.com/xhrobj/gophermart/internal/model"
)

var (
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// AuthService описывает операции регистрации и аутентификации пользователей.
type AuthService interface {
	// Register регистрирует нового пользователя по логину и паролю.
	Register(ctx context.Context, login, password string) (model.AuthResult, error)

	// Login аутентифицирует пользователя по логину и паролю.
	Login(ctx context.Context, login, password string) (model.AuthResult, error)
}

// NOTE: по ТЗ после успешной регистрации должна происходить автоматическая аутентификация,
// значит и register, и login должны в итоге возвращать bearer-токен
