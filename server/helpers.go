package server

import (
	"errors"
	"net/http"

	"github.com/moorzeen/loyalty-service/auth"
)

func errToStatus(err error) int {
	switch {
	case errors.Is(err, auth.ErrShortPassword):
		return http.StatusBadRequest
	case errors.Is(err, auth.ErrUsernameTaken):
		return http.StatusConflict
	case errors.Is(err, auth.ErrInvalidUser) ||
		errors.Is(err, auth.ErrInvalidAuthToken) ||
		errors.Is(err, auth.ErrNoUser) ||
		errors.Is(err, auth.ErrWrongPassword):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
