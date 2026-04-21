package auth

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestBcryptPasswordManager_HashAndCheck_OK(t *testing.T) {
	manager := NewBcryptPasswordManager()

	hash, err := manager.Hash("secret")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NotEqual(t, "secret", hash)

	err = manager.Check("secret", hash)
	require.NoError(t, err)
}

func TestBcryptPasswordManager_Check_Mismatch(t *testing.T) {
	manager := NewBcryptPasswordManager()

	hash, err := manager.Hash("secret")
	require.NoError(t, err)

	err = manager.Check("wrong", hash)
	require.ErrorIs(t, err, ErrPasswordMismatch)
}

func TestBcryptPasswordManager_Hash_PasswordTooLong(t *testing.T) {
	manager := NewBcryptPasswordManager()

	// bcrypt не принимает пароли длиннее 72 байт.
	// Это поведение самого пакета: GenerateFromPassword возвращает
	// ErrPasswordTooLong, если длина пароля превышает 72 байта.
	password := strings.Repeat("a", 73)

	_, err := manager.Hash(password)
	require.Error(t, err)
	require.True(t, errors.Is(err, bcrypt.ErrPasswordTooLong))
}
