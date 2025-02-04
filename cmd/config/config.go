package config

import "flag"

type Config struct {
	Address string
	BaseURL string
}

func GetCLParams() Config {
	var config Config

	flag.StringVar(&config.Address, "a", "localhost:8080", "Адрес запуска HTTP-сервера")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base URL")

	flag.Parse()
	
	return config
}
