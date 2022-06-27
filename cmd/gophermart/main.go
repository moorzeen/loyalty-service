package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/moorzeen/loyalty-service/internal/server"
	"github.com/moorzeen/loyalty-service/internal/storage"
)

func main() {

	cfg, err := server.NewConfig()
	if err != nil {
		log.Fatalf("Failed to get server configuration: %s", err)
	}

	log.Printf("Starting with config:\n"+
		"- run address: %s\n"+
		"- database URI: %s\n"+
		"- accrual system address: %s\n", cfg.RunAddress, cfg.DatabaseURI, cfg.AccrualSystemAddress)

	err = storage.Migrate(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Failed to migrate DB: %s", err)
	}

	_, err = server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to start the server: %s", err)
	}

	log.Println("Server is listening and serving...")

	awaitStop()
}

func awaitStop() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
}
