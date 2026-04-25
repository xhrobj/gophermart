package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const (
	jwtTestSecret = "secret"
	jwtTestUserID = int64(42)
)

func TestJWTTokenManager_GenerateAndParse_OK(t *testing.T) {
	manager := NewJWTTokenManager(jwtTestSecret, time.Hour)

	token, err := manager.Generate(jwtTestUserID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	userID, err := manager.Parse(token)
	require.NoError(t, err)
	require.Equal(t, jwtTestUserID, userID)
}

func TestJWTTokenManager_Parse_InvalidToken(t *testing.T) {
	manager := NewJWTTokenManager(jwtTestSecret, time.Hour)

	userID, err := manager.Parse("invalid-token")

	require.ErrorIs(t, err, ErrInvalidToken)
	require.Zero(t, userID)
}

func TestJWTTokenManager_Parse_WrongSecret(t *testing.T) {
	issuer := NewJWTTokenManager(jwtTestSecret, time.Hour)
	parser := NewJWTTokenManager("another-secret", time.Hour)

	token, err := issuer.Generate(jwtTestUserID)
	require.NoError(t, err)

	userID, err := parser.Parse(token)

	require.ErrorIs(t, err, ErrInvalidToken)
	require.Zero(t, userID)
}

func TestJWTTokenManager_Parse_ExpiredToken(t *testing.T) {
	manager := NewJWTTokenManager(jwtTestSecret, -time.Hour)

	token, err := manager.Generate(jwtTestUserID)
	require.NoError(t, err)

	userID, err := manager.Parse(token)

	require.ErrorIs(t, err, ErrInvalidToken)
	require.Zero(t, userID)
}

func TestJWTTokenManager_Parse_InvalidUserID(t *testing.T) {
	manager := NewJWTTokenManager(jwtTestSecret, time.Hour)

	token, err := manager.Generate(0)
	require.NoError(t, err)

	userID, err := manager.Parse(token)

	require.ErrorIs(t, err, ErrInvalidToken)
	require.Zero(t, userID)
}

func TestJWTTokenManager_Parse_UnsupportedSigningMethod(t *testing.T) {
	claims := authClaims{
		UserID: jwtTestUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString([]byte(jwtTestSecret))
	require.NoError(t, err)

	manager := NewJWTTokenManager(jwtTestSecret, time.Hour)

	userID, err := manager.Parse(token)

	require.ErrorIs(t, err, ErrInvalidToken)
	require.Zero(t, userID)
}
