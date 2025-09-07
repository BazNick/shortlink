package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

// JSONConfig представляет структуру конфигурации в JSON файле.
type JSONConfig struct {
	ServerAddress   string `json:"server_address"`    // Адрес HTTP-сервера
	BaseURL         string `json:"base_url"`          // Базовый URL сервиса
	FileStoragePath string `json:"file_storage_path"` // Путь к файлам хранилища
	DatabaseDSN     string `json:"database_dsn"`      // Строка подключения к базе данных
	EnableHTTPS     bool   `json:"enable_https"`      // Включить HTTPS сервер
}

// Config хранит настройки конфигурации приложения.
type Config struct {
	Address     string `env:"ADDRESS" json:"server_address"`              // Адрес HTTP-сервера
	BaseURL     string `env:"BASE_URL" json:"base_url"`                   // Базовый URL сервиса
	FilePath    string `env:"FILE_STORAGE_PATH" json:"file_storage_path"` // Путь к файлам хранилища
	DB          string `env:"DATABASE_DSN" json:"database_dsn"`           // Строка подключения к базе данных
	SecretKey   string `env:"SECRET_KEY" json:"secret_key"`               // Секретный ключ для JWT-токенов
	EnableHTTPS bool   `env:"ENABLE_HTTPS" json:"enable_https"`           // Включить HTTPS сервер
	CertFile    string `env:"CERT_FILE" json:"cert_file"`                 // Путь к файлу сертификата
	KeyFile     string `env:"KEY_FILE" json:"key_file"`                   // Путь к файлу приватного ключа
}

// loadJSONConfig загружает конфигурацию из JSON файла.
// Возвращает заполненную структуру JSONConfig или ошибку, если файл не найден или невалиден.
//
// Параметры:
//   - configPath: путь к JSON файлу конфигурации
//
// Возвращает:
//   - JSONConfig: структуру конфигурации из JSON файла
//   - error: ошибку, если не удалось загрузить или распарсить файл
//
// Пример:
//
//	jsonConfig, err := loadJSONConfig("config.json")
//	if err != nil {
//	    log.Printf("Warning: could not load JSON config: %v", err)
//	}
func loadJSONConfig(configPath string) (JSONConfig, error) {
	var jsonConfig JSONConfig

	file, err := os.Open(configPath)
	if err != nil {
		return JSONConfig{}, fmt.Errorf("failed to open config file %s: %w", configPath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&jsonConfig); err != nil {
		return JSONConfig{}, fmt.Errorf("failed to decode JSON config from %s: %w", configPath, err)
	}

	return jsonConfig, nil
}

// applyJSONConfig применяет значения из JSON конфигурации к Config структуре.
// Значения применяются только если соответствующие поля в Config пустые.
//
// Параметры:
//   - config: указатель на структуру Config для обновления
//   - jsonConfig: структура JSONConfig с значениями из файла
func applyJSONConfig(config *Config, jsonConfig JSONConfig) {
	if config.Address == "" && jsonConfig.ServerAddress != "" {
		config.Address = jsonConfig.ServerAddress
	}
	if config.BaseURL == "" && jsonConfig.BaseURL != "" {
		config.BaseURL = jsonConfig.BaseURL
	}
	if config.FilePath == "" && jsonConfig.FileStoragePath != "" {
		config.FilePath = jsonConfig.FileStoragePath
	}
	if config.DB == "" && jsonConfig.DatabaseDSN != "" {
		config.DB = jsonConfig.DatabaseDSN
	}
	// Для bool значений применяем только если значение в JSON не false
	if !config.EnableHTTPS && jsonConfig.EnableHTTPS {
		config.EnableHTTPS = jsonConfig.EnableHTTPS
	}
}

// GetCLParams считывает конфигурационные параметры из JSON файла, переменных среды и аргументов командной строки.
// Приоритет: аргументы командной строки > переменные среды > JSON файл > значения по умолчанию.
// Возвращает заполненную структуру Config с настройками.
//
// Поддерживаемые переменные среды:
//   - CONFIG: путь к JSON файлу конфигурации
//   - ADDRESS: адрес HTTP-сервера
//   - BASE_URL: базовый URL сервиса
//   - FILE_STORAGE_PATH: путь к файлам хранилища
//   - DATABASE_DSN: строка подключения к базе данных
//   - SECRET_KEY: секретный ключ для JWT-токенов
//   - ENABLE_HTTPS: включить HTTPS сервер (true/false)
//   - CERT_FILE: путь к файлу сертификата
//   - KEY_FILE: путь к файлу приватного ключа
//
// Поддерживаемые аргументы командной строки:
//   - -c/-config: путь к JSON файлу конфигурации
//   - -a: адрес HTTP-сервера (по умолчанию: localhost:8080)
//   - -b: базовый URL (по умолчанию: http://localhost:8080)
//   - -f: путь к файлам хранилища
//   - -d: настройка подключения к базе данных
//   - -k: секретный ключ для JWT-токенов
//   - -s: включить HTTPS сервер
//   - -cert: путь к файлу сертификата
//   - -key: путь к файлу приватного ключа
//
// Формат JSON файла конфигурации:
//
//	{
//	  "server_address": "localhost:8080",
//	  "base_url": "http://localhost",
//	  "file_storage_path": "/path/to/file.db",
//	  "database_dsn": "",
//	  "enable_https": true
//	}
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
	var configFile string

	// Сначала определяем путь к конфигурационному файлу из переменной среды
	configFile = os.Getenv("CONFIG")

	// Определяем флаги командной строки
	flag.StringVar(&configFile, "c", configFile, "path to JSON config file")
	flag.StringVar(&configFile, "config", configFile, "path to JSON config file")
	flag.StringVar(&config.Address, "a", "", "http server address")
	flag.StringVar(&config.BaseURL, "b", "", "base URL")
	flag.StringVar(&config.FilePath, "f", "", "path to file storage")
	flag.StringVar(&config.DB, "d", "", "db connection settings")
	flag.StringVar(&config.SecretKey, "k", "", "secret key for jwt token")
	flag.BoolVar(&config.EnableHTTPS, "s", false, "enable HTTPS server")
	flag.StringVar(&config.CertFile, "cert", "", "path to certificate file")
	flag.StringVar(&config.KeyFile, "key", "", "path to private key file")

	flag.Parse()

	// 1. Загружаем конфигурацию из JSON файла (если указан)
	if configFile != "" {
		jsonConfig, err := loadJSONConfig(configFile)
		if err != nil {
			return Config{}, fmt.Errorf("failed to load JSON config: %w", err)
		}
		applyJSONConfig(&config, jsonConfig)
	}

	// 2. Применяем переменные среды (перезаписывают значения из JSON)
	if err := env.Parse(&config); err != nil {
		return Config{}, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// 3. Устанавливаем значения по умолчанию для пустых полей
	if config.Address == "" {
		config.Address = "localhost:8080"
	}
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8080"
	}

	return config, nil
}
