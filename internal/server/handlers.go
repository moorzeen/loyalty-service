package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var c userCredentials
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.Auth.Register(c.Login, c.Password)
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

	authCookie, err := s.Auth.Login(c.Login, c.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), errToStatus(err))
		return
	}

	http.SetCookie(w, &authCookie)
	w.WriteHeader(http.StatusOK)
}
