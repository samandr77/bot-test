package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/samandr77/bot-test/internal/ai"
	"github.com/samandr77/bot-test/internal/bot"
	"github.com/samandr77/bot-test/internal/config"
	"github.com/samandr77/bot-test/internal/pkg/logger"
	"github.com/samandr77/bot-test/internal/repository"
	"github.com/samandr77/bot-test/internal/service"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	loadEnvErr := godotenv.Load()
	if loadEnvErr != nil {
		slog.Info("No .env file found or failed to load .env file", "error", loadEnvErr)
	}

	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		slog.Error("Failed to load config", "error", cfgErr)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)

	db, err := repository.New(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to init repository", "error", err)
		os.Exit(1)
	}
	sqlDB, sqlDBErr := db.DB.DB()
	if sqlDBErr != nil {
		slog.Error("Failed to get sql.DB from gorm", "error", sqlDBErr)
		os.Exit(1)
	}
	defer func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.Error("Failed to close database connection", "error", closeErr)
		}
	}()

	stateRepo, err := repository.NewRedisStateRepository(cfg.RedisURL)
	if err != nil {
		slog.Error("Failed to init state repository (Redis)", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := stateRepo.Close(); closeErr != nil {
			slog.Error("Failed to close state repository", "error", closeErr)
		}
	}()

	stateService := service.NewStateService(stateRepo)
	chatService := service.NewChatService(db)

	aiClient := ai.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.AIBaseURL)

	b, err := bot.New(cfg, db, stateService, chatService, aiClient)
	if err != nil {
		slog.Error("Failed to init bot", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting bot...")
	b.Start(ctx)
}
