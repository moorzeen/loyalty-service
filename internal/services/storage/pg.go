package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
	connection *pgxpool.Pool
}

// Migrate – creates DB tables if not exists
func Migrate(databaseURL string) error {
	m, err := migrate.New("file://internal/services/storage/", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to init DB migrations: %w", err)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

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

func (s *DB) AddUser(login, passHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlStatement := `INSERT INTO users (user_login, password_hash) VALUES ($1, $2)`
	_, err := s.connection.Exec(ctx, sqlStatement, login, passHash)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return ErrLoginTaken
	}

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *DB) IsUser(login, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userLogin, passwordHash string
	err := s.connection.QueryRow(ctx, "SELECT user_login, password_hash FROM users WHERE user_login = $1 AND password_hash = $2", login, hash).
		Scan(&userLogin, &passwordHash)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidUser
	}

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return nil
}
