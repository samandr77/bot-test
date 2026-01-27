package bot

import (
	"context"
	"fmt"

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

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "Купить", CallbackData: "buy_menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        balanceText,
		ReplyMarkup: kb,
	})
}

func (b *Bot) handleBuyMenu(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	if chatID == 0 {
		return
	}

	text := "Выберите тариф для пополнения:"
	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "GPT: 5 запросов - 200 руб", CallbackData: "buy_gpt_5"}},
			{{Text: "Sora 2: 1 генерация - 1000 руб", CallbackData: "buy_sora_1"}},
			{{Text: "NanoBanana: 5 генераций - 500 руб", CallbackData: "buy_nano_5"}},
			{{Text: "Назад в меню", CallbackData: "menu"}},
		},
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: kb,
	})
}

func (b *Bot) handleCreateInvoice(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update, modelType string, amount int, price int) {
	chatID := b.getChatID(update)
	if chatID == 0 {
		return
	}

	title := fmt.Sprintf("%s доступ", modelType)
	description := fmt.Sprintf("Доступ к %s %d запросов", modelType, amount)
	payload := fmt.Sprintf("pay_%s_%d", modelType, amount)

	tgBot.SendInvoice(ctx, &bot.SendInvoiceParams{
		ChatID:        chatID,
		Title:         title,
		Description:   description,
		Payload:       payload,
		ProviderToken: b.paymentProviderToken,
		Currency:      "RUB",
		Prices: []tgmodels.LabeledPrice{
			{Label: "Оплата", Amount: price * 100}, // в копейках
		},
	})
}

func (b *Bot) handlePreCheckoutQuery(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	tgBot.AnswerPreCheckoutQuery(ctx, &bot.AnswerPreCheckoutQueryParams{
		PreCheckoutQueryID: update.PreCheckoutQuery.ID,
		OK:                 true,
	})
}

func (b *Bot) handleSuccessfulPayment(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	if update.Message == nil || update.Message.SuccessfulPayment == nil {
		return
	}

	payload := update.Message.SuccessfulPayment.InvoicePayload
	var modelType string
	var amount int
	fmt.Sscanf(payload, "pay_%s_%d", &modelType, &amount)

	userID := update.Message.From.ID
	if err := b.service.AddCredits(ctx, userID, modelType, amount); err != nil {
		tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Ошибка при начислении кредитов. Пожалуйста, обратитесь в поддержку.",
		})
		return
	}

	tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Оплата прошла успешно! Вам начислено %d запросов для %s.", amount, modelType),
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
	case "buy_menu":
		b.handleBuyMenu(ctx, tgBot, update)
	case "buy_gpt_5":
		b.handleCreateInvoice(ctx, tgBot, update, "gpt", 5, 200)
	case "buy_sora_1":
		b.handleCreateInvoice(ctx, tgBot, update, "sora2", 1, 1000)
	case "buy_nano_5":
		b.handleCreateInvoice(ctx, tgBot, update, "nanobanano", 5, 500)
	default:
		tgBot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Эта функция будет доступна позже",
		})
	}
}
