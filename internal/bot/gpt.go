package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/pkg/logger"
)

func (b *Bot) handleGPT(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	userID := b.getUserID(update)
	if chatID == 0 || userID == 0 {
		return
	}

	if err := b.stateService.SetMode(ctx, userID, models.ModeGPT); err != nil {
		b.handleError(ctx, chatID, err, "Failed to set GPT mode")
		return
	}

	text := `GPT

Отправьте ваш запрос, и я отвечу.`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, text, kb)
}

func (b *Bot) handleGPTCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	data := update.CallbackQuery.Data

	b.answerCallback(ctx, update.CallbackQuery.ID)

	switch data {
	case "gpt:start":
		b.handleGPT(ctx, tgBot, update)
	}
}

func (b *Bot) handleGPTMessage(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	userID := b.getUserID(update)

	if update.Message == nil || update.Message.Text == "" {
		return
	}

	hasCredits, creditsErr := b.service.HasCredits(ctx, userID, "gpt")
	if creditsErr != nil {
		b.handleError(ctx, chatID, creditsErr, "Failed to check GPT credits")
		return
	}
	if !hasCredits {
		b.sendNoCreditsMessage(ctx, chatID)
		return
	}

	userText := update.Message.Text

	response := fmt.Sprintf("GPT ответ на: %s\n\n(Это заглушка - реальный GPT будет добавлен позже)", userText)

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, response, kb)

	if err := b.stateService.SetMode(ctx, userID, models.ModeGPT); err != nil {
		traceID := logger.GetTraceID(ctx)
		slog.Error("Failed to reset GPT mode", "handler", "handleGPTMessage", "error", err, "user_id", userID, "traceID", traceID)
	}
}
