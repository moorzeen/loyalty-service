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
		"Starting configuration:\n- run address: %s\n- database URI: %s\n- accrual system address: %s\n",
		cfg.RunAddress, cfg.DatabaseURI, cfg.AccrualAddress)

	err = postgres.Migration(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to migrate DB: %s", err)
	}

	ls, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to init the server: %s", err)
	}

	ls.Run()
	log.Println("Server is listening and serving...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)
	<-quit

	log.Println("Server stopped")
}
