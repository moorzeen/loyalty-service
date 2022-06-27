package storage

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
	connection *pgxpool.Pool
}

// Migrate – creates DB tables if not exists
func Migrate(databaseURL string) error {
	m, err := migrate.New("file://internal/storage/", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to init DB migrations: %w", err)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

// NewConnection – создает новое соединение с БД
func NewConnection(ctx context.Context, connString string) (Storage, error) {
	storage := &DB{}
	var err error

	storage.connection, err = pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new connection pool: %w", err)
	}
	//defer connection.Close()

	return storage, nil
}
