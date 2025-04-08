package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Address  string `env:"ADDRESS"`
	BaseURL  string `env:"BASE_URL"`
	FilePath string `env:"FILE_STORAGE_PATH"`
	DB       string `env:"DATABASE_DSN"`
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

	if config.FilePath == "" {
		flag.StringVar(&config.FilePath, "f", "data.json", "path to file")
	}
	
	if config.DB == "" {
		flag.StringVar(&config.DB, "d", "postgres://user:password@localhost:5432/dbname", "db connection settings")
	}

	flag.StringVar(&config.Address, "a", "localhost:8080", "http server adress")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base URL")

	flag.Parse()

	return config
}
