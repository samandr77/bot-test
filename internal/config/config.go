package config

import (
	"errors"
	"os"
)

type Config struct {
	TelegramBotToken     string
	DatabaseURL          string
	RedisURL             string
	PaymentProviderToken string
	LogLevel             string
	OpenAIAPIKey         string
	OpenAIModel          string
}

func Load() (*Config, error) {
	cfg := &Config{
		TelegramBotToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		RedisURL:             os.Getenv("REDIS_URL"),
		PaymentProviderToken: os.Getenv("PAYMENT_PROVIDER_TOKEN"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		OpenAIAPIKey:         os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:          getEnv("OPENAI_MODEL", "gpt-4"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, errors.New("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return nil, errors.New("REDIS_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
