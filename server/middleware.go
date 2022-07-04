package server

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/moorzeen/loyalty-service/auth"
)

type ctxKey string

const UserIDContextKey ctxKey = "userID"

func RequestDecompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			var err error
			r.Body, err = gzip.NewReader(r.Body)
			if err != nil {
				msg := fmt.Sprintf("Failed to decompress request: %s", err)
				log.Println(msg)
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

/* panic on "err = a.AuthService.TokenCheck(r.Context(), cookie.Value)"
func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie(auth.UserAuthCookieName)
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if err != nil {
			log.Println("cookie parse error:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var a *requestAuth
		err = a.AuthService.TokenCheck(r.Context(), cookie.Value)
		//err = a.TokenCheck(r.Context(), cookie.Value)
		if err != nil {
			log.Println(err)
			w.WriteHeader(errToStatus(err))
			return
		}

		ctx := context.WithValue(r.Context(), "keyName", "value")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
*/

func Authenticator(s auth.Auth) func(http.Handler) http.Handler {
	ra := requestAuth{s}
	return func(next http.Handler) http.Handler {
		serveHTTP := func(w http.ResponseWriter, r *http.Request) {
			userID, err := ra.validateCookie(r)
			if err != nil {
				log.Println(err)
				http.Error(w, "Login to access this endpoint", http.StatusUnauthorized)
				return
			}

			newContext := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(newContext))
		}
		return http.HandlerFunc(serveHTTP)
	}
}

type requestAuth struct {
	AuthService auth.Auth
}

func (a *requestAuth) validateCookie(r *http.Request) (uint64, error) {
	cookie, err := r.Cookie(auth.UserAuthCookieName)
	if err == http.ErrNoCookie {
		msg := fmt.Sprintf("cookie is not found: %s", err)
		return 0, errors.New(msg)
	}
	if err != nil {
		msg := fmt.Sprintf("cookie parse error: %s", err)
		return 0, errors.New(msg)
	}

	userID, err := a.AuthService.TokenCheck(r.Context(), cookie.Value)
	if err != nil {
		msg := fmt.Sprintf("%s", err)
		return 0, errors.New(msg)
	}

	return userID, nil
}
