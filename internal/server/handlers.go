package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/moorzeen/loyalty-service/internal/services/auth"
	"github.com/moorzeen/loyalty-service/internal/services/storage"
)

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var c userCredentials
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = auth.PassComplexity(c.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash := auth.GenerateHash(c.Password)

	err = s.Storage.AddUser(c.Login, hash)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), errToStatus(err))
		return
	}

	http.Redirect(w, r, "/api/user/login", http.StatusTemporaryRedirect)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var c userCredentials
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := storage.User{}
	user, err = s.Storage.GetUser(c.Login)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), errToStatus(err))
		return
	}

	hash := auth.GenerateHash(c.Password)
	if user.PasswordHash != hash {
		log.Println(storage.ErrInvalidUser)
		http.Error(w, storage.ErrInvalidUser.Error(), http.StatusUnauthorized)
		return
	}

	user.SessionUUID = uuid.New()

	err = s.Storage.SetSession(user.Login, user.SessionUUID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authCookie := http.Cookie{Name: "authToken", Value: user.SessionUUID.String()}
	http.SetCookie(w, &authCookie)

	w.WriteHeader(http.StatusOK)
}
