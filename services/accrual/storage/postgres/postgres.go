package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/moorzeen/loyalty-service/services/accrual"
)

type Postgres struct {
	connection *pgxpool.Pool
}

func NewStorage(conn *pgxpool.Pool) accrual.Storage {
	return &Postgres{connection: conn}
}

func (db *Postgres) GetUnprocessedOrder() ([]int64, error) {
	var result []int64

	query := `UPDATE user_orders SET in_buffer = true WHERE status in ('NEW') AND in_buffer = false RETURNING order_number`

	rows, err := db.connection.Query(context.Background(), query)
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

func (db *Postgres) UpdateOrder(accrual accrual.Accrual) (uint64, error) {
	var result uint64

	updateQuery := `UPDATE user_orders SET status = $1, accrual = $2 WHERE order_number = $3 RETURNING order_number`
	rows, err := db.connection.Query(context.Background(), updateQuery, accrual.Status, accrual.Accrual, accrual.OrderNumber)
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		var o uint64
		err = rows.Scan(&o)
		if err != nil {
			return 0, err
		}
		result = 0
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return 0, err
	}

	return result, nil
}
func (db *Postgres) UpdateBalance(userID uint64, acc float64) error {

	updateQuery := `UPDATE accounts SET balance = balance + $1 WHERE user_id = $2`
	_, err := db.connection.Exec(context.Background(), updateQuery, acc, userID)
	if err != nil {
		return err
	}
	return nil
}
