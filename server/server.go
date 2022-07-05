package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/moorzeen/loyalty-service/services/accrual"
	accrualPG "github.com/moorzeen/loyalty-service/services/accrual/storage/postgres"
	"github.com/moorzeen/loyalty-service/services/auth"
	authPG "github.com/moorzeen/loyalty-service/services/auth/storage/postgres"
	"github.com/moorzeen/loyalty-service/services/order"
	ordersPG "github.com/moorzeen/loyalty-service/services/order/storage/postgres"
)

type LoyaltyServer struct {
	Config
	AuthService    *auth.Service
	OrderService   *order.Service
	AccrualService *accrual.Service
	Router         *chi.Mux
}

func NewServer(cfg *Config) (*LoyaltyServer, error) {
	db, err := initDB(cfg.DatabaseURI)
	if err != nil {
		return nil, err
	}

	srv := &LoyaltyServer{}
	srv.Config = *cfg

	authStorage := authPG.NewStorage(db)
	srv.AuthService = auth.NewService(authStorage)

	ordersStorage := ordersPG.NewStorage(db)
	srv.OrderService = order.NewService(ordersStorage)

	accrualStorage := accrualPG.NewStorage(db)
	client := accrual.NewClient(srv.RunAddress)
	srv.AccrualService = accrual.NewService(accrualStorage, client)

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

func initDB(databaseURI string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, databaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new connection pool: %w", err)
	}

	return pool, nil
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
