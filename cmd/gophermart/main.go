package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/moorzeen/loyalty-service/api"
)

func main() {
	cfg, err := api.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get server configuration: %s", err)
	}

	log.Printf("Strarting server with configuration:\n"+
		"- run address: %s\n"+
		"- database URI: %s\n"+
		"- accrual system address: %s\n", cfg.RunAddress, cfg.DatabaseURI, cfg.AccrualSystemAddress)

	_, err = api.StartServer(cfg)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
	log.Println("Server is listening and serving...")

	awaitStop()
	log.Println("Server stopped.")
}

func awaitStop() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
}
