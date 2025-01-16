package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

// ConfigData представляет конфигурацию приложения.
// Это структура содержит параметры, необходимые для настройки серверного приложения.
type ConfigData struct {
	// ServerAddress определяет адрес HTTP-сервера, на котором будет работать приложение.
	ServerAddress string `json:"server_address"`

	// BaseURL представляет базовый адрес, который используется для сокращенных URL.
	BaseURL string `json:"base_url"`

	// DatabaseDsn представляет строку подключения к базе данных.
	DatabaseDsn string `json:"database_dsn"`
}

// isParsed отслеживает, выполнена ли обработка аргументов командной строки.
var isParsed bool

// ConfigurationInterface интерфейс, в рамках проекта используется для моков юинт тестов
type ConfigurationInterface interface {
	InitConfig() (*ConfigData, error)
}

// Configuration структура конфигурации, реализующая интерфейс ConfigurationInterface
type Configuration struct{}

// InitConfig инициализирует конфигурацию приложения.
func (cs *Configuration) InitConfig() (*ConfigData, error) {
	cfg := &ConfigData{}

	configFile := os.Getenv("CONFIG")
	if configFile == "" {
		configFile = "config.json"
	}

	file, err := os.Open(configFile)
	if err != nil {
		log.Printf("Warning: Error opening config file: %v. Using default configuration.\n", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(cfg); err != nil {
		log.Printf("Warning: Error decoding config file: %v. Using default configuration.\n", err)
	}

	if !isParsed {
		flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080/", "Базовый адрес для сокращенных URL")
		flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Адрес HTTP-сервера")
		flag.StringVar(
			&cfg.DatabaseDsn,
			"d", "",
			"Строка подключения к базе данных")

		flag.Parse()
		isParsed = true
	}

	if ServerAddress := os.Getenv("SERVER_ADDRESS"); ServerAddress != "" {
		cfg.ServerAddress = ServerAddress
	}

	if BaseURL := os.Getenv("BASE_URL"); BaseURL != "" {
		cfg.BaseURL = BaseURL
	}

	if DatabaseDsn := os.Getenv("DATABASE_DSN"); DatabaseDsn != "" {
		cfg.DatabaseDsn = DatabaseDsn
	}

	if cfg.ServerAddress == "" {
		return nil, fmt.Errorf("ServerAddress is required")
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("BaseURL is required")
	}

	return cfg, nil
}
