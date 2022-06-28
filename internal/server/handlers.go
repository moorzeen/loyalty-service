package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/moorzeen/loyalty-service/internal/services/auth"
)

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	u := user{}

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = auth.PassComplexity(u.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash := auth.GenerateHash(u.Password)

	err = s.Storage.AddUser(u.Login, hash)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), errToStatus(err))
		return
	}

	http.Redirect(w, r, "/api/user/login", http.StatusTemporaryRedirect)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	u := user{}

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash := auth.GenerateHash(u.Password)

	err = s.Storage.IsUser(u.Login, hash)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), errToStatus(err))
		return
	}

	v := fmt.Sprintf("%s%s", u.Login, hash)
	authCookie := http.Cookie{Name: "authToken", Value: v}
	http.SetCookie(w, &authCookie)

	w.WriteHeader(http.StatusOK)
}
