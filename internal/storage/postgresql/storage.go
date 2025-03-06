package psql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

// New создает новое подключение к PostgreSQL по переданному DSN (storagePath).
func New(storagePath string) (*pgx.Conn, error) {
	const operation = "storage.postgresql.New"

	if storagePath == "" {
		return nil, fmt.Errorf("%s: storagePath is empty", operation)
	}

	db, err := pgx.Connect(context.Background(), storagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to connect to database: %v\n", operation, err)
		return nil, err
	}

	// Выполняем миграции при подключении
	if err := ApplyMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "%s: migration error: %v\n", operation, err)
		return nil, err
	}

	return db, nil
}

// ApplyMigrations применяет миграции к БД
func ApplyMigrations(conn *pgx.Conn) error {
	const operation = "storage.postgresql.ApplyMigrations"

	migrationPath := filepath.Join("migrations", "000_init.up.sql")
	migrationFile, err := os.ReadFile(migrationPath)

	if err != nil {
		return fmt.Errorf("%s: failed to read migration file: %w", operation, err)
	}

	_, err = conn.Exec(context.Background(), string(migrationFile))
	if err != nil {
		return fmt.Errorf("%s: failed to apply migrations script: %w", operation, err)
	}

	fmt.Println("Migrations applied successfully")
	return nil
}
