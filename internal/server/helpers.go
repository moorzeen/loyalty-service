package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/moorzeen/loyalty-service/internal/auth"
	"github.com/moorzeen/loyalty-service/internal/order"
)

func getUserID(ctx context.Context) uint64 {
	return ctx.Value(UserIDContextKey).(uint64)
}

func errToStatus(err error) int {
	switch {
	case
		errors.Is(err, auth.ErrShortPassword):
		return http.StatusBadRequest
	case
		errors.Is(err, auth.ErrUsernameTaken) ||
			errors.Is(err, order.ErrAddedByOther):
		return http.StatusConflict
	case
		errors.Is(err, auth.ErrInvalidUser) ||
			errors.Is(err, auth.ErrInvalidAuthToken) ||
			errors.Is(err, auth.ErrNoUser) ||
			errors.Is(err, auth.ErrWrongPassword):
		return http.StatusUnauthorized
	case
		errors.Is(err, order.ErrInvalidOrderNumber):
		return http.StatusUnprocessableEntity
	case
		errors.Is(err, order.ErrAlreadyAddByThis):
		return http.StatusOK
	case
		errors.Is(err, order.ErrInsufficientFunds):
		return http.StatusPaymentRequired
	default:
		return http.StatusInternalServerError
	}
}
