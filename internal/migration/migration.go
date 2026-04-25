package migration

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"

	// Регистрирует PostgreSQL-драйвер для golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	// Регистрирует file source-драйвер для чтения миграций из локальной файловой системы.
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Run применяет миграции базы данных.
func Run(databaseDSN string) error {
	if databaseDSN == "" {
		return errors.New("database connection string is empty")
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	migrationsPath := fmt.Sprintf("file://%s/migrations", wd)

	m, err := migrate.New(migrationsPath, databaseDSN)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
