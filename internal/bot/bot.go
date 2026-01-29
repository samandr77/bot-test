package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/ai"
	"github.com/samandr77/bot-test/internal/config"
	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/pkg/logger"
	"github.com/samandr77/bot-test/internal/repository"
	"github.com/samandr77/bot-test/internal/service"
)

type Bot struct {
	tgBot                *bot.Bot
	repo                 *repository.Repo
	service              service.Service
	stateService         service.StateService
	chatService          service.ChatService
	paymentProviderToken string
	aiClient             ai.AIClient
	cfg                  *config.Config
}

func New(cfg *config.Config, repo *repository.Repo, stateService service.StateService, chatService service.ChatService, aiClient ai.AIClient) (*Bot, error) {
	svc := service.New(repo)

	b := &Bot{
		repo:                 repo,
		service:              svc,
		stateService:         stateService,
		chatService:          chatService,
		paymentProviderToken: cfg.PaymentProviderToken,
		aiClient:             aiClient,
		cfg:                  cfg,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(b.onMessage),
		bot.WithMiddlewares(b.LoggingMiddleware),
	}

	tgBot, botErr := bot.New(cfg.TelegramBotToken, opts...)
	if botErr != nil {
		return nil, fmt.Errorf("failed to create bot: %w", botErr)
	}

	b.tgBot = tgBot

	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, b.onStart)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/menu", bot.MatchTypeExact, b.handleMenu)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/gpt", bot.MatchTypeExact, b.handleGPT)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/sora2", bot.MatchTypeExact, b.handleSora2)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/nanobanano", bot.MatchTypeExact, b.handleNanoBanana)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeExact, b.handleBalance)

	b.tgBot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, b.handleCallback)

	b.tgBot.RegisterHandlerMatchFunc(func(u *tgmodels.Update) bool {
		return u.PreCheckoutQuery != nil
	}, b.handlePreCheckoutQuery)

	b.tgBot.RegisterHandlerMatchFunc(func(u *tgmodels.Update) bool {
		return u.Message != nil && u.Message.SuccessfulPayment != nil
	}, b.handleSuccessfulPayment)

	return b, nil
}

func (b *Bot) Start(ctx context.Context) {
	b.tgBot.Start(ctx)
}

func (b *Bot) handleError(ctx context.Context, chatID int64, err error, message string) {
	traceID := logger.GetTraceID(ctx)
	slog.Error(message, "error", err, "chat_id", chatID, "traceID", traceID)
	if chatID == 0 {
		return
	}
	b.sendMessage(ctx, chatID, "Извините, произошла техническая ошибка. Мы уже работаем над её исправлением.", nil)
}

func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string, kb *tgmodels.InlineKeyboardMarkup) *tgmodels.Message {
	params := &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
		ParseMode:   tgmodels.ParseModeHTML,
	}

	msg, sendErr := b.tgBot.SendMessage(ctx, params)
	if sendErr != nil {
		slog.Error("Failed to send message", "error", sendErr, "chat_id", chatID)
		return nil
	}
	return msg
}

func (b *Bot) sendTyping(ctx context.Context, chatID int64) {
	if _, actionErr := b.tgBot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: tgmodels.ChatActionTyping,
	}); actionErr != nil {
		slog.Error("Failed to send chat action", "error", actionErr, "chat_id", chatID)
	}
}

func (b *Bot) sendVariant(ctx context.Context, chatID int64, userID int64, key string, kb *tgmodels.InlineKeyboardMarkup) *tgmodels.Message {
	variant, variantErr := b.stateService.GetNextVariant(ctx, userID, key)
	if variantErr != nil {
		slog.Warn("Failed to get next variant", "key", key, "userID", userID, "error", variantErr)
	}

	return b.sendMessage(ctx, chatID, variant, kb)
}

func (b *Bot) answerCallback(ctx context.Context, callbackQueryID string) {
	if _, answerErr := b.tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackQueryID,
	}); answerErr != nil {
		slog.Error("Failed to answer callback query", "error", answerErr, "callback_id", callbackQueryID)
	}
}

func (b *Bot) sendNoCreditsMessage(ctx context.Context, chatID int64) {
	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{
				{Text: "💳 Купить", CallbackData: "buy:start"},
				{Text: "🏠 Меню", CallbackData: "menu"},
			},
		},
	}
	b.sendMessage(ctx, chatID, "❌ У тебя закончились запросы", kb)
}

