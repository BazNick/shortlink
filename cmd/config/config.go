package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Address string `env:"ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func GetCLParams() Config {
	var config Config

	err := env.Parse(&config)

	if err != nil {
		log.Fatal(err)
	}

	if config.BaseURL != "" && config.Address != "" {
		return config
	}

	flag.StringVar(&config.Address, "a", "localhost:8080", "Адрес запуска HTTP-сервера")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base URL")

	flag.Parse()

	return config
}
