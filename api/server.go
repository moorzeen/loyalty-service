package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/moorzeen/loyalty-service/internal/services/auth"
	"github.com/moorzeen/loyalty-service/storage"
	"github.com/moorzeen/loyalty-service/storage/postgres"
)

type LoyaltyServer struct {
	Config
	Storage storage.Service
	Auth    auth.Service
	Router  *chi.Mux
}

func StartServer(cfg *Config) (*LoyaltyServer, error) {
	srv := &LoyaltyServer{}
	srv.Config = *cfg

	err := postgres.Migrate(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate DB: %w", err)
	}

	srv.Storage, err = postgres.Open(context.Background(), srv.DatabaseURI)
	if err != nil {
		return nil, err
	}

	srv.Auth = auth.NewService(srv.Storage)
	srv.Router = NewRouter(srv)

	go func() {
		err = http.ListenAndServe(srv.RunAddress, srv.Router)
		if err != nil {
			log.Printf("Server failed: %s", err)
		}
	}()

	return srv, nil
}

func NewRouter(srv *LoyaltyServer) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(requestDecompress)
	r.Use(middleware.Compress(5))
	r.Post("/api/user/register", srv.register)
	r.Post("/api/user/login", srv.login)

	// authorization required handlers
	r.Group(func(r chi.Router) {
		r.Use(authentication)
		r.Post("/api/user/orders", srv.postOrder)
	})
	return r
}
