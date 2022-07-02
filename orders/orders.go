package orders

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/moorzeen/loyalty-service/orders/helpers"
)

type Orders struct {
	storage Storage
}

func NewOrders(str Storage) Orders {
	return Orders{storage: str}
}

func (o *Orders) AddOrder(ctx context.Context, orderNumber string, userID uint64) error {
	number, err := helpers.ParseOrderNumber(orderNumber)
	if err != nil {
		log.Println(err)
		return err
	}

	err = o.storage.AddOrder(ctx, number, userID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		fmt.Println(err)
		order, err := o.storage.GetOrderByNumber(ctx, number)
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
