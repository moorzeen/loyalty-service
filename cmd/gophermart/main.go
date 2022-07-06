package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/moorzeen/loyalty-service/internal/server"
	"github.com/moorzeen/loyalty-service/internal/storage/postgres"
)

func main() {
	cfg, err := server.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get configuration: %s", err)
	}

	log.Printf(
		"Configuration:\n- run address: %s\n- database URI: %s\n- accrual system address: %s\n",
		cfg.RunAddress, cfg.DatabaseURI, cfg.AccrualSystemAddress)

	err = postgres.Migration(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to migrate DB: %s", err)
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to init the server: %s", err)
	}

	srv.Run()

	log.Println("Server is listening and serving...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)
	<-quit

	log.Println("Server stopped")
}
