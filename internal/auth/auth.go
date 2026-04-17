package auth

// PasswordManager описывает операции хеширования и проверки паролей.
type PasswordManager interface {
	// Hash возвращает хеш пароля.
	Hash(password string) (string, error)

	// Check проверяет, что пароль соответствует хешу.
	Check(password, hash string) error
}

// TokenManager описывает операции создания и разбора токенов аутентификации.
type TokenManager interface {
	// Generate создаёт токен аутентификации для указанного пользователя.
	Generate(userID int64) (string, error)

	// Parse проверяет токен аутентификации и возвращает идентификатор пользователя.
	Parse(token string) (int64, error)
}
