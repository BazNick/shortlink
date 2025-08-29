package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

// Config хранит настройки конфигурации приложения.
type Config struct {
	Address   string `env:"ADDRESS"`           // Адрес HTTP-сервера
	BaseURL   string `env:"BASE_URL"`          // Базовый URL сервиса
	FilePath  string `env:"FILE_STORAGE_PATH"` // Путь к файлам хранилища
	DB        string `env:"DATABASE_DSN"`      // Строка подключения к базе данных
	SecretKey string `env:"SECRET_KEY"`        // Секретный ключ для JWT-токенов
}

// GetCLParams считывает конфигурационные параметры из переменных среды и аргументов командной строки.
// Возвращает заполненную структуру Config с настройками.
//
// Поддерживаемые переменные среды:
//   - ADDRESS: адрес HTTP-сервера
//   - BASE_URL: базовый URL сервиса
//   - FILE_STORAGE_PATH: путь к файлам хранилища
//   - DATABASE_DSN: строка подключения к базе данных
//   - SECRET_KEY: секретный ключ для JWT-токенов
//
// Поддерживаемые аргументы командной строки:
//   - -a: адрес HTTP-сервера (по умолчанию: localhost:8080)
//   - -b: базовый URL (по умолчанию: http://localhost:8080)
//   - -f: путь к файлам хранилища
//   - -d: настройка подключения к базе данных
//   - -k: секретный ключ для JWT-токенов
//
// Возвращает:
//   - Config: структуру конфигурации с установленными параметрами
//   - error: ошибку, если не удалось считать параметры
//
// Пример:
//
//	config, err := GetCLParams()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Running server on address: %s\n", config.Address)
func GetCLParams() (Config, error) {
	var config Config

	err := env.Parse(&config)

	if err != nil {
		return Config{}, err
	}

	if config.BaseURL != "" && config.Address != "" {
		return config, nil
	}

	if config.FilePath == "" {
		flag.StringVar(&config.FilePath, "f", "", "path to file")
	}

	if config.DB == "" {
		flag.StringVar(&config.DB, "d", "", "db connection settings") // postgres://user:password@localhost:5432/dbname
	}

	if config.SecretKey == "" {
		flag.StringVar(&config.SecretKey, "k", "", "secret key for jwt token")
	}

	flag.StringVar(&config.Address, "a", "localhost:8080", "http server adress")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "base URL")

	flag.Parse()

	return config, nil
}
