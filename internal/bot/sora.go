package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/models"
)

func (b *Bot) handleSora2(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	userID := b.getUserID(update)
	if chatID == 0 || userID == 0 {
		return
	}

	if err := b.stateService.SetMode(ctx, userID, models.ModeSora); err != nil {
		b.handleError(ctx, chatID, err, "Failed to set Sora mode")
		return
	}
	b.showSoraMainScreen(ctx, tgBot, chatID, userID)
}

func (b *Bot) showSoraMainScreen(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	state, err := b.stateService.Get(ctx, userID)
	if err != nil {
		b.handleError(ctx, chatID, err, "Failed to get user state for Sora main screen")
		return
	}

	promptText := "не указан"
	if state.Sora.Prompt != "" {
		if len(state.Sora.Prompt) > 50 {
			promptText = state.Sora.Prompt[:50] + "..."
		} else {
			promptText = state.Sora.Prompt
		}
	}

	imageText := "не добавлено"
	if state.Sora.ImageURL != "" {
		imageText = "добавлено"
	}

	hdIcon := "❌"
	if state.Sora.HD {
		hdIcon = "✅"
	}

	dur10, dur15, dur25 := "10сек", "15сек", "25сек"
	switch state.Sora.Duration {
	case 10:
		dur10 = "✅ 10сек"
	case 15:
		dur15 = "✅ 15сек"
	case 25:
		dur25 = "✅ 25сек"
	}

	fmt16_9, fmt9_16 := "16:9", "9:16"
	if state.Sora.Format == "16:9" {
		fmt16_9 = "✅ 16:9"
	} else {
		fmt9_16 = "✅ 9:16"
	}

	text := fmt.Sprintf(`Sora 2

В этом разделе нужно задать настройки для видео:

1. Опишите видео в разделе «Промпт»
2. Можете добавить изображение: оно станет основой для видео
3. Выберите длительность и соотношение сторон
4. Опция «HD» увеличивает время генерации, но даёт лучшее качество

Текущий промпт: %s
Изображение: %s`, promptText, imageText)

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Промпт", CallbackData: "sora:prompt"}, {Text: "Изображение", CallbackData: "sora:image"}},
			{
				{Text: dur10, CallbackData: "sora:dur_10"},
				{Text: dur15, CallbackData: "sora:dur_15"},
				{Text: dur25, CallbackData: "sora:dur_25"},
			},
			{
				{Text: fmt16_9, CallbackData: "sora:fmt_16_9"},
				{Text: fmt9_16, CallbackData: "sora:fmt_9_16"},
			},
			{{Text: fmt.Sprintf("HD: %s", hdIcon), CallbackData: "sora:hd"}},
			{{Text: "Сгенерировать", CallbackData: "sora:generate"}},
			{{Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) showSoraPromptScreen(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	if err := b.stateService.SetWaitingFor(ctx, userID, models.WaitingSoraPrompt); err != nil {
		b.handleError(ctx, chatID, err, "Failed to set waiting for Sora prompt")
		return
	}

	text := `Отправьте в чат описание видео, которое хотите получить.

Пример:
Камера следует за молодой женщиной в черной кожаной куртке, которая идёт по освещённой улице города ночью. Неоновые вывески магазинов отражаются в лужах на тротуаре.`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "sora:back"}, {Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) showSoraImageScreen(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	if err := b.stateService.SetWaitingFor(ctx, userID, models.WaitingSoraImage); err != nil {
		b.handleError(ctx, chatID, err, "Failed to set waiting for Sora image")
		return
	}

	text := `Отправьте в чат изображение, оно станет основой вашего видео.`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "sora:back"}, {Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) handleSoraCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	data := update.CallbackQuery.Data
	chatID := b.getChatID(update)
	userID := b.getUserID(update)

	tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	switch data {
	case "sora:start":
		b.handleSora2(ctx, tgBot, update)
	case "sora:prompt":
		b.showSoraPromptScreen(ctx, tgBot, chatID, userID)
	case "sora:image":
		b.showSoraImageScreen(ctx, tgBot, chatID, userID)
	case "sora:back":
		if err := b.stateService.ClearWaiting(ctx, userID); err != nil {
			b.handleError(ctx, chatID, err, "Failed to clear waiting for Sora")
			return
		}
		b.showSoraMainScreen(ctx, tgBot, chatID, userID)
	case "sora:dur_10":
		if err := b.stateService.SetSoraDuration(ctx, userID, 10); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set Sora duration")
			return
		}
		b.showSoraMainScreen(ctx, tgBot, chatID, userID)
	case "sora:dur_15":
		if err := b.stateService.SetSoraDuration(ctx, userID, 15); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set Sora duration")
			return
		}
		b.showSoraMainScreen(ctx, tgBot, chatID, userID)
	case "sora:dur_25":
		if err := b.stateService.SetSoraDuration(ctx, userID, 25); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set Sora duration")
			return
		}
		b.showSoraMainScreen(ctx, tgBot, chatID, userID)
	case "sora:fmt_16_9":
		if err := b.stateService.SetSoraFormat(ctx, userID, "16:9"); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set Sora format")
			return
		}
		b.showSoraMainScreen(ctx, tgBot, chatID, userID)
	case "sora:fmt_9_16":
		if err := b.stateService.SetSoraFormat(ctx, userID, "9:16"); err != nil {
			b.handleError(ctx, chatID, err, "Failed to set Sora format")
			return
		}
		b.showSoraMainScreen(ctx, tgBot, chatID, userID)
	case "sora:hd":
		if _, err := b.stateService.ToggleSoraHD(ctx, userID); err != nil {
			b.handleError(ctx, chatID, err, "Failed to toggle Sora HD")
			return
		}
		b.showSoraMainScreen(ctx, tgBot, chatID, userID)
	case "sora:generate":
		b.handleSoraGenerate(ctx, tgBot, chatID, userID)
	}
}

func (b *Bot) handleSoraGenerate(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	state, err := b.stateService.Get(ctx, userID)
	if err != nil {
		b.handleError(ctx, chatID, err, "Failed to get user state for Sora generation")
		return
	}

	if state.Sora.Prompt == "" {
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Сначала укажите промпт для генерации",
		})
		return
	}

	imageStatus := "нет"
	if state.Sora.ImageURL != "" {
		imageStatus = "да"
	}

	text := fmt.Sprintf(`Генерация видео запущена!

Промпт: %s
Длительность: %d сек
Формат: %s
HD: %v
Изображение: %s

(Это заглушка - реальная генерация будет добавлена позже)`, state.Sora.Prompt, state.Sora.Duration, state.Sora.Format, state.Sora.HD, imageStatus)

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Sora 2", CallbackData: "sora:start"}, {Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})

	if err := b.stateService.ResetSora(ctx, userID); err != nil {
		slog.Error("Failed to reset Sora state", "error", err, "user_id", userID)
	}
}
