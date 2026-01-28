package bot

import (
	"context"
	"fmt"

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

	b.stateService.SetMode(ctx, userID, models.ModeNano)
	b.showNanoMainScreen(ctx, tgBot, chatID, userID)
}

func (b *Bot) showNanoMainScreen(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	state := b.stateService.Get(ctx, userID)

	promptText := "не указан"
	if state.Nano.Prompt != "" {
		if len(state.Nano.Prompt) > 50 {
			promptText = state.Nano.Prompt[:50] + "..."
		} else {
			promptText = state.Nano.Prompt
		}
	}

	imageText := "не добавлено"
	if state.Nano.ImageURL != "" {
		imageText = "добавлено"
	}

	text := fmt.Sprintf(`Nano Banano

Рекомендация по использованию:
Отправьте изображения, если хотите изменить ИЛИ напишите промпт, что сгенерировать

Текущий промпт: %s
Изображение: %s`, promptText, imageText)

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Промпт", CallbackData: "nano:prompt"}, {Text: "Изображения", CallbackData: "nano:images"}},
			{{Text: "Сгенерировать", CallbackData: "nano:generate"}},
			{{Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) showNanoPromptScreen(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	b.stateService.SetWaitingFor(ctx, userID, models.WaitingNanoPrompt)

	text := `Напишите в чат, что хотите сгенерировать.

Пример: Космический корабль летит к далёкой планете на закате`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "nano:back"}, {Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) showNanoImageScreen(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	b.stateService.SetWaitingFor(ctx, userID, models.WaitingNanoImage)

	text := `Отправьте в чат изображение, которое хотите изменить.

После загрузки напишите, что изменить на изображении.`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Назад", CallbackData: "nano:back"}, {Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) handleNanoCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	data := update.CallbackQuery.Data
	chatID := b.getChatID(update)
	userID := b.getUserID(update)

	tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	switch data {
	case "nano:start":
		b.handleNanoBanana(ctx, tgBot, update)
	case "nano:prompt":
		b.showNanoPromptScreen(ctx, tgBot, chatID, userID)
	case "nano:images":
		b.showNanoImageScreen(ctx, tgBot, chatID, userID)
	case "nano:back":
		b.stateService.ClearWaiting(ctx, userID)
		b.showNanoMainScreen(ctx, tgBot, chatID, userID)
	case "nano:generate":
		b.handleNanoGenerate(ctx, tgBot, chatID, userID)
	}
}

func (b *Bot) handleNanoGenerate(ctx context.Context, tgBot *bot.Bot, chatID, userID int64) {
	state := b.stateService.Get(ctx, userID)

	if state.Nano.Prompt == "" && state.Nano.ImageURL == "" {
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Сначала укажите промпт или загрузите изображение.",
		})
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
			{{Text: "NanoBanana", CallbackData: "nano:start"}, {Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})

	b.stateService.ResetNano(ctx, userID)
}