func (b *Bot) onMessage(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	state, stateErr := b.stateService.Get(ctx, userID)
	if stateErr != nil {
		b.handleError(ctx, chatID, stateErr, "Failed to get user state")
		return
	}

	if len(update.Message.Photo) > 0 {
		b.handlePhotoMessage(ctx, tgBot, update, state)
		return
	}

	if update.Message.Document != nil {
		b.handleDocumentMessage(ctx, tgBot, update, state)
		return
	}

	if update.Message.Text == "" {
		return
	}

	switch state.WaitingFor {
	case models.WaitingSoraPrompt:
		if promptErr := b.stateService.SetSoraPrompt(ctx, userID, update.Message.Text); promptErr != nil {
			b.handleError(ctx, chatID, promptErr, "Failed to set Sora prompt")
			return
		}
		b.showSoraMainScreen(ctx, chatID, userID)
		return
	case models.WaitingNanoPrompt:
		if promptErr := b.stateService.SetNanoPrompt(ctx, userID, update.Message.Text); promptErr != nil {
			b.handleError(ctx, chatID, promptErr, "Failed to set Nano prompt")
			return
		}
		b.showNanoMainScreen(ctx, chatID, userID)
		return
	case models.WaitingSoraImage:
		b.sendMessage(ctx, chatID, "Пожалуйста, отправьте изображение, а не текст.\n\nЕсли хотите добавить описание к видео, используйте раздел «Промпт».", &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{{Text: "Назад", CallbackData: "sora:back"}, {Text: "Меню", CallbackData: "menu"}},
			},
		})
		return
	case models.WaitingNanoImage:
		b.sendMessage(ctx, chatID, "Пожалуйста, отправьте изображение, а не текст.\n\nЕсли хотите добавить описание, используйте раздел «Промпт».", &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{{Text: "Назад", CallbackData: "nano:back"}, {Text: "Меню", CallbackData: "menu"}},
			},
		})
		return
	}

	b.handleGPTMessage(ctx, tgBot, update)
}

func (b *Bot) handlePhotoMessage(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update, state *models.UserState) {
	photos := update.Message.Photo
	fileID := photos[len(photos)-1].FileID
	b.saveImageFromMessage(ctx, tgBot, update, state, fileID)
}

func (b *Bot) handleDocumentMessage(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update, state *models.UserState) {
	doc := update.Message.Document
	if doc == nil {
		return
	}

	isImage := strings.HasPrefix(doc.MimeType, "image/")

	if !isImage {
		b.sendMessage(ctx, update.Message.Chat.ID, "Пожалуйста, отправьте изображение (в формате JPG, PNG, WEBP).", nil)
		return
	}

	b.saveImageFromMessage(ctx, tgBot, update, state, doc.FileID)
}

func (b *Bot) saveImageFromMessage(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update, state *models.UserState, fileID string) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	switch state.WaitingFor {
	case models.WaitingSoraImage:
		if imageErr := b.stateService.SetSoraImage(ctx, userID, fileID); imageErr != nil {
			b.handleError(ctx, chatID, imageErr, "Failed to set Sora image")
			return
		}
		b.sendMessage(ctx, chatID, "Изображение сохранено!", nil)
		b.showSoraMainScreen(ctx, chatID, userID)
	case models.WaitingNanoImage:
		if imageErr := b.stateService.SetNanoImage(ctx, userID, fileID); imageErr != nil {
			b.handleError(ctx, chatID, imageErr, "Failed to set Nano image")
			return
		}

		b.showNanoMainScreen(ctx, chatID, userID)

		if waitErr := b.stateService.SetWaitingFor(ctx, userID, models.WaitingNanoPrompt); waitErr != nil {
			b.handleError(ctx, chatID, waitErr, "Failed to set waiting for Nano prompt")
		}
	default:
		b.sendMessage(ctx, chatID, "Используйте /menu для выбора модели.", nil)
	}
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

	if saveErr := b.service.SaveUser(ctx, user); saveErr != nil {
		slog.Error("Failed to save user", "error", saveErr)
	}

	b.sendMessage(ctx, update.Message.Chat.ID, "👋 Привет! Я ИИ бот.\n\nЯ могу генерировать:\n🤖 Текст (GPT)\n🎥 Видео (Sora 2)\n🖼️ Фото (NanoBanana)", nil)

	b.handleMenu(ctx, tgBot, update)
}
