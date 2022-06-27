package server

import (
	"errors"
	"net/http"

	"github.com/moorzeen/loyalty-service/internal/storage"
)

func passComplexity(pass string) error {
	if len([]rune(pass)) < 8 {
		return errors.New("the password is too short, requires more than 7 characters")
	}
	return nil
}

func errToStatus(err error) int {
	switch {
	case errors.Is(err, storage.ErrLoginTaken):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
