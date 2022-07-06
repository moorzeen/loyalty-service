package server

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgresql://localhost:5432/postgres?sslmode=disable"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8080"`
}

func GetConfig() (*Config, error) {
	config := &Config{}

	err := env.Parse(config)
	if err != nil {
		return nil, err
	}

	flag.StringVar(&config.RunAddress, "a", config.RunAddress, "api address and port")
	flag.StringVar(&config.DatabaseURI, "d", config.DatabaseURI, "database URI")
	flag.StringVar(&config.AccrualSystemAddress, "r", config.AccrualSystemAddress, "accrual system address")
	flag.Parse()

	return config, nil
}
