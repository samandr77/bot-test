package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

func (b *Bot) handleBalance(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	userID := b.getUserID(update)

	if chatID == 0 || userID == 0 {
		return
	}

	balanceText, err := b.service.GetBalance(ctx, userID)
	if err != nil {
		b.handleError(ctx, chatID, err, "Failed to get balance")
		return
	}

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Купить", CallbackData: "buy:start"}, {Text: "Меню", CallbackData: "menu"}},
		},
	}

	if _, err := tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        balanceText,
		ReplyMarkup: kb,
	}); err != nil {
		slog.Error("Failed to send balance message", "error", err, "chat_id", chatID)
	}
}

func (b *Bot) handleBuyMenu(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	if chatID == 0 {
		return
	}

	text := "Выберите тариф для пополнения:"
	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "GPT: 5 запросов - 200 руб", CallbackData: "buy:gpt_5"}},
			{{Text: "Sora 2: 1 генерация - 1000 руб", CallbackData: "buy:sora2_1"}},
			{{Text: "NanoBanana: 1 генерация - 500 руб", CallbackData: "buy:nanobanana_1"}},
			{{Text: "Назад в меню", CallbackData: "menu"}},
		},
	}

	if _, err := tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	}); err != nil {
		slog.Error("Failed to send buy menu message", "error", err, "chat_id", chatID)
	}
}

func (b *Bot) handleCreateInvoice(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update, modelType string, amount int, price int) {
	if b.paymentProviderToken == "" {
		slog.Error("Payment provider token is empty! Cannot create invoice.")
		return
	}
	chatID := b.getChatID(update)
	if chatID == 0 {
		return
	}

	title := fmt.Sprintf("%s доступ", modelType)
	description := fmt.Sprintf("Доступ к %s %d запросов", modelType, amount)
	payload := fmt.Sprintf("pay_%s_%d", modelType, amount)

	_, err := tgBot.SendInvoice(ctx, &bot.SendInvoiceParams{
		ChatID:        chatID,
		Title:         title,
		Description:   description,
		Payload:       payload,
		ProviderToken: b.paymentProviderToken,
		Currency:      "RUB",
		Prices: []tgmodels.LabeledPrice{
			{Label: "Оплата", Amount: price * 100},
		},
	})
	if err != nil {
		slog.Error("Failed to send invoice", "error", err, "user_id", b.getUserID(update))
	}
}

func (b *Bot) handlePreCheckoutQuery(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if _, err := tgBot.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
		PreCheckoutQueryID: update.PreCheckoutQuery.ID,
		OK:                 true,
	}); err != nil {
		slog.Error("Failed to answer pre-checkout query", "error", err, "query_id", update.PreCheckoutQuery.ID)
	}
}

func (b *Bot) handleSuccessfulPayment(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if update.Message == nil || update.Message.SuccessfulPayment == nil {
		return
	}

	payment := update.Message.SuccessfulPayment
	payload := payment.InvoicePayload

	parts := strings.Split(payload, "_")
	if len(parts) != 3 {
		slog.Error("Invalid payment payload format", "payload", payload)
		return
	}

	modelType := parts[1]
	amount, err := strconv.Atoi(parts[2])
	if err != nil {
		slog.Error("Invalid amount in payment payload", "error", err, "payload", payload)
		return
	}

	userID := update.Message.From.ID

	slog.Info("Processing successful payment",
		"userID", userID,
		"modelType", modelType,
		"amount", amount,
		"telegramPaymentID", payment.TelegramPaymentChargeID,
	)

	slog.Info("Attempting to add credits", "userID", userID, "modelType", modelType, "amount", amount)

	if addErr := b.service.AddCredits(ctx, userID, modelType, amount); addErr != nil {
		slog.Error("Failed to add credits", "error", addErr, "userID", userID, "modelType", modelType, "amount", amount)
		b.handleError(ctx, update.Message.Chat.ID, addErr, "Failed to add credits after payment")
		return
	}

	slog.Info("Credits added successfully", "userID", userID, "modelType", modelType, "amount", amount)

	balanceText, err := b.service.GetBalance(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get balance for success message", "error", err, "user_id", userID)
		balanceText = "Ваш баланс обновлен."
	}

	var modelButton tgmodels.InlineKeyboardButton
	switch modelType {
	case "gpt":
		modelButton = tgmodels.InlineKeyboardButton{Text: "GPT", CallbackData: "gpt"}
	case "sora2":
		modelButton = tgmodels.InlineKeyboardButton{Text: "Sora 2", CallbackData: "sora"}
	case "nanobanana", "nanobanano":
		modelButton = tgmodels.InlineKeyboardButton{Text: "NanoBanana", CallbackData: "nano"}
	default:
		modelButton = tgmodels.InlineKeyboardButton{Text: "Меню", CallbackData: "menu"}
	}

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{modelButton, {Text: "Меню", CallbackData: "menu"}},
		},
	}

	if _, err := tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        fmt.Sprintf("Оплата прошла успешно!\n\nВам начислено: %d запр. (%s)\n\n%s", amount, modelType, balanceText),
		ReplyMarkup: kb,
	}); err != nil {
		slog.Error("Failed to send success payment message", "error", err, "user_id", userID)
	}
}
