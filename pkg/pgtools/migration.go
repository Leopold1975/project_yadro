package pgtools

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // used for migrate tool.
	_ "github.com/golang-migrate/migrate/v4/source/file"       // used for migrate tool.
	_ "github.com/golang-migrate/migrate/v4/source/github"     // used for migrate tool.
)

// Применить миграции до целевой версии cfg.Version.
// Если текущая версия < целевой версии, применяются миграции Up.
// Если текущая версия > целевой версии, применяются миграции Down.
// connString := "postgres://Username:Password@Addr/DB?sslmode=SSLmode
func ApplyMigration(connString string, version int) error {
	migrationsDir := "file://migrations"

	m, err := migrate.New(migrationsDir, connString)
	if err != nil {
		return fmt.Errorf("migrate new error: %w", err)
	}

	v, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		if dirty {
			err = m.Force(1)
			if err != nil {
				return fmt.Errorf("force migration error %w", err)
			}
		}

		return fmt.Errorf("migrate version error %w", err)
	}

	if version == int(v) {
		return nil
	}

	steps := version - int(v) // Вычисление количества шагов до достижения целевой версии.

	if err := m.Steps(steps); err != nil {
		if v == 0 {
			v = 1 // Базовая версия миграций.
		}

		if errF := m.Force(int(v)); errF != nil { // Откат к версии при последнем успешном применении миграций.
			return fmt.Errorf("migrate up error: %w force error: %w", err, errF)
		}

		return fmt.Errorf("migrate up error: %w", err)
	}

	return nil
}
