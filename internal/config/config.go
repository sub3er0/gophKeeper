package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

// isParsed отслеживает, выполнена ли обработка аргументов командной строки.
var isParsed bool

// ConfigurationInterface интерфейс, в рамках проекта используется для моков юинт тестов
type ConfigurationInterface interface {
	InitConfig() error
}

// Configuration структура конфигурации, реализующая интерфейс ConfigurationInterface
type Configuration struct {
	// ServerAddress определяет адрес HTTP-сервера, на котором будет работать приложение.
	ServerAddress string `json:"server_address"`

	// BaseURL представляет базовый адрес, который используется для сокращенных URL.
	BaseURL string `json:"base_url"`

	// DatabaseDsn представляет строку подключения к базе данных.
	DatabaseDsn string `json:"database_dsn"`
}

// InitConfig инициализирует конфигурацию приложения.
func (cs *Configuration) InitConfig() error {
	configFile := os.Getenv("CONFIG")
	if configFile == "" {
		configFile = "config.json"
	}

	file, err := os.Open(configFile)

	if err != nil {
		log.Printf("Warning: Error opening config file: %v. Using default configuration.\n", err)
	} else {
		isParsed = true
		defer file.Close()

		if err = json.NewDecoder(file).Decode(cs); err != nil {
			log.Printf("Warning: Error decoding config file: %v. Using default configuration.\n", err)
			return err
		}
	}

	if !isParsed {
		flag.StringVar(&cs.BaseURL, "b", "http://localhost:8080/", "Базовый адрес для сокращенных URL")
		flag.StringVar(&cs.ServerAddress, "a", "localhost:8080", "Адрес HTTP-сервера")
		flag.StringVar(
			&cs.DatabaseDsn,
			"d", "",
			"Строка подключения к базе данных")

		flag.Parse()
		isParsed = true
	}

	if ServerAddress := os.Getenv("SERVER_ADDRESS"); ServerAddress != "" {
		cs.ServerAddress = ServerAddress
	}

	if BaseURL := os.Getenv("BASE_URL"); BaseURL != "" {
		cs.BaseURL = BaseURL
	}

	if DatabaseDsn := os.Getenv("DATABASE_DSN"); DatabaseDsn != "" {
		cs.DatabaseDsn = DatabaseDsn
	}

	return nil
}
