package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/samandr77/bot-test/internal/pkg/logger"
)

func (b *Bot) LoggingMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
		start := time.Now()
		traceID := uuid.New().String()

		ctx = logger.WithTraceID(ctx, traceID)

		chatID := b.getChatID(update)
		userID := b.getUserID(update)

		var username string
		var chatType string
		var updateID int64 = update.ID
		var messageID int

		if update.Message != nil {
			if update.Message.From != nil {
				username = update.Message.From.Username
			}
			chatType = string(update.Message.Chat.Type)
			messageID = update.Message.ID
		} else if update.CallbackQuery != nil {
			username = update.CallbackQuery.From.Username
			if update.CallbackQuery.Message.Message != nil {
				chatType = string(update.CallbackQuery.Message.Message.Chat.Type)
				messageID = update.CallbackQuery.Message.Message.ID
			}
		}

		slog.Debug("Processing update",
			"traceID", traceID,
			"updateID", updateID,
			"chatID", chatID,
			"userID", userID,
			"username", username,
			"chatType", chatType,
			"messageID", messageID,
		)

		next(ctx, tgBot, update)

		slog.Info("Update handled",
			"traceID", traceID,
			"chatID", chatID,
			"userID", userID,
			"duration", time.Since(start).String(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}
