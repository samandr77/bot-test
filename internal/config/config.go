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
	OpenAIKey            string
	PaymentProviderToken string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		TelegramBotToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		RedisURL:             os.Getenv("REDIS_URL"),
		OpenAIKey:            os.Getenv("OPENAI_API_KEY"),
		PaymentProviderToken: os.Getenv("PAYMENTS_TOKEN"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, errors.New("TELEGRAM_BOT_TOKEN обязателен")
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL обязателен")
	}

	return cfg, nil
}
