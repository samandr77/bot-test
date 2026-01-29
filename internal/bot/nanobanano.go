package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/models"
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
	b.showNanoMainScreen(ctx, tgBot, chatID, userID)
}

func (b *Bot) showNanoMainScreen(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	state, getErr := b.stateService.Get(ctx, userID)
	if getErr != nil {
		b.handleError(ctx, chatID, getErr, "Failed to get user state for Nano main screen")
		return
	}

	promptText := "не указан"
	if state.Nano.Prompt != "" {
		runes := []rune(state.Nano.Prompt)
		if len(runes) > 50 {
			promptText = string(runes[:50]) + "..."
		} else {
			promptText = state.Nano.Prompt
		}
	}

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

func (b *Bot) showNanoPromptScreen(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
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

func (b *Bot) showNanoImageScreen(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
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
		b.showNanoPromptScreen(ctx, tgBot, chatID, userID)
	case "nano:images":
		b.showNanoImageScreen(ctx, tgBot, chatID, userID)
	case "nano:back":
		if clearErr := b.stateService.ClearWaiting(ctx, userID); clearErr != nil {
			b.handleError(ctx, chatID, clearErr, "Failed to clear waiting for Nano")
			return
		}
		b.showNanoMainScreen(ctx, tgBot, chatID, userID)
	case "nano:generate":
		b.handleNanoGenerate(ctx, tgBot, chatID, userID)
	}
}

func (b *Bot) handleNanoGenerate(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	hasCredits, creditsErr := b.service.HasCredits(ctx, userID, "nanobanano")
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

	text := fmt.Sprintf(`Генерация изображения запущена!

Промпт: %s
Изображение: %s

(Это заглушка - реальная генерация будет добавлена позже)`,
		func() string {
			if state.Nano.Prompt != "" {
				return state.Nano.Prompt
			}
			return "не указан"
		}(),
		func() string {
			if state.Nano.ImageURL != "" {
				return "загружено"
			}
			return "не загружено"
		}())

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "🍌 NanoBanana", CallbackData: "nano:start"}, {Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}
	b.sendMessage(ctx, chatID, "✅ "+text, kb)

	if usageErr := b.service.UseCredits(ctx, userID, "nanobanano"); usageErr != nil {
		slog.Error("Failed to deduct NanoBanana credits", "error", usageErr, "user_id", userID)
	}

	if resetErr := b.stateService.ResetNano(ctx, userID); resetErr != nil {
		slog.Error("Failed to reset Nano state", "error", resetErr, "user_id", userID)
	}
}
