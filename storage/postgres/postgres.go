package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/moorzeen/loyalty-service/internal/services/auth"
	"github.com/moorzeen/loyalty-service/storage"
)

type DB struct {
	Connection *pgxpool.Pool
}

// Migrate – creates DB tables if not exists
func Migrate(databaseURL string) error {
	m, err := migrate.New("file://storage/postgres/", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to init DB migrations: %w", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

func Open(ctx context.Context, connString string) (storage.Service, error) {
	db := &DB{}
	var err error

	db.Connection, err = pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new connection pool: %w", err)
	}

	return db, nil
}

func (s *DB) AddUser(login string, passHash []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO users (user_login, password_hash) VALUES ($1, $2)`
	_, err := s.Connection.Exec(ctx, query, login, passHash)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return storage.ErrLoginTaken
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *DB) GetUserByLogin(login string) (auth.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u auth.User
	query := `SELECT user_login, password_hash, id FROM users WHERE user_login = $1`
	err := s.Connection.QueryRow(ctx, query, login).
		Scan(&u.Login, &u.PasswordHash, &u.ID)

	if errors.Is(err, pgx.ErrNoRows) {
		return u, storage.ErrInvalidUser
	}
	if err != nil {
		return u, err
	}

	return u, nil
}

func (s *DB) AddSession(session auth.Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT into USER_SESSIONS (USERS_ID, SIGN_KEY) values($1, $2) on conflict (USERS_ID) do update set SIGN_KEY = $2`
	_, err := s.Connection.Exec(ctx, query, session.UserID, session.SignatureKey)
	if err != nil {
		return err
	}

	return nil
}
