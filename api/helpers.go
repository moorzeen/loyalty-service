package api

import (
	"errors"
	"net/http"

	"github.com/moorzeen/loyalty-service/internal/services/auth"
	"github.com/moorzeen/loyalty-service/storage"
)

func errToStatus(err error) int {
	switch {
	case errors.Is(err, storage.ErrLoginTaken):
		return http.StatusConflict
	case errors.Is(err, storage.ErrInvalidUser) || errors.Is(err, auth.ErrInvalidSession) || errors.Is(err, auth.ErrWrongUser):
		return http.StatusUnauthorized
	case errors.Is(err, auth.ErrShortPassword):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
