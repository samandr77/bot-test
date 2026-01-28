package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/samandr77/bot-test/internal/bot"
	"github.com/samandr77/bot-test/internal/config"
	"github.com/samandr77/bot-test/internal/repository"
	"github.com/samandr77/bot-test/internal/service"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := repository.New(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to init repository", "error", err)
		os.Exit(1)
	}

	stateRepo, err := repository.NewRedisStateRepository(cfg.RedisURL)
	if err != nil {
		slog.Error("Failed to init state repository (Redis)", "error", err)
		os.Exit(1)
	}
	defer stateRepo.Close()

	stateService := service.NewStateService(stateRepo)

	b, err := bot.New(cfg.TelegramBotToken, db, cfg.PaymentProviderToken, stateService)
	if err != nil {
		slog.Error("Failed to init bot", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting bot...")
	b.Start(ctx)
}
