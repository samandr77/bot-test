package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := godotenv.Load(); err != nil {
		slog.Warn("Error loading .env file, using system env")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		slog.Error("TELEGRAM_BOT_TOKEN is not set")
		os.Exit(1)
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(onMessage),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, onStart)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeExact, onBalance)

	slog.Info("Bot started")

	b.Start(ctx)
}

func onMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Я получил сообщение. Используйте /start для начала.",
	})
}

func onStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет! Я ИИ бот. Я могу генерировать текст, видео (Sora 2) и фото (NanoBanana). Используй меню для настройки.",
	})
}

func onBalance(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Твой баланс:\n- GPT: 0\n- Sora 2: 0\n- NanoBanana: 0",
	})
}
