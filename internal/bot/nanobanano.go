package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/ai"
	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/service"
)

func (b *Bot) handleNanoBanana(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	userID := b.getUserID(update)
	if chatID == 0 || userID == 0 {
		return
	}

	if modeErr := b.stateService.SetMode(ctx, userID, models.ModeNano); modeErr != nil {
		b.handleError(ctx, chatID, modeErr, "Failed to set Nano mode")
		return
	}
	b.showNanoMainScreen(ctx, chatID, userID)
}

func (b *Bot) showNanoMainScreen(ctx context.Context, chatID, userID int64) {
	state, getErr := b.stateService.Get(ctx, userID)
	if getErr != nil {
		b.handleError(ctx, chatID, getErr, "Failed to get user state for Nano main screen")
		return
	}

	promptText := truncateText(state.Nano.Prompt, maxPromptPreviewLength, "не указан")

	imageText := "не добавлено"
	if state.Nano.ImageURL != "" {
		imageText = "добавлено"
	}

	text := "Nano Banano\n\n" +
		"Рекомендация по использованию:\n" +
		"Отправьте изображение, если хотите изменить, или напишите промпт для генерации\n\n" +
		"Текущий промпт: " + promptText + "\n" +
		"Изображение: " + imageText

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "✍️ Промпт", CallbackData: "nano:prompt"}, {Text: "🖼️ Изображения", CallbackData: "nano:images"}},
			{{Text: "🚀 Сгенерировать", CallbackData: "nano:generate"}},
			{{Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, text, kb)
}

func (b *Bot) showNanoPromptScreen(ctx context.Context, chatID, userID int64) {
	if waitErr := b.stateService.SetWaitingFor(ctx, userID, models.WaitingNanoPrompt); waitErr != nil {
		b.handleError(ctx, chatID, waitErr, "Failed to set waiting for Nano prompt")
		return
	}

	text := `Напишите в чат, что хотите сгенерировать.

Пример: Космический корабль летит к далёкой планете на закате`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "⬅️ Назад", CallbackData: "nano:back"}, {Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, text, kb)
}

func (b *Bot) showNanoImageScreen(ctx context.Context, chatID, userID int64) {
	if waitErr := b.stateService.SetWaitingFor(ctx, userID, models.WaitingNanoImage); waitErr != nil {
		b.handleError(ctx, chatID, waitErr, "Failed to set waiting for Nano image")
		return
	}

	text := `Отправьте в чат изображение, которое хотите изменить.

После загрузки напишите, что изменить на изображении.`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "⬅️ Назад", CallbackData: "nano:back"}, {Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, text, kb)
}

func (b *Bot) handleNanoCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	data := update.CallbackQuery.Data
	chatID := b.getChatID(update)
	userID := b.getUserID(update)

	b.answerCallback(ctx, update.CallbackQuery.ID)

	switch data {
	case "nano", "nano:start":
		b.handleNanoBanana(ctx, tgBot, update)
	case "nano:prompt":
		b.showNanoPromptScreen(ctx, chatID, userID)
	case "nano:images":
		b.showNanoImageScreen(ctx, chatID, userID)
	case "nano:back":
		if clearErr := b.stateService.ClearWaiting(ctx, userID); clearErr != nil {
			b.handleError(ctx, chatID, clearErr, "Failed to clear waiting for Nano")
			return
		}
		b.showNanoMainScreen(ctx, chatID, userID)
	case "nano:generate":
		b.handleNanoGenerate(ctx, chatID, userID)
	}
}

func (b *Bot) handleNanoGenerate(ctx context.Context, chatID, userID int64) {
	hasCredits, creditsErr := b.service.HasCredits(ctx, userID, service.ModelNanoBanana)
	if creditsErr != nil {
		b.handleError(ctx, chatID, creditsErr, "Failed to check NanoBanana credits")
		return
	}
	if !hasCredits {
		b.sendNoCreditsMessage(ctx, chatID)
		return
	}

	state, getErr := b.stateService.Get(ctx, userID)
	if getErr != nil {
		b.handleError(ctx, chatID, getErr, "Failed to get user state for Nano generation")
		return
	}

	if state.Nano.Prompt == "" && state.Nano.ImageURL == "" {
		kb := &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{{Text: "✍️ Добавить промпт", CallbackData: "nano:prompt"}, {Text: "🖼️ Изображение", CallbackData: "nano:images"}},
				{{Text: "⬅️ Назад", CallbackData: "nano:back"}, {Text: "🏠 Меню", CallbackData: "menu"}},
			},
		}
		b.sendMessage(ctx, chatID, "Сначала укажите промпт или загрузите изображение.", kb)
		return
	}

	b.sendTyping(ctx, chatID)

	systemPrompt := fmt.Sprintf(`Ты - AI ассистент для генерации описаний изображений.
Пользователь хочет создать изображение со следующими параметрами:
- Промпт: %s
- Исходное изображение: %s

Создай детальное и креативное описание того, как будет выглядеть это изображение.
Опиши композицию, цвета, стиль, настроение. Пиши на русском языке, 2-3 абзаца.`,
		func() string {
			if state.Nano.Prompt != "" {
				return state.Nano.Prompt
			}
			return "не указан"
		}(),
		func() string {
			if state.Nano.ImageURL != "" {
				return "загружено (будет использовано как основа)"
			}
			return "нет"
		}())

	aiMessages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "Создай описание изображения"},
	}

	response, aiErr := b.aiClient.Chat(ctx, aiMessages)
	if aiErr != nil {
		b.handleError(ctx, chatID, aiErr, "Failed to generate image description")
		return
	}

	text := fmt.Sprintf("✅ Генерация изображения завершена!\n\n%s", response)

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "🍌 NanoBanana", CallbackData: "nano:start"}, {Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}
	b.sendMessage(ctx, chatID, text, kb)

	if usageErr := b.service.UseCredits(ctx, userID, service.ModelNanoBanana); usageErr != nil {
		slog.Error("Failed to deduct NanoBanana credits", "error", usageErr, "user_id", userID)
	}

	if resetErr := b.stateService.ResetNano(ctx, userID); resetErr != nil {
		slog.Error("Failed to reset Nano state", "error", resetErr, "user_id", userID)
	}
}
