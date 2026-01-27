package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/samandr77/bot-test/internal/bot"
	"github.com/samandr77/bot-test/internal/config"
	"github.com/samandr77/bot-test/internal/repository"
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

	b, err := bot.New(cfg.TelegramBotToken, db)
	if err != nil {
		slog.Error("Failed to init bot", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting bot...")
	b.Start(ctx)
}
