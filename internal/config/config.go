package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port   string
	AppEnv string
	DBUrl  string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("No .env file found")
	}

	PORT := getEnv("PORT", "8080")
	APP_ENV := getEnv("APP_ENV", "dev")
	DB_URL := getEnv("DB_URL", "")

	return &Config{
		Port:   PORT,
		AppEnv: APP_ENV,
		DBUrl:  DB_URL,
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
