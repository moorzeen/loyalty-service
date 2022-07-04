package orders

import (
	"context"
	"time"
)

type Order struct {
	OrderNumber int64
	UserID      uint64
	UploadedAt  time.Time
	Status      string
	Accrual     int64
}

type Storage interface {
	AddOrder(ctx context.Context, number int64, userID uint64) error
	GetOrderByNumber(ctx context.Context, number int64) (*Order, error)
	GetOrders(ctx context.Context, userID uint64) ([]Order, error)
}
