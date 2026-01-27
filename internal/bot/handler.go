package bot

import (
	"context"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
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
	if chatID == 0 {
		return
	}

	if update.CallbackQuery != nil {
		tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
		})
	}

	text := "Выберите нужную модель или команду:"
	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "gpt", CallbackData: "cmd_gpt"}, {Text: "sora2", CallbackData: "cmd_sora2"}},
			{{Text: "nanobanano", CallbackData: "cmd_nanobanano"}},
			{{Text: "balance", CallbackData: "cmd_balance"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) handleGPT(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	if chatID == 0 {
		return
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Отправьте ваш запрос.",
	})
}

func (b *Bot) handleSora2(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	if chatID == 0 {
		return
	}

	text := `Привет, Sora 2
В этом разделе нужно задать настройки для видео, которое будет сгенерировано с помощью Sora Video 2:

1. Опишите видео в разделе «Промпт»
2. Можете добавить изображение: оно станет основой для видео
3. Выберите длительность (10, 15 или 25 сек.) и соотношение сторон (16:9 или 9:16)
4. Опция «HD» увеличивает время генерации, но даёт лучшее качество

Текущий промпт: не указан
Изображение: не добавлено`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Промпт", CallbackData: "sora_prompt"}, {Text: "Изображение", CallbackData: "sora_image"}},
			{{Text: "Длительность", CallbackData: "sora_duration"}, {Text: "Формат", CallbackData: "sora_format"}},
			{{Text: "HD", CallbackData: "sora_hd"}},
			{{Text: "Сгенерировать", CallbackData: "sora_generate"}},
			{{Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) handleNanoBanana(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	if chatID == 0 {
		return
	}

	text := `Привет, Nano Banano
Рекомендация по использованию:
Отправьте изображения, если хотите изменить ИЛИ напишите промт, что сгенерировать`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Промпт", CallbackData: "nano_prompt"}, {Text: "Изображения", CallbackData: "nano_images"}},
			{{Text: "Сгенерировать", CallbackData: "nano_generate"}},
			{{Text: "Меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) handleBalance(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	userID := b.getUserID(update)

	if chatID == 0 || userID == 0 {
		return
	}

	balanceText, err := b.service.GetBalance(ctx, userID)
	if err != nil {
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Ошибка при получении баланса.",
		})
		return
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   balanceText,
	})
}

func (b *Bot) handleCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data
	switch data {
	case "menu":
		b.handleMenu(ctx, tgBot, update)
	case "cmd_gpt":
		b.handleGPT(ctx, tgBot, update)
	case "cmd_sora2":
		b.handleSora2(ctx, tgBot, update)
	case "cmd_nanobanano":
		b.handleNanoBanana(ctx, tgBot, update)
	case "cmd_balance":
		b.handleBalance(ctx, tgBot, update)
	default:
		tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Эта функция будет доступна позже",
		})
	}
}
