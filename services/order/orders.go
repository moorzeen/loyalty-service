package order

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/moorzeen/loyalty-service/services/order/helpers"
)

type Service struct {
	storage Storage
}

type WithdrawRequest struct {
	UserID      uint64
	OrderNumber string  `json:"order"`
	WithdrawSum float64 `json:"sum"`
}

func NewService(str Storage) *Service {
	return &Service{storage: str}
}

func (o *Service) AddOrder(ctx context.Context, orderNumber string, userID uint64) error {
	number, err := helpers.ParseOrderNumber(orderNumber)
	if err != nil {
		log.Println(err)
		return ErrInvalidOrderNumber
	}

	err = o.storage.AddOrder(ctx, number, userID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		order, err := o.storage.GetOrder(ctx, number)
		if err != nil {
			return err
		}

		if order.UserID == userID {
			return ErrAlreadyAddByThis
		}

		return ErrAddedByOther
	}

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (o *Service) GetOrders(ctx context.Context, userID uint64) (*[]Order, error) {
	orders, err := o.storage.GetOrdersList(ctx, userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return orders, nil
}

func (o *Service) GetBalance(ctx context.Context, userID uint64) (float64, float64, error) {
	bal, wtn, err := o.storage.GetBalance(ctx, userID)
	if err != nil {
		log.Println(err)
		return 0, 0, err
	}

	return bal, wtn, nil
}

func (o *Service) Withdraw(ctx context.Context, request WithdrawRequest) error {

	number, err := helpers.ParseOrderNumber(request.OrderNumber)
	if err != nil {
		log.Println(err)
		return ErrInvalidOrderNumber
	}

	bal, wtn, err := o.GetBalance(ctx, request.UserID)
	if err != nil {
		return err
	}

	if request.WithdrawSum > bal {
		return ErrInsufficientFunds
	}

	err = o.storage.AddWithdrawal(ctx, request.UserID, number, request.WithdrawSum)
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

func (o *Service) GetWithdrawals(ctx context.Context, userID uint64) (*[]Withdrawal, error) {
	withdrawals, err := o.storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return withdrawals, nil
}
