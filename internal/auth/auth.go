package auth

type PasswordManager interface {
	Hash(password string) (string, error)
	Check(password, hash string) error
}

type TokenManager interface {
	Generate(userID int64) (string, error)
	Parse(token string) (int64, error)
}
