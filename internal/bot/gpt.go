package bot

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/ai"
	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/pkg/logger"
)

func (b *Bot) handleGPT(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	userID := b.getUserID(update)
	if chatID == 0 || userID == 0 {
		return
	}

	if modeErr := b.stateService.SetMode(ctx, userID, models.ModeGPT); modeErr != nil {
		b.handleError(ctx, chatID, modeErr, "Failed to set GPT mode")
		return
	}

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}

	b.sendVariant(ctx, chatID, userID, "gpt_start", kb)
}

func (b *Bot) handleGPTCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	data := update.CallbackQuery.Data

	b.answerCallback(ctx, update.CallbackQuery.ID)

	switch data {
	case "gpt", "gpt:start":
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

	b.sendTyping(ctx, chatID)

	aiMessages, contextErr := b.chatService.GetAIChatContext(ctx, userID)
	if contextErr != nil {
		slog.Warn("Failed to get chat context", "userID", userID, "error", contextErr)
	}

	aiMessages = append(aiMessages, ai.Message{
		Role:    "user",
		Content: userText,
	})

	response, aiErr := b.aiClient.Chat(ctx, aiMessages)
	if aiErr != nil {
		b.handleError(ctx, chatID, aiErr, "Failed to get AI response")
		return
	}

	if response == "" {
		b.handleError(ctx, chatID, nil, "AI returned empty response")
		return
	}

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, response, kb)

	if saveUserMsgErr := b.chatService.AddMessage(ctx, userID, "user", userText); saveUserMsgErr != nil {
		slog.Error("Failed to save user message", "userID", userID, "error", saveUserMsgErr)
	}
	if saveAssistantMsgErr := b.chatService.AddMessage(ctx, userID, "assistant", response); saveAssistantMsgErr != nil {
		slog.Error("Failed to save assistant message", "userID", userID, "error", saveAssistantMsgErr)
	}

	if usageErr := b.service.UseCredits(ctx, userID, "gpt"); usageErr != nil {
		traceID := logger.GetTraceID(ctx)
		slog.Error("Failed to deduct GPT credits", "userID", userID, "error", usageErr, "traceID", traceID)
	}
}
