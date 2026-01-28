package bot

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

func (b *Bot) getChatID(update *tgmodels.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message.Message != nil {
		return update.CallbackQuery.Message.Message.Chat.ID
	}
	return 0
}

func (b *Bot) getUserID(update *tgmodels.Update) int64 {
	if update.Message != nil && update.Message.From != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

func (b *Bot) handleMenu(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	if chatID == 0 {
		return
	}

	if update.CallbackQuery != nil {
		tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
		})
	}

	text := "Выберите нужную модель или команду:"
	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "gpt", CallbackData: "gpt:start"}, {Text: "sora2", CallbackData: "sora:start"}},
			{{Text: "nanobanano", CallbackData: "nano:start"}},
			{{Text: "balance", CallbackData: "balance:start"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) handleCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data
	slog.Debug("Callback received", "data", data)

	switch {
	case data == "menu":
		b.handleMenu(ctx, tgBot, update)
	case strings.HasPrefix(data, "gpt:"):
		b.handleGPTCallback(ctx, tgBot, update)
	case strings.HasPrefix(data, "sora:"):
		b.handleSoraCallback(ctx, tgBot, update)
	case strings.HasPrefix(data, "nano:"):
		b.handleNanoCallback(ctx, tgBot, update)
	case strings.HasPrefix(data, "balance:"):
		b.handleBalanceCallback(ctx, tgBot, update)
	case strings.HasPrefix(data, "buy:"):
		b.handleBuyCallback(ctx, tgBot, update)
	default:
		tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Эта функция будет доступна позже",
		})
	}
}

func (b *Bot) handleBalanceCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	data := update.CallbackQuery.Data
	switch data {
	case "balance:start":
		b.handleBalance(ctx, tgBot, update)
	}
}

func (b *Bot) handleBuyCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	data := update.CallbackQuery.Data
	switch data {
	case "buy:start":
		b.handleBuyMenu(ctx, tgBot, update)
	case "buy:gpt_5":
		b.handleCreateInvoice(ctx, tgBot, update, "gpt", 5, 200)
	case "buy:sora_1":
		b.handleCreateInvoice(ctx, tgBot, update, "sora2", 1, 1000)
	case "buy:nano_5":
		b.handleCreateInvoice(ctx, tgBot, update, "nanobanano", 5, 500)
	}
}
