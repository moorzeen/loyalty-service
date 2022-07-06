package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/moorzeen/loyalty-service/internal/accrual"
	"github.com/moorzeen/loyalty-service/internal/auth"
	"github.com/moorzeen/loyalty-service/internal/order"
	"github.com/moorzeen/loyalty-service/internal/storage/postgres"
)

type LoyaltyServer struct {
	Config
	AuthService    *auth.Service
	OrderService   *order.Service
	AccrualService *accrual.Service
	Router         *chi.Mux
}

func NewServer(cfg *Config) (*LoyaltyServer, error) {

	storage, err := postgres.NewStorage(cfg.DatabaseURI)
	if err != nil {

	}

	srv := &LoyaltyServer{}
	srv.Config = *cfg

	srv.AuthService = auth.NewService(storage)
	srv.OrderService = order.NewService(storage)

	client := accrual.NewClient(srv.AccrualSystemAddress)
	srv.AccrualService = accrual.NewService(storage, client)

	srv.Router = newRouter(srv)

	return srv, nil
}

func (s *LoyaltyServer) Run() {
	go func() {
		err := http.ListenAndServe(s.RunAddress, s.Router)
		if err != nil {
			log.Printf("Server failed: %s", err)
		}
	}()
}

func newRouter(srv *LoyaltyServer) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(RequestDecompress)
	r.Use(middleware.Compress(5))
	r.Post("/api/user/register", srv.Register)
	r.Post("/api/user/login", srv.Login)

	// authorization required handlers
	r.Group(func(r chi.Router) {
		//r.Use(Authentication)
		r.Use(Authenticator(*srv.AuthService))
		r.Post("/api/user/orders", srv.NewOrder)
		r.Get("/api/user/orders", srv.GetOrders)
		r.Get("/api/user/balance", srv.GetBalance)
		r.Post("/api/user/balance/withdraw", srv.Withdraw)
		r.Get("/api/user/withdrawals", srv.GetWithdrawals)

	})
	return r
}
