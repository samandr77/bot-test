package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/repository"
	"github.com/samandr77/bot-test/internal/service"
)

type Bot struct {
	tgBot   *bot.Bot
	repo    *repository.Repo
	service service.Service
}

func New(token string, repo *repository.Repo) (*Bot, error) {
	svc := service.New(repo)

	b := &Bot{
		repo:    repo,
		service: svc,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(b.onMessage),
	}

	tgBot, err := bot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	b.tgBot = tgBot

	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, b.onStart)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/menu", bot.MatchTypeExact, b.handleMenu)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/gpt", bot.MatchTypeExact, b.handleGPT)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/sora2", bot.MatchTypeExact, b.handleSora2)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/nanobanano", bot.MatchTypeExact, b.handleNanoBanana)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeExact, b.handleBalance)

	b.tgBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, b.handleCallback)

	return b, nil
}

func (b *Bot) Start(ctx context.Context) {
	b.tgBot.Start(ctx)
}

func (b *Bot) onMessage(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if update.Message == nil {
		return
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Я получил сообщение. Используйте /menu для управления ботом.",
	})
}

func (b *Bot) onStart(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}

	user := &models.User{
		ID:        update.Message.From.ID,
		Username:  update.Message.From.Username,
		FirstName: update.Message.From.FirstName,
		LastName:  update.Message.From.LastName,
	}

	if err := b.service.SaveUser(ctx, user); err != nil {
		slog.Error("Failed to save user", "error", err)
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет! Я ИИ бот. Я могу генерировать текст, видео (Sora 2) и фото (NanoBanana).",
	})

	b.handleMenu(ctx, tgBot, update)
}
