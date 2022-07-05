package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/moorzeen/loyalty-service/services/auth"
	"github.com/moorzeen/loyalty-service/services/order"
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

	err = s.AuthService.SignUp(r.Context(), cred.Username, cred.Password)
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

	authToken, err := s.AuthService.SignIn(r.Context(), cred.Username, cred.Password)
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

func (s *LoyaltyServer) NewOrder(w http.ResponseWriter, r *http.Request) {
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
	err = s.OrderService.AddOrder(r.Context(), string(orderNumber), userID)
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

	ordersList, err := s.OrderService.GetOrders(r.Context(), userID)
	if err != nil {
		msg := fmt.Sprintf("Failed to get order: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	type responseJSON struct {
		Number     string    `json:"number"`
		Status     string    `json:"status"`
		Accrual    float64   `json:"accrual,omitempty"`
		UploadedAt time.Time `json:"uploaded_at"`
	}
	result := make([]responseJSON, 0)
	for _, v := range *ordersList {
		newItem := responseJSON{strconv.FormatInt(v.OrderNumber, 10), v.Status, v.Accrual, v.UploadedAt}
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

func (s *LoyaltyServer) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())

	bal, wtn, err := s.OrderService.GetBalance(r.Context(), userID)
	if err != nil {
		msg := fmt.Sprintf("Failed to get balance: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	type responseJSON struct {
		Balance   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}

	result := responseJSON{
		Balance:   bal,
		Withdrawn: wtn,
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

func (s *LoyaltyServer) Withdraw(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf("Unsupported content type \"%s\"", contentType)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	wr := order.WithdrawRequest{}

	err := json.NewDecoder(r.Body).Decode(&wr)
	if err != nil {
		msg := fmt.Sprintf("Failed to parse withdraw data: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	wr.UserID = GetUserID(r.Context())

	err = s.OrderService.Withdraw(r.Context(), wr)
	if err != nil {
		msg := fmt.Sprintf("Failed to withdraw: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *LoyaltyServer) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())

	withdrawalsList, err := s.OrderService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		msg := fmt.Sprintf("Failed to get withdrawals: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	type responseJSON struct {
		Number     string    `json:"number"`
		Sum        float64   `json:"sum"`
		UploadedAt time.Time `json:"processed_at"`
	}
	result := make([]responseJSON, 0)

	for _, v := range *withdrawalsList {
		item := responseJSON{strconv.FormatInt(v.OrderNumber, 10), v.Sum, v.ProcessedAt}
		result = append(result, item)
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
