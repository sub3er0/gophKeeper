package config

import (
	"os"
	"testing"
)

func createTempConfigFile(content string) (string, error) {
	file, err := os.CreateTemp("", "*.json")
	if err != nil {
		return "", err
	}

	if _, err := file.Write([]byte(content)); err != nil {
		return "", err
	}

	if err := file.Close(); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func TestInitConfigFromFile(t *testing.T) {
	tempFile, err := createTempConfigFile(`{
		"server_address": "localhost:8080",
		"base_url": "http://localhost:8080/",
		"database_dsn": "postgres://user:pass@localhost:5432/dbname"
	}`)
	if err != nil {
		t.Fatalf("Не удалось создать временный файл: %v", err)
	}
	defer os.Remove(tempFile) // Удаляем файл после теста

	os.Setenv("CONFIG", tempFile)

	config := Configuration{}
	cfg, err := config.InitConfig()
	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if cfg.ServerAddress != "localhost:8080" {
		t.Errorf("Ожидалось 'localhost:8080', получено '%s'", cfg.ServerAddress)
	}
	if cfg.BaseURL != "http://localhost:8080/" {
		t.Errorf("Ожидалось 'http://localhost:8080/', получено '%s'", cfg.BaseURL)
	}
	if cfg.DatabaseDsn != "postgres://user:pass@localhost:5432/dbname" {
		t.Errorf("Ожидалось 'postgres://user:pass@localhost:5432/dbname', получено '%s'", cfg.DatabaseDsn)
	}
}

func TestInitConfigFromEnv(t *testing.T) {
	os.Setenv("SERVER_ADDRESS", "0.0.0.0:8080")
	os.Setenv("BASE_URL", "http://example.com/")
	os.Setenv("DATABASE_DSN", "mysql://user:pass@localhost:3306/dbname")

	// Инициализация конфигурации
	config := Configuration{}
	cfg, err := config.InitConfig()
	if err != nil {
		t.Fatalf("Неожиданная ошибка: %v", err)
	}

	if cfg.ServerAddress != "0.0.0.0:8080" {
		t.Errorf("Ожидалось '0.0.0.0:8080', получено '%s'", cfg.ServerAddress)
	}
	if cfg.BaseURL != "http://example.com/" {
		t.Errorf("Ожидалось 'http://example.com/', получено '%s'", cfg.BaseURL)
	}
	if cfg.DatabaseDsn != "mysql://user:pass@localhost:3306/dbname" {
		t.Errorf("Ожидалось 'mysql://user:pass@localhost:3306/dbname', получено '%s'", cfg.DatabaseDsn)
	}
}

func TestInitConfigWithInvalidJSON(t *testing.T) {
	tempFile, err := createTempConfigFile(`{
		"server_address": "localhost:8080"
		"base_url": "http://localhost:8080/"}
		"database_dsn": "postgres://user:pass@localhost:5432/dbname"
		}}`)
	if err != nil {
		t.Fatalf("Не удалось создать временный файл: %v", err)
	}
	defer os.Remove(tempFile)

	os.Setenv("CONFIG", tempFile)

	config := Configuration{}
	_, err = config.InitConfig()
	if err == nil {
		t.Fatalf("Ожидалась ошибка, но она не произошла")
	}
}
