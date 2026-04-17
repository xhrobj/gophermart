package service

import (
	"context"
	"errors"

	"github.com/xhrobj/gophermart/internal/auth"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
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

// authService реализует сценарии регистрации и аутентификации пользователей.
type authService struct {

	// 1. при регистрации сделать hash / при логине проверить пароль
	passwordManager auth.PasswordManager

	// 2. создать пользователя / найти пользователя по логину
	userRepo repository.UserRepository

	// NOTE: по ТЗ после успешной регистрации должна происходить автоматическая аутентификация,
	// значит и register, и login должны в итоге возвращать bearer-токен

	// 3. выдать токен после успешного register / login
	tokenManager auth.TokenManager
}

// NewAuthService создаёт сервис регистрации и аутентификации пользователей.
func NewAuthService(
	userRepo repository.UserRepository,
	passwordManager auth.PasswordManager,
	tokenManager auth.TokenManager,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		passwordManager: passwordManager,
		tokenManager:    tokenManager,
	}
}

func (s *authService) Register(ctx context.Context, login, password string) (model.AuthResult, error) {

	// 1. получили login, password
	// 2. проверили что они не пустые
	// 3. захешировали пароль
	// 4. попросили userRepo.Create(...) создать пользователя
	//    -> если логин уже занят — вернули ErrLoginAlreadyExists
	//    -> если пользователь создался — сгенерировали jwt
	// 5. вернули AuthResult

	return model.AuthResult{
		UserID: 0,
		Token:  "",
	}, nil
}

// Login

func (s *authService) Login(ctx context.Context, login, password string) (model.AuthResult, error) {
	// 1. получили login, password
	// 2. проверили что они не пустые
	// 3. нашли пользователя по логину
	//    -> если не найден — ErrInvalidCredentials
	// 4. проверили пароль через passwordManager.Check(...)
	//    -> если пароль неверный — ErrInvalidCredentials
	//    -> если все ок — сгенерировали jwt
	// 5. вернули AuthResult

	return model.AuthResult{
		UserID: 0,
		Token:  "",
	}, nil
}
