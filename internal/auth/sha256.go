package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
)

// ErrPasswordMismatch означает, что пароль не соответствует сохраненному хешу.
var ErrPasswordMismatch = errors.New("password mismatch")

// SHA256PasswordManager хеширует и проверяет пароли с помощью SHA-256.
type SHA256PasswordManager struct{}

// NewSHA256PasswordManager создает менеджер паролей на базе SHA-256.
func NewSHA256PasswordManager() *SHA256PasswordManager {
	return &SHA256PasswordManager{}
}

// Hash возвращает SHA-256 хеш пароля в hex-формате.
func (m *SHA256PasswordManager) Hash(password string) (string, error) {
	return hash(password), nil
}

// Check проверяет, что пароль соответствует хешу.
func (m *SHA256PasswordManager) Check(password, encoded string) error {
	actual := hash(password)

	if subtle.ConstantTimeCompare([]byte(actual), []byte(encoded)) != 1 {
		return ErrPasswordMismatch
	}

	return nil
}

// hash вычисляет SHA-256 хеш строки и возвращает его в hex-формате.
func hash(x string) string {
	sum := sha256.Sum256([]byte(x))

	return hex.EncodeToString(sum[:])
}
