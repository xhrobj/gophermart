package model

import "time"

// User описывает зарегистрированного пользователя Gophermart.
type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

// AuthResult содержит результат успешной регистрации или аутентификации.
type AuthResult struct {
	UserID int64
	Token  string
}
