package api

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/moorzeen/loyalty-service/internal/services/auth"
)

type ctxKey string

const authToken ctxKey = "authToken"

func requestDecompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			var err error
			r.Body, err = gzip.NewReader(r.Body)
			if err != nil {
				log.Println(err)
				fmt.Errorf("failed to decompress request: %w", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("authToken")
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var a auth.Service
		err = a.Validate(cookie.Value)
		if err != nil {
			w.WriteHeader(errToStatus(err))
			return
		}

		var login string
		ctx := context.WithValue(r.Context(), authToken, login)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
