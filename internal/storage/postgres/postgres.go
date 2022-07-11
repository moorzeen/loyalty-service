package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/moorzeen/loyalty-service/internal/order"
	"github.com/moorzeen/loyalty-service/internal/storage"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewStorage(ctx context.Context, link string) (storage.Service, error) {
	pool, err := pgxpool.Connect(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new connection pool: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) AddUser(ctx context.Context, username string, passwordHash []byte) (uint64, error) {
	var userID uint64

	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`
	err := db.pool.QueryRow(ctx, query, username, passwordHash).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (db *DB) AddAccount(ctx context.Context, userID uint64) error {
	query := `INSERT INTO accounts (user_id) VALUES ($1)`
	_, err := db.pool.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetUser(ctx context.Context, username string) (*storage.User, error) {
	user := &storage.User{}

	query := `SELECT id, username, password_hash FROM users WHERE username = $1`
	err := db.pool.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *DB) SetSession(ctx context.Context, userID uint64, signKey []byte) error {
	query := `INSERT INTO sessions (user_id, sign_key) VALUES($1, $2) ON CONFLICT (user_id) DO UPDATE SET sign_key = $2`
	_, err := db.pool.Exec(ctx, query, userID, signKey)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetSession(ctx context.Context, userID uint64) (*storage.Session, error) {
	session := &storage.Session{}

	query := `SELECT user_id, sign_key FROM sessions WHERE user_id = $1`
	err := db.pool.QueryRow(ctx, query, userID).Scan(&session.UserID, &session.SignKey)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (db *DB) AddOrder(ctx context.Context, number string, userID uint64) error {
	query := `INSERT INTO orders (order_number, user_id, status) VALUES ($1, $2, $3)`
	_, err := db.pool.Exec(ctx, query, number, userID, "NEW")
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetOrder(ctx context.Context, number string) (*storage.Order, error) {
	order := &storage.Order{}

	query := `SELECT order_number, user_id, status, uploaded_at, accrual FROM orders WHERE order_number = $1`
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

func (db *DB) GetOrders(ctx context.Context, userID uint64) ([]storage.Order, error) {
	var orders []storage.Order

	query := `SELECT user_id, order_number, status, uploaded_at, accrual
				FROM orders WHERE user_id = $1 order by uploaded_at`
	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var o storage.Order
		err = rows.Scan(&o.UserID, &o.OrderNumber, &o.Status, &o.UploadedAt, &o.Accrual)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func (db *DB) GetBalance(ctx context.Context, userID uint64) (float64, float64, error) {
	var bal, wtn float64

	balQuery := `SELECT balance, withdrawn FROM accounts WHERE user_id = $1`
	err := db.pool.QueryRow(ctx, balQuery, userID).Scan(&bal, &wtn)
	if err != nil {
		return 0, 0, err
	}

	return bal, wtn, nil
}

func (db *DB) Withdraw(ctx context.Context, userID uint64, number string, wth float64) error {
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	// confirm that funds is enough for the withdrawal
	var balance float64
	balanceQuery := `SELECT balance FROM accounts WHERE user_id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, balanceQuery, userID).Scan(&balance)
	if err != nil {
		return err
	}
	if wth > balance {
		return order.ErrInsufficientFunds
	}

	// create a new row in the withdrawals table
	addWithdrawQuery := `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, addWithdrawQuery, userID, number, wth)
	if err != nil {
		return err
	}

	// update balance
	updateBalanceQuery := `UPDATE accounts SET balance = balance - $1, withdrawn = withdrawn + $1 WHERE user_id = $2`
	_, err = tx.Exec(ctx, updateBalanceQuery, wth, userID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetWithdrawals(ctx context.Context, userID uint64) ([]storage.Withdrawal, error) {
	var result []storage.Withdrawal

	query := `SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 order by processed_at`
	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		return result, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var o storage.Withdrawal
		err = rows.Scan(&o.OrderNumber, &o.Sum, &o.ProcessedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, o)
	}

	return result, nil

}

func (db *DB) GetUnprocessedOrder() ([]string, error) {
	var orders []string

	query := `UPDATE orders SET status = 'PROCESSING' WHERE status = 'NEW' RETURNING order_number`
	rows, err := db.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var o string
		err = rows.Scan(&o)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func (db *DB) UpdateOrder(accrual storage.Accrual) (uint64, error) {
	var result uint64

	updateQuery := `UPDATE orders SET status = $1, accrual = $2 WHERE order_number = $3 RETURNING user_id`
	rows, err := db.pool.Query(context.Background(), updateQuery, accrual.Status, accrual.Accrual, accrual.OrderNumber)
	if err != nil {
		return 0, err
	}

	err = rows.Err()
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

	return result, nil
}
func (db *DB) Accrual(userID uint64, acc float64) error {
	updateQuery := `UPDATE accounts SET balance = balance + $1 WHERE user_id = $2`
	_, err := db.pool.Exec(context.Background(), updateQuery, acc, userID)
	if err != nil {
		return err
	}
	return nil
}
