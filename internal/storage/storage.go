package storage

import (
	"context"
	"time"
)

type User struct {
	ID           uint64
	Username     string
	PasswordHash []byte
}

type Session struct {
	UserID  uint64
	SignKey []byte
}

type Accrual struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
}

type Order struct {
	OrderNumber string
	UserID      uint64
	UploadedAt  time.Time
	Status      string
	Accrual     float64
}

type Withdrawal struct {
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}

type Service interface {
	AddUser(ctx context.Context, username string, passwordHash []byte) (uint64, error)
	AddAccount(ctx context.Context, userID uint64) error
	GetUser(ctx context.Context, username string) (*User, error)
	SetSession(ctx context.Context, userID uint64, signKey []byte) error
	GetSession(ctx context.Context, userID uint64) (*Session, error)

	AddOrder(ctx context.Context, number string, userID uint64) error
	GetOrder(ctx context.Context, number string) (*Order, error)
	GetOrders(ctx context.Context, userID uint64) ([]Order, error)
	GetBalance(ctx context.Context, userID uint64) (float64, float64, error)
	AddWithdrawal(ctx context.Context, userID uint64, number string, sum float64) error
	UpdateBalance(ctx context.Context, userID uint64, bal float64, wth float64) error
	GetWithdrawals(ctx context.Context, userID uint64) ([]Withdrawal, error)

	GetUnprocessedOrder() ([]string, error)
	UpdateOrder(accrual Accrual) (uint64, error)
	Accrual(userID uint64, acc float64) error
}
