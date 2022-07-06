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
	OrderNumber int64   `json:"order"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
}

type Order struct {
	OrderNumber int64
	UserID      uint64
	UploadedAt  time.Time
	Status      string
	Accrual     float64
}

type Withdrawal struct {
	OrderNumber int64
	Sum         float64
	ProcessedAt time.Time
}

type Service interface {
	AddUser(ctx context.Context, username string, passwordHash []byte) error
	GetUser(ctx context.Context, username string) (*User, error)
	SetSession(ctx context.Context, userID uint64, signKey []byte) error
	GetSession(ctx context.Context, userID uint64) (*Session, error)
	AddAccount(ctx context.Context, userID uint64) error

	AddOrder(ctx context.Context, number int64, userID uint64) error
	GetOrder(ctx context.Context, number int64) (*Order, error)
	GetOrdersList(ctx context.Context, userID uint64) (*[]Order, error)
	GetBalance(ctx context.Context, userID uint64) (float64, float64, error)
	UpdateBalance(ctx context.Context, userID uint64, bal float64, wth float64) error
	AddWithdrawal(ctx context.Context, userID uint64, number int64, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID uint64) (*[]Withdrawal, error)

	GetUnprocessedOrder() ([]int64, error)
	UpdateOrder(accrual Accrual) (uint64, error)
	UpdateBalance2(userID uint64, acc float64) error
}
