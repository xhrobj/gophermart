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
	ErrPasswordTooLong    = errors.New("password too long")
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

func (s *authService) Register(ctx context.Context, login, password string) (model.AuthResult, error) {
	login = strings.TrimSpace(login)

	if login == "" || password == "" || strings.TrimSpace(password) == "" {
		return model.AuthResult{}, ErrInvalidAuthInput
	}

	passwordHash, err := s.passwordManager.Hash(password)
	if err != nil {
		if errors.Is(err, auth.ErrPasswordTooLong) {
			return model.AuthResult{}, ErrPasswordTooLong
		}

		return model.AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, login, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return model.AuthResult{}, ErrLoginAlreadyExists
		}

		return model.AuthResult{}, fmt.Errorf("create user: %w", err)
	}

	token, err := s.tokenManager.Generate(user.ID)
	if err != nil {
		return model.AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	return model.AuthResult{
		UserID: user.ID,
		Token:  token,
	}, nil
}

func (s *authService) Login(ctx context.Context, login, password string) (model.AuthResult, error) {
	login = strings.TrimSpace(login)

	if login == "" || password == "" {
		return model.AuthResult{}, ErrInvalidAuthInput
	}

	user, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return model.AuthResult{}, ErrInvalidCredentials
		}

		return model.AuthResult{}, fmt.Errorf("find user by login: %w", err)
	}

	if err := s.passwordManager.Check(password, user.PasswordHash); err != nil {
		return model.AuthResult{}, ErrInvalidCredentials
	}

	token, err := s.tokenManager.Generate(user.ID)
	if err != nil {
		return model.AuthResult{}, fmt.Errorf("generate token: %w", err)
	}

	return model.AuthResult{
		UserID: user.ID,
		Token:  token,
	}, nil
}
