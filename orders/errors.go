package orders

import (
	"errors"
)

var (
	ErrAlreadyAddByThis   = errors.New("already added by you")
	ErrAddedByOther       = errors.New("already added by other")
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)
