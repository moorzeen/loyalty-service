package postgres

import (
	"context"

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

func (db *Postgres) GetOrderByNumber(ctx context.Context, number int64) (*orders.Order, error) {
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
