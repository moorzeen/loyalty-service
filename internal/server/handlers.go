package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/moorzeen/loyalty-service/internal/auth"
	"github.com/moorzeen/loyalty-service/internal/order"
)

func (ls *LoyaltyServer) register(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf("Unsupported content type \"%s\"", contentType)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	cred := auth.Credentials{}

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		msg := fmt.Sprintf("Failed to parse login or password: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if cred.Username == "" || cred.Password == "" {
		msg := "Empty login or password is not allowed"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	err = ls.auth.SignUp(r.Context(), cred)
	if err != nil {
		msg := fmt.Sprintf("Can't regitser: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	http.Redirect(w, r, "/api/user/login", http.StatusTemporaryRedirect)
}

func (ls *LoyaltyServer) login(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf("Unsupported content type \"%s\"", contentType)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	cred := auth.Credentials{}

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		msg := fmt.Sprintf("Failed to parse login or password: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if cred.Username == "" || cred.Password == "" {
		msg := "Empty login or password"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	authToken, err := ls.auth.SignIn(r.Context(), cred)
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

func (ls *LoyaltyServer) newOrder(w http.ResponseWriter, r *http.Request) {
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

	userID := getUserID(r.Context())

	err = ls.order.AddOrder(r.Context(), string(orderNumber), userID)
	if err != nil {
		msg := fmt.Sprintf("Failed to add the order: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (ls *LoyaltyServer) getOrders(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	orders, err := ls.order.GetOrders(r.Context(), userID)
	if err != nil {
		msg := fmt.Sprintf("Failed to get order: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	result := make([]order.Order, 0)

	for _, v := range orders {
		item := order.Order{
			Number:     v.OrderNumber,
			Status:     v.Status,
			Accrual:    v.Accrual,
			UploadedAt: v.UploadedAt,
		}
		result = append(result, item)
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

func (ls *LoyaltyServer) getBalance(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	bal, wtn, err := ls.order.GetBalance(r.Context(), userID)
	if err != nil {
		msg := fmt.Sprintf("Failed to get balance: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	result := order.Balance{
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

func (ls *LoyaltyServer) withdraw(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		msg := fmt.Sprintf("Unsupported content type \"%s\"", contentType)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	wr := order.Withdraw{}

	err := json.NewDecoder(r.Body).Decode(&wr)
	if err != nil {
		msg := fmt.Sprintf("Failed to parse withdraw data: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	wr.UserID = getUserID(r.Context())

	err = ls.order.Withdraw(r.Context(), wr)
	if err != nil {
		msg := fmt.Sprintf("Failed to withdraw: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (ls *LoyaltyServer) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	withdrawals, err := ls.order.GetWithdrawals(r.Context(), userID)
	if err != nil {
		msg := fmt.Sprintf("Failed to get withdrawals: %s", err)
		log.Println(msg)
		http.Error(w, msg, errToStatus(err))
		return
	}

	type responseJSON struct {
		Number     string    `json:"order"`
		Sum        float64   `json:"sum"`
		UploadedAt time.Time `json:"processed_at"`
	}
	result := make([]responseJSON, 0)

	for _, v := range withdrawals {
		item := responseJSON{v.OrderNumber, v.Sum, v.ProcessedAt}
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
