package postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/moorzeen/loyalty-service/orders"
)

type Postgres struct {
	connection *pgxpool.Pool
}

func NewStorage(conn *pgxpool.Pool) orders.Storage {
	return &Postgres{connection: conn}
}

func (db *Postgres) AddOrder(ctx context.Context, number int64, userID uint64) error {
	query := `INSERT INTO user_orders (order_number, user_id, status) VALUES ($1, $2, $3)`

	_, err := db.connection.Exec(ctx, query, number, userID, "NEW")
	if err != nil {
		return err
	}

	return nil
}

func (db *Postgres) GetOrder(ctx context.Context, number int64) (*orders.Order, error) {
	order := &orders.Order{}

	query := `SELECT order_number, user_id, status, uploaded_at, accrual FROM user_orders WHERE order_number = $1`

	err := db.connection.QueryRow(ctx, query, number).Scan(
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

func (db *Postgres) GetOrdersList(ctx context.Context, userID uint64) (*[]orders.Order, error) {

	var result []orders.Order

	query := `SELECT user_id, order_number, status, uploaded_at, accrual
				FROM user_orders WHERE user_id = $1 order by uploaded_at`
	rows, err := db.connection.Query(ctx, query, userID)
	if err != nil {
		log.Println(err)
		return &result, err
	}

	for rows.Next() {
		var o orders.Order
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

func (db *Postgres) GetBalance(ctx context.Context, userID uint64) (int64, int64, error) {
	var bal, wtn int64

	balQuery := `SELECT balance, withdrawn FROM accounts WHERE user_id = $1`
	err := db.connection.QueryRow(ctx, balQuery, userID).Scan(&bal, &wtn)
	if err == pgx.ErrNoRows {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}

	return bal, wtn, nil
}

func (db *Postgres) UpdateBalance(ctx context.Context, userID uint64, bal int64, wth int64) error {
	updateQuery := `UPDATE accounts SET balance = $1, withdrawn = $2 WHERE user_id = $3`
	_, err := db.connection.Exec(ctx, updateQuery, bal, wth, userID)
	if err != nil {
		return err
	}
	return nil
}

func (db *Postgres) AddWithdrawal(ctx context.Context, userID uint64, number int64, sum int64) error {
	query := `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`

	_, err := db.connection.Exec(ctx, query, userID, number, sum)
	if err != nil {
		return err
	}

	return nil
}

func (db *Postgres) GetUserWithdrawals(ctx context.Context, userID uint64) (*[]orders.Withdrawal, error) {
	var result []orders.Withdrawal

	query := `SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 order by processed_at`
	rows, err := db.connection.Query(ctx, query, userID)
	if err != nil {
		log.Println(err)
		return &result, err
	}

	for rows.Next() {
		var o orders.Withdrawal
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
