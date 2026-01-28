package bot

import (
	"context"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

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
