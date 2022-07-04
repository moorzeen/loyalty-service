package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/moorzeen/loyalty-service/auth"
)

type credentials struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

func (s *LoyaltyServer) Register(w http.ResponseWriter, r *http.Request) {
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

	if cred.Username == "" || cred.Password == "" {
		msg := "Empty login/password not allowed"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	err = s.Auth.SignUp(r.Context(), cred.Username, cred.Password)
	if err != nil {
		msg := fmt.Sprintf("Can't regitser: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	http.Redirect(w, r, "/api/user/login", http.StatusTemporaryRedirect)
}

func (s *LoyaltyServer) Login(w http.ResponseWriter, r *http.Request) {
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

	if cred.Username == "" || cred.Password == "" {
		msg := "Empty login/password not allowed"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	authToken, err := s.Auth.SignIn(r.Context(), cred.Username, cred.Password)
	if err != nil {
		msg := fmt.Sprintf("Can not login: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	authCookie := http.Cookie{
		Value: authToken,
		Name:  auth.UserAuthCookieName,
	}
	http.SetCookie(w, &authCookie)
	w.WriteHeader(http.StatusOK)
}

func (s *LoyaltyServer) PostOrder(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		msg := fmt.Sprintf("Unsupported content type \"%s\"", contentType)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	orderNumber, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("Filed to read request body: %s", err)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	userID := GetUserID(r.Context())
	err = s.Orders.AddOrder(r.Context(), string(orderNumber), userID)
	if err != nil {
		msg := fmt.Sprintf("Failed to add the order: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *LoyaltyServer) GetOrders(w http.ResponseWriter, r *http.Request) {

	userID := GetUserID(r.Context())

	orders, err := s.Orders.GetOrders(r.Context(), userID)
	if err != nil {
		msg := fmt.Sprintf("Failed to get orders: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	type responseJSON struct {
		Number     int64     `json:"number"`
		Status     string    `json:"status"`
		Accrual    int64     `json:"accrual,omitempty"`
		UploadedAt time.Time `json:"uploaded_at"`
	}
	result := make([]responseJSON, 0)
	for _, v := range orders {
		newItem := responseJSON{v.OrderNumber, v.Status, v.Accrual, v.UploadedAt}
		result = append(result, newItem)
	}
	if len(result) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&result)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
