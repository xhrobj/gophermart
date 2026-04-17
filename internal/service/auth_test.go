package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
)

type stubUserRepo struct {
	createFunc      func(ctx context.Context, login, passwordHash string) (model.User, error)
	findByLoginFunc func(ctx context.Context, login string) (model.User, error)
}

func (s *stubUserRepo) Create(ctx context.Context, login, passwordHash string) (model.User, error) {
	return s.createFunc(ctx, login, passwordHash)
}

func (s *stubUserRepo) FindByLogin(ctx context.Context, login string) (model.User, error) {
	return s.findByLoginFunc(ctx, login)
}

type stubPasswordManager struct {
	hashFunc  func(password string) (string, error)
	checkFunc func(password, hash string) error
}

func (s *stubPasswordManager) Hash(password string) (string, error) {
	return s.hashFunc(password)
}

func (s *stubPasswordManager) Check(password, hash string) error {
	return s.checkFunc(password, hash)
}

type stubTokenManager struct {
	generateFunc func(userID int64) (string, error)
	parseFunc    func(token string) (int64, error)
}

func (s *stubTokenManager) Generate(userID int64) (string, error) {
	return s.generateFunc(userID)
}

func (s *stubTokenManager) Parse(token string) (int64, error) {
	return s.parseFunc(token)
}

func TestAuthService_Register_OK(t *testing.T) {
	userRepo := &stubUserRepo{
		createFunc: func(ctx context.Context, login, passwordHash string) (model.User, error) {
			require.Equal(t, "admin", login)
			require.Equal(t, "hashed-secret", passwordHash)

			return model.User{
				ID:           42,
				Login:        login,
				PasswordHash: passwordHash,
			}, nil
		},
	}

	passwordManager := &stubPasswordManager{
		hashFunc: func(password string) (string, error) {
			require.Equal(t, "secret", password)
			return "hashed-secret", nil
		},
	}

	tokenManager := &stubTokenManager{
		generateFunc: func(userID int64) (string, error) {
			require.Equal(t, int64(42), userID)
			return "jwt-token", nil
		},
	}

	svc := NewAuthService(userRepo, passwordManager, tokenManager)

	got, err := svc.Register(context.Background(), "admin", "secret")
	require.NoError(t, err)
	require.Equal(t, model.AuthResult{
		UserID: 42,
		Token:  "jwt-token",
	}, got)
}

func TestAuthService_Register_InvalidAuthInput(t *testing.T) {
	userRepo := &stubUserRepo{}
	passwordManager := &stubPasswordManager{}
	tokenManager := &stubTokenManager{}

	svc := NewAuthService(userRepo, passwordManager, tokenManager)

	_, err := svc.Register(context.Background(), "   ", "secret")
	require.ErrorIs(t, err, ErrInvalidAuthInput)
}

func TestAuthService_Register_LoginAlreadyExists(t *testing.T) {
	userRepo := &stubUserRepo{
		createFunc: func(ctx context.Context, login, passwordHash string) (model.User, error) {
			return model.User{}, repository.ErrUserAlreadyExists
		},
	}

	passwordManager := &stubPasswordManager{
		hashFunc: func(password string) (string, error) {
			return "hashed-secret", nil
		},
	}

	tokenManager := &stubTokenManager{}

	svc := NewAuthService(userRepo, passwordManager, tokenManager)

	_, err := svc.Register(context.Background(), "admin", "secret")
	require.ErrorIs(t, err, ErrLoginAlreadyExists)
}

func TestAuthService_Login_OK(t *testing.T) {
	userRepo := &stubUserRepo{
		findByLoginFunc: func(ctx context.Context, login string) (model.User, error) {
			require.Equal(t, "admin", login)

			return model.User{
				ID:           42,
				Login:        "admin",
				PasswordHash: "hashed-secret",
			}, nil
		},
	}

	passwordManager := &stubPasswordManager{
		checkFunc: func(password, hash string) error {
			require.Equal(t, "secret", password)
			require.Equal(t, "hashed-secret", hash)
			return nil
		},
	}

	tokenManager := &stubTokenManager{
		generateFunc: func(userID int64) (string, error) {
			require.Equal(t, int64(42), userID)
			return "jwt-token", nil
		},
	}

	svc := NewAuthService(userRepo, passwordManager, tokenManager)

	got, err := svc.Login(context.Background(), "admin", "secret")
	require.NoError(t, err)
	require.Equal(t, model.AuthResult{
		UserID: 42,
		Token:  "jwt-token",
	}, got)
}

func TestAuthService_Login_InvalidAuthInput(t *testing.T) {
	userRepo := &stubUserRepo{}
	passwordManager := &stubPasswordManager{}
	tokenManager := &stubTokenManager{}

	svc := NewAuthService(userRepo, passwordManager, tokenManager)

	_, err := svc.Login(context.Background(), "   ", "secret")
	require.ErrorIs(t, err, ErrInvalidAuthInput)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := &stubUserRepo{
		findByLoginFunc: func(ctx context.Context, login string) (model.User, error) {
			return model.User{}, repository.ErrUserNotFound
		},
	}

	passwordManager := &stubPasswordManager{}
	tokenManager := &stubTokenManager{}

	svc := NewAuthService(userRepo, passwordManager, tokenManager)

	_, err := svc.Login(context.Background(), "admin", "secret")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	userRepo := &stubUserRepo{
		findByLoginFunc: func(ctx context.Context, login string) (model.User, error) {
			return model.User{
				ID:           42,
				Login:        "admin",
				PasswordHash: "hashed-secret",
			}, nil
		},
	}

	passwordManager := &stubPasswordManager{
		checkFunc: func(password, hash string) error {
			require.Equal(t, "secret", password)
			require.Equal(t, "hashed-secret", hash)
			return errors.New("password mismatch")
		},
	}

	tokenManager := &stubTokenManager{}

	svc := NewAuthService(userRepo, passwordManager, tokenManager)

	_, err := svc.Login(context.Background(), "admin", "secret")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}
