package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xhrobj/gophermart/internal/auth"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
)

var (
	ErrInvalidAuthInput   = errors.New("invalid auth input")
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

	// создать пользователя / найти пользователя по логину
	userRepo repository.UserRepository

	// при регистрации сделать hash / при логине проверить пароль
	passwordManager auth.PasswordManager

	// выдать токен после успешного register / login
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

// Register регистрирует нового пользователя и выдаёт токен аутентификации.
func (s *authService) Register(ctx context.Context, login, password string) (model.AuthResult, error) {

	// 1. получили login, password

	login = strings.TrimSpace(login)

	// 2. проверили что они не пустые

	if login == "" || password == "" || strings.TrimSpace(password) == "" {
		return model.AuthResult{}, ErrInvalidAuthInput
	}

	// 3. захешировали пароль

	passwordHash, err := s.passwordManager.Hash(password)
	if err != nil {
		return model.AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	// 4. попросили userRepo.Create(...) создать пользователя

	user, err := s.userRepo.Create(ctx, login, passwordHash)

	// -> если логин уже занят — вернули ErrLoginAlreadyExists

	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return model.AuthResult{}, ErrLoginAlreadyExists
		}

		return model.AuthResult{}, fmt.Errorf("create user: %w", err)
	}

	// -> если пользователь создался — сгенерировали jwt

	token, err := s.tokenManager.Generate(user.ID)
	if err != nil {
		return model.AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	// 5. вернули AuthResult

	return model.AuthResult{
		UserID: user.ID,
		Token:  token,
	}, nil
}

// Login аутентифицирует пользователя по логину и паролю.
func (s *authService) Login(ctx context.Context, login, password string) (model.AuthResult, error) {
	// 1. получили login, password

	login = strings.TrimSpace(login)

	// 2. проверили что они не пустые

	if login == "" || password == "" {
		return model.AuthResult{}, ErrInvalidAuthInput
	}

	// 3. нашли пользователя по логину

	user, err := s.userRepo.FindByLogin(ctx, login)

	// -> если не найден — ErrInvalidCredentials

	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return model.AuthResult{}, ErrInvalidCredentials
		}

		return model.AuthResult{}, fmt.Errorf("find user by login: %w", err)
	}

	// 4. проверили пароль через passwordManager.Check(...)
	// -> если пароль неверный — ErrInvalidCredentials

	if err := s.passwordManager.Check(password, user.PasswordHash); err != nil {
		return model.AuthResult{}, ErrInvalidCredentials
	}

	// -> если все ок — сгенерировали jwt

	token, err := s.tokenManager.Generate(user.ID)
	if err != nil {
		return model.AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	// 5. вернули AuthResult

	return model.AuthResult{
		UserID: user.ID,
		Token:  token,
	}, nil
}
