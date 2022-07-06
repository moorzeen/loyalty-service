package order

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/moorzeen/loyalty-service/internal/storage"
)

type Service struct {
	storage storage.Service
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Balance   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdraw struct {
	UserID      uint64
	OrderNumber string  `json:"order"`
	WithdrawSum float64 `json:"sum"`
}

func NewService(str storage.Service) Service {
	return Service{storage: str}
}

func (o *Service) AddOrder(ctx context.Context, orderNumber string, userID uint64) error {

	if err := parseOrderNumber(orderNumber); err != nil {
		return ErrInvalidOrderNumber
	}

	err := o.storage.AddOrder(ctx, orderNumber, userID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		order, err := o.storage.GetOrder(ctx, orderNumber)
		if err != nil {
			return err
		}
		if order.UserID == userID {
			return ErrAlreadyAddByThis
		}
		return ErrAddedByOther
	}

	if err != nil {
		return err
	}

	return nil
}

func (o *Service) GetOrders(ctx context.Context, userID uint64) (*[]storage.Order, error) {
	orders, err := o.storage.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (o *Service) GetBalance(ctx context.Context, userID uint64) (float64, float64, error) {
	bal, wtn, err := o.storage.GetBalance(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	return bal, wtn, nil
}

func (o *Service) Withdraw(ctx context.Context, request Withdraw) error {

	if err := parseOrderNumber(request.OrderNumber); err != nil {
		return ErrInvalidOrderNumber
	}

	bal, wtn, err := o.GetBalance(ctx, request.UserID)
	if err != nil {
		return err
	}

	if request.WithdrawSum > bal {
		return ErrInsufficientFunds
	}

	err = o.storage.AddWithdrawal(ctx, request.UserID, request.OrderNumber, request.WithdrawSum)
	if err != nil {
		return err
	}

	bal -= request.WithdrawSum
	wtn += request.WithdrawSum

	err = o.storage.UpdateBalance(ctx, request.UserID, bal, wtn)
	if err != nil {
		return err
	}

	return nil
}

func (o *Service) GetWithdrawals(ctx context.Context, userID uint64) (*[]storage.Withdrawal, error) {
	withdrawals, err := o.storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}
