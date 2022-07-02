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
	"github.com/moorzeen/loyalty-service/auth"
	"github.com/moorzeen/loyalty-service/auth/storage/postgres"
)

type LoyaltyServer struct {
	Config
	Auth   auth.Service
	Router *chi.Mux
}

func New(cfg *Config) (*LoyaltyServer, error) {
	db, err := initDB(cfg.DatabaseURI)
	if err != nil {
		return nil, err
	}
	authStorage := postgres.NewStorage(db)

	srv := &LoyaltyServer{}
	srv.Config = *cfg
	srv.Auth = auth.NewAuth(authStorage)
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
		r.Use(Authenticator(srv.Auth))
		r.Post("/api/user/orders", srv.PostOrder)
	})
	return r
}