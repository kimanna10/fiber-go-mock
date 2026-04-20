package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DB        DBConfig
	JWTSecret string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() *Config {
	// Пытаемся загрузить .env файл, если он есть
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: .env файл не найден, используем системные переменные")
	}

	return &Config{
		Port:      getEnv("PORT", "3000"),
		JWTSecret: getEnv("JWT_ACCESS_SECRET", "super_secret_key"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "12345"),
			Name:     getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
