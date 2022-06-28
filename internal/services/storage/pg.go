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
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DB struct {
	Connection *pgxpool.Pool
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

	storage.Connection, err = pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new connection pool: %w", err)
	}

	return storage, nil
}

func (s *DB) AddUser(login, passHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO users (user_login, password_hash) VALUES ($1, $2)`
	_, err := s.Connection.Exec(ctx, query, login, passHash)
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

func (s *DB) GetUser(login string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u User
	query := `SELECT user_login, password_hash, session_uuid FROM users WHERE user_login = $1`
	err := s.Connection.QueryRow(ctx, query, login).
		Scan(&u.Login, &u.PasswordHash, &u.SessionUUID)

	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrInvalidUser
	}

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return u, err
	}

	return u, nil
}

func (s *DB) SetSession(login string, token uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET session_uuid = $1 WHERE user_login = $2`
	_, err := s.Connection.Exec(ctx, query, token, login)
	if err != nil {
		return err
	}

	return nil
}
