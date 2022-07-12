package server

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type config struct {
	RunAddress     string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI    string `env:"DATABASE_URI" envDefault:"postgresql://localhost:5432/postgres?sslmode=disable"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8080"`
}

func GetConfig() (*config, error) {
	config := &config{}

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	flag.StringVar(&config.RunAddress, "a", config.RunAddress, "server address and port")
	flag.StringVar(&config.DatabaseURI, "d", config.DatabaseURI, "database URI")
	flag.StringVar(&config.AccrualAddress, "r", config.AccrualAddress, "accrual system address")
	flag.Parse()

	return config, nil
}
