package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/moorzeen/loyalty-service/internal/accrual"
	"github.com/moorzeen/loyalty-service/internal/auth"
	"github.com/moorzeen/loyalty-service/internal/order"
	"github.com/moorzeen/loyalty-service/internal/storage"
	"github.com/moorzeen/loyalty-service/internal/storage/postgres"
)

type LoyaltyServer struct {
	config
	storage storage.Service
	auth    auth.Service
	order   order.Service
	accrual accrual.Service
	Router  *chi.Mux
}

func NewServer(cfg *config) (*LoyaltyServer, error) {
	ls := &LoyaltyServer{}
	ls.config = *cfg

	var err error
	ls.storage, err = postgres.NewStorage(cfg.DatabaseURI)
	if err != nil {
		return nil, err
	}

	ls.auth = auth.NewService(ls.storage)
	ls.order = order.NewService(ls.storage)
	client := accrual.NewClient(ls.AccrualAddress)
	ls.accrual = accrual.NewService(ls.storage, client)
	ls.Router = newRouter(ls)

	return ls, nil
}

func (ls *LoyaltyServer) Run() {
	go func() {
		err := http.ListenAndServe(ls.RunAddress, ls.Router)
		if err != nil {
			log.Fatalf("Server error: %s", err)
		}
	}()
}

func newRouter(ls *LoyaltyServer) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(RequestDecompress)
	r.Use(middleware.Compress(5))
	r.Post("/api/user/register", ls.register)
	r.Post("/api/user/login", ls.login)

	// authorization required handlers
	r.Group(func(r chi.Router) {
		r.Use(Authentication(ls.auth))
		r.Post("/api/user/orders", ls.newOrder)
		r.Get("/api/user/orders", ls.getOrders)
		r.Get("/api/user/balance", ls.getBalance)
		r.Post("/api/user/balance/withdraw", ls.withdraw)
		r.Get("/api/user/withdrawals", ls.getWithdrawals)

	})
	return r
}
