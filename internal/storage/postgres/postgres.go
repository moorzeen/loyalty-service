package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/moorzeen/loyalty-service/internal/storage"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewStorage(link string) (storage.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new connection pool: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) GetUnprocessedOrder() ([]int64, error) {
	var result []int64

	query := `UPDATE user_orders SET in_buffer = true WHERE status in ('NEW') AND in_buffer = false RETURNING order_number`

	rows, err := db.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var o int64
		err = rows.Scan(&o)
		if err != nil {
			return nil, err
		}
		result = append(result, o)
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *DB) UpdateOrder(accrual storage.Accrual) (uint64, error) {
	var result uint64

	updateQuery := `UPDATE user_orders SET status = $1, accrual = $2 WHERE order_number = $3 RETURNING user_id`
	rows, err := db.pool.Query(context.Background(), updateQuery, accrual.Status, accrual.Accrual, accrual.OrderNumber)
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		var o uint64
		err = rows.Scan(&o)
		if err != nil {
			return 0, err
		}
		result = o
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return 0, err
	}

	return result, nil
}
func (db *DB) UpdateBalance2(userID uint64, acc float64) error {

	log.Printf("db.UpdateBalance/ userID: %d, acrrual: %f", userID, acc)

	updateQuery := `UPDATE accounts SET balance = balance + $1 WHERE user_id = $2`
	_, err := db.pool.Exec(context.Background(), updateQuery, acc, userID)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) AddUser(ctx context.Context, username string, passwordHash []byte) error {
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2)`

	_, err := db.pool.Exec(ctx, query, username, passwordHash)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) AddAccount(ctx context.Context, userID uint64) error {
	query := `INSERT INTO accounts (user_id, balance, withdrawn) VALUES ($1, $2, $3)`

	_, err := db.pool.Exec(ctx, query, userID, 0, 0)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetUser(ctx context.Context, username string) (*storage.User, error) {
	user := &storage.User{}

	query := `SELECT username, password_hash, id FROM users WHERE username = $1`

	err := db.pool.QueryRow(ctx, query, username).Scan(&user.Username, &user.PasswordHash, &user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *DB) SetSession(ctx context.Context, userID uint64, signKey []byte) error {
	query := `INSERT INTO user_sessions (user_id, sign_key) VALUES($1, $2) ON CONFLICT (user_id) DO UPDATE SET sign_key = $2`

	_, err := db.pool.Exec(ctx, query, userID, signKey)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetSession(ctx context.Context, userID uint64) (*storage.Session, error) {
	session := &storage.Session{}

	query := `SELECT user_id, sign_key FROM user_sessions WHERE user_id = $1`

	err := db.pool.QueryRow(ctx, query, userID).Scan(&session.UserID, &session.SignKey)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (db *DB) AddOrder(ctx context.Context, number int64, userID uint64) error {
	query := `INSERT INTO user_orders (order_number, user_id, status) VALUES ($1, $2, $3)`

	_, err := db.pool.Exec(ctx, query, number, userID, "NEW")
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetOrder(ctx context.Context, number int64) (*storage.Order, error) {
	order := &storage.Order{}

	query := `SELECT order_number, user_id, status, uploaded_at, accrual FROM user_orders WHERE order_number = $1`

	err := db.pool.QueryRow(ctx, query, number).Scan(
		&order.OrderNumber,
		&order.UserID,
		&order.Status,
		&order.UploadedAt,
		&order.Accrual,
	)
	if err != nil {
		return order, err
	}

	return order, nil
}

func (db *DB) GetOrdersList(ctx context.Context, userID uint64) (*[]storage.Order, error) {

	var result []storage.Order

	query := `SELECT user_id, order_number, status, uploaded_at, accrual
				FROM user_orders WHERE user_id = $1 order by uploaded_at`
	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		log.Println(err)
		return &result, err
	}

	for rows.Next() {
		var o storage.Order
		err = rows.Scan(&o.UserID, &o.OrderNumber, &o.Status, &o.UploadedAt, &o.Accrual)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		result = append(result, o)
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &result, nil
}

func (db *DB) GetBalance(ctx context.Context, userID uint64) (float64, float64, error) {
	var bal, wtn float64

	balQuery := `SELECT balance, withdrawn FROM accounts WHERE user_id = $1`
	err := db.pool.QueryRow(ctx, balQuery, userID).Scan(&bal, &wtn)
	if err == pgx.ErrNoRows {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}

	return bal, wtn, nil
}

func (db *DB) UpdateBalance(ctx context.Context, userID uint64, bal float64, wth float64) error {
	updateQuery := `UPDATE accounts SET balance = $1, withdrawn = $2 WHERE user_id = $3`
	_, err := db.pool.Exec(ctx, updateQuery, bal, wth, userID)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) AddWithdrawal(ctx context.Context, userID uint64, number int64, sum float64) error {
	query := `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`

	_, err := db.pool.Exec(ctx, query, userID, number, sum)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetUserWithdrawals(ctx context.Context, userID uint64) (*[]storage.Withdrawal, error) {
	var result []storage.Withdrawal

	query := `SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 order by processed_at`
	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		log.Println(err)
		return &result, err
	}

	for rows.Next() {
		var o storage.Withdrawal
		err = rows.Scan(&o.OrderNumber, &o.Sum, &o.ProcessedAt)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		result = append(result, o)
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &result, nil

}
