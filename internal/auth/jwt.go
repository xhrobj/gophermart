package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken означает, что токен нельзя использовать для аутентификации.
// Если токен: битый, просроченный, подписан не тем ключом, неправильного формата, без userID
var ErrInvalidToken = errors.New("invalid token")

type jwtClaims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

// JWTTokenManager создаёт и проверяет JWT-токены пользователей.
type JWTTokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTTokenManager создаёт менеджер JWT-токенов.
func NewJWTTokenManager(secret string, ttl time.Duration) *JWTTokenManager {
	return &JWTTokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// Generate создаёт подписанный JWT-токен для указанного пользователя.
func (m *JWTTokenManager) Generate(userID int64) (string, error) {
	now := time.Now()

	claims := jwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign jwt token: %w", err)
	}

	return signedToken, nil
}

// Parse проверяет JWT-токен и возвращает идентификатор пользователя из claims.
func (m *JWTTokenManager) Parse(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}

		return m.secret, nil
	})
	if err != nil {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return 0, ErrInvalidToken
	}

	if !token.Valid {
		return 0, ErrInvalidToken
	}

	if claims.UserID <= 0 {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}
