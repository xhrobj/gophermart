package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// ErrPasswordMismatch означает, что пароль не соответствует сохраненному хешу.
var ErrPasswordMismatch = errors.New("password mismatch")

// BcryptPasswordManager хеширует и проверяет пароли с помощью bcrypt.
type BcryptPasswordManager struct {
	cost int
}

// NewBcryptPasswordManager создает менеджер паролей на базе bcrypt.
func NewBcryptPasswordManager() *BcryptPasswordManager {
	return &BcryptPasswordManager{
		cost: bcrypt.DefaultCost,
	}
}

// Hash возвращает bcrypt-хеш пароля.
func (m *BcryptPasswordManager) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), m.cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// Check проверяет, что пароль соответствует хешу.
func (m *BcryptPasswordManager) Check(password, encoded string) error {
	err := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password))
	if err == nil {
		return nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrPasswordMismatch
	}

	return err
}
