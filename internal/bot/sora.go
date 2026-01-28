package bot

import (
	"context"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

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
			{{Text: "Промпт", CallbackData: "sora:prompt"}, {Text: "Изображение", CallbackData: "sora:image"}},
			{{Text: "Длительность", CallbackData: "sora:duration"}, {Text: "Формат", CallbackData: "sora:format"}},
			{{Text: "HD", CallbackData: "sora:hd"}},
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
