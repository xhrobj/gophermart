package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrPasswordMismatch = errors.New("password mismatch")

type SHA256PasswordManager struct{}

func NewSHA256PasswordManager() *SHA256PasswordManager {
	return &SHA256PasswordManager{}
}

func (m *SHA256PasswordManager) Hash(password string) (string, error) {
	return hash(password), nil
}

func (m *SHA256PasswordManager) Check(password, encoded string) error {
	if hash(password) != encoded {
		return ErrPasswordMismatch
	}
	return nil
}

func hash(x string) string {
	sum := sha256.Sum256([]byte(x))
	return hex.EncodeToString(sum[:])
}
