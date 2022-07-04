package orders

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/moorzeen/loyalty-service/orders/helpers"
)

type Orders struct {
	storage Storage
}

type WithdrawRequest struct {
	UserID      uint64
	OrderNumber string `json:"order"`
	WithdrawSum string `json:"sum"`
}

func NewOrders(str Storage) Orders {
	return Orders{storage: str}
}

func (o *Orders) AddOrder(ctx context.Context, orderNumber string, userID uint64) error {
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

func (o *Orders) GetOrders(ctx context.Context, userID uint64) (*[]Order, error) {
	orders, err := o.storage.GetOrdersList(ctx, userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return orders, nil
}

func (o *Orders) GetBalance(ctx context.Context, userID uint64) (int64, int64, error) {
	bal, wtn, err := o.storage.GetBalance(ctx, userID)
	if err != nil {
		log.Println(err)
		return 0, 0, err
	}

	return bal, wtn, nil
}

func (o *Orders) Withdraw(ctx context.Context, request WithdrawRequest) error {

	number, err := helpers.ParseOrderNumber(request.OrderNumber)
	if err != nil {
		log.Println(err)
		return ErrInvalidOrderNumber
	}

	bal, wtn, err := o.GetBalance(ctx, request.UserID)
	if err != nil {
		return err
	}

	sum, err := strconv.ParseInt(request.WithdrawSum, 10, 64)
	if sum > bal {
		return ErrInsufficientFunds
	}

	err = o.storage.AddWithdrawal(ctx, request.UserID, number, sum)
	if err != nil {
		return err
	}

	bal -= sum
	wtn += sum

	err = o.storage.UpdateBalance(ctx, request.UserID, bal, wtn)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orders) GetWithdrawals(ctx context.Context, userID uint64) (*[]Withdrawal, error) {
	withdrawals, err := o.storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return withdrawals, nil
}
