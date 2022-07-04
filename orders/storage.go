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

type Withdrawal struct {
	OrderNumber int64
	Sum         int64
	ProcessedAt time.Time
}

type Storage interface {
	AddOrder(ctx context.Context, number int64, userID uint64) error
	GetOrder(ctx context.Context, number int64) (*Order, error)
	GetOrdersList(ctx context.Context, userID uint64) (*[]Order, error)
	GetBalance(ctx context.Context, userID uint64) (int64, int64, error)
	UpdateBalance(ctx context.Context, userID uint64, bal int64, wth int64) error
	AddWithdrawal(ctx context.Context, userID uint64, number int64, sum int64) error
	GetUserWithdrawals(ctx context.Context, userID uint64) (*[]Withdrawal, error)
}
