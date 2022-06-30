package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/moorzeen/loyalty-service/internal/services/auth"
)

func (s *LoyaltyServer) register(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf("Unsupported content type \"%s\"", contentType)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	cred := credentials{}

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		msg := fmt.Sprintf("Failed to parse login/password: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if cred.Login == "" || cred.Password == "" {
		msg := fmt.Sprintf("Empty login/password not allowed")
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	err = s.Auth.Register()

	if err != nil {
		msg := fmt.Sprintf("Failed to register: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	http.Redirect(w, r, "/api/user/login", http.StatusTemporaryRedirect)
}

func (s *LoyaltyServer) login(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf("Unsupported content type \"%s\"", contentType)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	cred := credentials{}

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		msg := fmt.Sprintf("Failed to parse login/password: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	signedCred := auth.Credentials{Login: cred.Login, Password: cred.Password}

	signedUser, err := s.Auth.Login(signedCred)

	if err != nil {
		msg := fmt.Sprintf("Failed to login: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	authCookie := auth.MakeAuthCookie(signedUser)
	http.SetCookie(w, &authCookie)

	w.WriteHeader(http.StatusOK)
}

func (s *LoyaltyServer) postOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(body)

}
