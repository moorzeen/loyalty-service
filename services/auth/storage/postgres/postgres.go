package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/moorzeen/loyalty-service/services/auth"
)

type Postgres struct {
	connection *pgxpool.Pool
}

func NewStorage(conn *pgxpool.Pool) auth.Storage {
	return &Postgres{connection: conn}
}

func (db *Postgres) AddUser(ctx context.Context, username string, passwordHash []byte) error {
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2)`

	_, err := db.connection.Exec(ctx, query, username, passwordHash)
	if err != nil {
		return err
	}

	return nil
}

func (db *Postgres) AddAccount(ctx context.Context, userID uint64) error {
	query := `INSERT INTO accounts (user_id, balance, withdrawn) VALUES ($1, $2, $3)`

	_, err := db.connection.Exec(ctx, query, userID, 0, 0)
	if err != nil {
		return err
	}

	return nil
}

func (db *Postgres) GetUser(ctx context.Context, username string) (*auth.User, error) {
	user := &auth.User{}

	query := `SELECT username, password_hash, id FROM users WHERE username = $1`

	err := db.connection.QueryRow(ctx, query, username).Scan(&user.Username, &user.PasswordHash, &user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *Postgres) SetSession(ctx context.Context, userID uint64, signKey []byte) error {
	query := `INSERT INTO user_sessions (user_id, sign_key) VALUES($1, $2) ON CONFLICT (user_id) DO UPDATE SET sign_key = $2`

	_, err := db.connection.Exec(ctx, query, userID, signKey)
	if err != nil {
		return err
	}

	return nil
}

func (db *Postgres) GetSession(ctx context.Context, userID uint64) (*auth.Session, error) {
	session := &auth.Session{}

	query := `SELECT user_id, sign_key FROM user_sessions WHERE user_id = $1`

	err := db.connection.QueryRow(ctx, query, userID).Scan(&session.UserID, &session.SignKey)
	if err != nil {
		return nil, err
	}

	return session, nil
}
