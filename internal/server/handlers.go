package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	u := user{}

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = passComplexity(u.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash := generateHash(u.Password, secret)

	err = s.Storage.Register(u.Login, hash)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), errToStatus(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}
