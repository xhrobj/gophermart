package service

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

// AuthService описывает операции регистрации и аутентификации пользователей.
type AuthService interface {
	// Register регистрирует нового пользователя по логину и паролю.
	Register(ctx context.Context, login, password string) (model.AuthResult, error)

	// Login аутентифицирует пользователя по логину и паролю.
	Login(ctx context.Context, login, password string) (model.AuthResult, error)
}
