package bot

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/storage"
)

type Bot struct {
	tgBot   *bot.Bot
	storage *storage.Storage
}

func New(token string, storage *storage.Storage) (*Bot, error) {
	b := &Bot{
		storage: storage,
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
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeExact, b.onBalance)

	return b, nil
}

func (b *Bot) Start(ctx context.Context) {
	b.tgBot.Start(ctx)
}

func (b *Bot) onMessage(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Я получил сообщение. Используйте /start для начала.",
	})
}

func (b *Bot) onStart(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет! Я ИИ бот. Я могу генерировать текст, видео (Sora 2) и фото (NanoBanana). Используй меню для настройки.",
	})
}

func (b *Bot) onBalance(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Твой баланс:\n- GPT: 0\n- Sora 2: 0\n- NanoBanana: 0",
	})
}
