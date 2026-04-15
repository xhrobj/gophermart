package service

import (
	"context"

	"github.com/xhrobj/gophermart/internal/model"
)

type AuthService interface {
	Register(ctx context.Context, login, password string) (model.User, error)
	Login(ctx context.Context, login, password string) (model.User, error)
}
