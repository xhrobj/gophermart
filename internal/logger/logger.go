package logger

import "go.uber.org/zap"

// New создаёт логгер приложения.
func New() (*zap.Logger, error) {
	return zap.NewDevelopment()
}
