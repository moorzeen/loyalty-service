package server

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/moorzeen/loyalty-service/internal/storage"
)

type Server struct {
	Config
	Router  *chi.Mux
	Storage storage.Storage
}

func NewServer(c *Config) (*Server, error) {
	srv := &Server{}
	srv.Config = *c
	var err error

	srv.Storage, err = storage.NewConnection(context.Background(), srv.DatabaseURI)
	if err != nil {
		return nil, err
	}

	srv.Router = NewRouter(srv)

	go func() {
		err = http.ListenAndServe(srv.RunAddress, srv.Router)
		if err != nil {
			log.Printf("Server failed: %s", err)
		}
	}()

	return srv, nil
}

func NewRouter(srv *Server) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	//r.Use(RequestDecompress)
	//r.Use(Authentication)
	r.Use(middleware.Compress(5))
	r.Post("/api/user/register", srv.register)
	//r.Post("/api/shorten", srv.ShortenJSON)
	//r.Post("/api/shorten/batch", srv.ShortenBatch)
	//r.Get("/{key}", srv.Expand)
	//r.Get("/api/user/urls", srv.UserURLs)
	//r.Get("/ping", srv.PingDB)
	//r.Delete("/api/user/urls", srv.DeleteBatch)
	return r
}
