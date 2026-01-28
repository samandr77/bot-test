package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
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
	paymentProviderToken string
}

func New(token string, repo *repository.Repo, paymentToken string, stateService service.StateService) (*Bot, error) {
	svc := service.New(repo)

	b := &Bot{
		repo:                 repo,
		service:              svc,
		stateService:         stateService,
		paymentProviderToken: paymentToken,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(b.onMessage),
		bot.WithMiddlewares(b.LoggingMiddleware),
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
	_, _ = b.tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Извините, произошла техническая ошибка. Мы уже работаем над её исправлением.",
	})
}

func (b *Bot) onMessage(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	state, err := b.stateService.Get(ctx, userID)
	if err != nil {
		b.handleError(ctx, chatID, err, "Failed to get user state")
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
		if err := b.stateService.SetSoraPrompt(ctx, userID, update.Message.Text); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set Sora prompt")
			return
		}
		b.showSoraMainScreen(ctx, tgBot, chatID, userID)
		return
	case models.WaitingNanoPrompt:
		if err := b.stateService.SetNanoPrompt(ctx, userID, update.Message.Text); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set Nano prompt")
			return
		}
		b.showNanoMainScreen(ctx, tgBot, chatID, userID)
		return
	case models.WaitingSoraImage:
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Пожалуйста, отправьте изображение, а не текст.\n\nЕсли хотите добавить описание к видео, используйте раздел «Промпт».",
			ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
					{{Text: "Назад", CallbackData: "sora:back"}, {Text: "Меню", CallbackData: "menu"}},
				},
			},
		})
		return
	case models.WaitingNanoImage:
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Пожалуйста, отправьте изображение, а не текст.\n\nЕсли хотите добавить описание, используйте раздел «Промпт».",
			ReplyMarkup: &tgmodels.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
					{{Text: "Назад", CallbackData: "nano:back"}, {Text: "Меню", CallbackData: "menu"}},
				},
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

	isImage := false
	switch doc.MimeType {
	case "image/jpeg", "image/jpg", "image/png", "image/webp", "image/gif":
		isImage = true
	}

	if !isImage {
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Пожалуйста, отправьте изображение (в формате JPG, PNG, WEBP).",
		})
		return
	}

	b.saveImageFromMessage(ctx, tgBot, update, state, doc.FileID)
}

func (b *Bot) saveImageFromMessage(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update, state *models.UserState, fileID string) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	switch state.WaitingFor {
	case models.WaitingSoraImage:
		if err := b.stateService.SetSoraImage(ctx, userID, fileID); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set Sora image")
			return
		}
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Изображение сохранено!",
		})
		b.showSoraMainScreen(ctx, tgBot, chatID, userID)
	case models.WaitingNanoImage:
		if err := b.stateService.SetNanoImage(ctx, userID, fileID); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set Nano image")
			return
		}
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Изображение сохранено! Теперь напишите, что изменить.",
		})
		if err := b.stateService.SetWaitingFor(ctx, userID, models.WaitingNanoPrompt); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set waiting for Nano prompt")
		}
	default:
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Используйте /menu для выбора модели.",
		})
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

	if err := b.service.SaveUser(ctx, user); err != nil {
		slog.Error("Failed to save user", "error", err)
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет! Я ИИ бот. Я могу генерировать текст, видео (Sora 2) и фото (NanoBanana).",
	})

	b.handleMenu(ctx, tgBot, update)
}
