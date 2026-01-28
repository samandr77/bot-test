package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken     string
	DatabaseURL          string
	RedisURL             string
	PaymentProviderToken string
	LogLevel             string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		TelegramBotToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		RedisURL:             os.Getenv("REDIS_URL"),
		PaymentProviderToken: os.Getenv("PAYMENT_PROVIDER_TOKEN"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, errors.New("TELEGRAM_BOT_TOKEN обязателен")
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL обязателен")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
