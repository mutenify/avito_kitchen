// Package config собирает настройки приложения из переменных окружения.
//
// Вынесен отдельным пакетом (а не читается прямо в main), чтобы main.go
// оставался только "сборкой" зависимостей и не знал деталей вроде имён
// переменных окружения или значений по умолчанию.
package config

import "os"

// Config — все параметры, нужные для запуска Core API.
// У каждого поля есть разумное значение по умолчанию, поэтому
// `go run ./cmd/api` работает и без .env — достаточно локального Postgres
// с этими же дефолтами (см. docker-compose.yml).
type Config struct {
	HTTPPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// Load читает конфигурацию из окружения. Ошибок не возвращает намеренно —
// для MVP отсутствие переменной означает "используем дефолт", а не сбой
// запуска; при необходимости валидацию обязательных полей можно добавить
// здесь позже, не трогая вызывающий код.
func Load() Config {
	return Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "avito_kitchen"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
