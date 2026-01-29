package bot

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/pkg/logger"
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
	userID := b.getUserID(update)
	if chatID == 0 || userID == 0 {
		return
	}

	traceID := logger.GetTraceID(ctx)

	if err := b.stateService.ClearWaiting(ctx, userID); err != nil {
		slog.Warn("Failed to clear waiting state", "handler", "handleMenu", "traceID", traceID, "error", err, "user_id", userID)
	}

	if update.CallbackQuery != nil {
		b.answerCallback(ctx, update.CallbackQuery.ID)
	}

	text := "Выберите нужную модель или команду:"
	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "gpt", CallbackData: "gpt:start"}, {Text: "sora2", CallbackData: "sora:start"}},
			{{Text: "nanobanano", CallbackData: "nano:start"}},
			{{Text: "balance", CallbackData: "balance:start"}},
		},
	}

	b.sendMessage(ctx, chatID, text, kb)
}

func (b *Bot) handleCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data
	traceID := logger.GetTraceID(ctx)
	slog.Info("Callback received", "handler", "handleCallback", "data", data, "traceID", traceID)

	switch {
	case data == "menu":
		b.handleMenu(ctx, tgBot, update)
	case data == "gpt" || strings.HasPrefix(data, "gpt:"):
		b.handleGPTCallback(ctx, tgBot, update)
	case data == "sora" || strings.HasPrefix(data, "sora:"):
		b.handleSoraCallback(ctx, tgBot, update)
	case data == "nano" || strings.HasPrefix(data, "nano:"):
		b.handleNanoCallback(ctx, tgBot, update)
	case data == "nanobanana" || strings.HasPrefix(data, "nanobanana:"):
		b.handleNanoCallback(ctx, tgBot, update)
	case strings.HasPrefix(data, "balance:"):
		b.handleBalanceCallback(ctx, tgBot, update)
	case strings.HasPrefix(data, "buy:"):
		b.handleBuyCallback(ctx, tgBot, update)
	default:
		if _, err := tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Эта функция будет доступна позже",
		}); err != nil {
			slog.Error("Failed to answer callback query with text", "error", err, "callback_id", update.CallbackQuery.ID)
		}
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
	b.answerCallback(ctx, update.CallbackQuery.ID)

	data := update.CallbackQuery.Data
	userID := b.getUserID(update)
	traceID := logger.GetTraceID(ctx)

	slog.Info("Buy callback received", "data", data, "userID", userID, "traceID", traceID)

	switch data {
	case "buy:start":
		slog.Info("Opening buy menu", "userID", userID, "traceID", traceID)
		b.handleBuyMenu(ctx, tgBot, update)
	case "buy:gpt_5":
		slog.Info("Creating invoice for GPT", "userID", userID, "amount", 5, "traceID", traceID)
		b.handleCreateInvoice(ctx, tgBot, update, "gpt", 5, 200)
	case "buy:sora2_1":
		slog.Info("Creating invoice for Sora2", "userID", userID, "amount", 1, "traceID", traceID)
		b.handleCreateInvoice(ctx, tgBot, update, "sora2", 1, 1000)
	case "buy:nanobanana_1":
		slog.Info("Creating invoice for NanoBanana", "userID", userID, "amount", 1, "traceID", traceID)
		b.handleCreateInvoice(ctx, tgBot, update, "nanobanana", 1, 500)
	default:
		slog.Warn("Unknown buy callback", "data", data, "userID", userID, "traceID", traceID)
	}
}
