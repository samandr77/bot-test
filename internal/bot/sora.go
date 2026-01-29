package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/samandr77/bot-test/internal/ai"
	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/service"
)

const maxPromptPreviewLength = 50

func (b *Bot) handleSora2(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	chatID := b.getChatID(update)
	userID := b.getUserID(update)
	if chatID == 0 || userID == 0 {
		return
	}

	if modeErr := b.stateService.SetMode(ctx, userID, models.ModeSora); modeErr != nil {
		b.handleError(ctx, chatID, modeErr, "Failed to set Sora mode")
		return
	}
	b.showSoraMainScreen(ctx, chatID, userID)
}

func truncateText(text string, maxLen int, defaultText string) string {
	if text == "" {
		return defaultText
	}
	runes := []rune(text)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return text
}

func (b *Bot) showSoraMainScreen(ctx context.Context, chatID, userID int64) {
	state, getErr := b.stateService.Get(ctx, userID)
	if getErr != nil {
		b.handleError(ctx, chatID, getErr, "Failed to get user state for Sora main screen")
		return
	}

	promptText := truncateText(state.Sora.Prompt, maxPromptPreviewLength, "не указан")

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
			{{Text: "✍️ Промпт", CallbackData: "sora:prompt"}, {Text: "🖼️ Изображение", CallbackData: "sora:image"}},
			{
				{Text: dur10, CallbackData: "sora:dur_10"},
				{Text: dur15, CallbackData: "sora:dur_15"},
				{Text: dur25, CallbackData: "sora:dur_25"},
			},
			{
				{Text: fmt16_9, CallbackData: "sora:fmt_16_9"},
				{Text: fmt9_16, CallbackData: "sora:fmt_9_16"},
			},
			{{Text: fmt.Sprintf("📺 HD: %s", hdIcon), CallbackData: "sora:hd"}},
			{{Text: "🚀 Сгенерировать", CallbackData: "sora:generate"}},
			{{Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, text, kb)
}

func (b *Bot) showSoraPromptScreen(ctx context.Context, chatID, userID int64) {
	if waitErr := b.stateService.SetWaitingFor(ctx, userID, models.WaitingSoraPrompt); waitErr != nil {
		b.handleError(ctx, chatID, waitErr, "Failed to set waiting for Sora prompt")
		return
	}

	text := `Отправьте в чат описание видео, которое хотите получить.

Пример:
Камера следует за молодой женщиной в черной кожаной куртке, которая идёт по освещённой улице города ночью. Неоновые вывески магазинов отражаются в лужах на тротуаре.`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "⬅️ Назад", CallbackData: "sora:back"}, {Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, text, kb)
}

func (b *Bot) showSoraImageScreen(ctx context.Context, chatID, userID int64) {
	if waitErr := b.stateService.SetWaitingFor(ctx, userID, models.WaitingSoraImage); waitErr != nil {
		b.handleError(ctx, chatID, waitErr, "Failed to set waiting for Sora image")
		return
	}

	text := `Отправьте в чат изображение, оно станет основой вашего видео.`

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "⬅️ Назад", CallbackData: "sora:back"}, {Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, text, kb)
}

func (b *Bot) handleSoraCallback(ctx context.Context, tgBot *bot.Bot, update *tgmodels.Update) {
	data := update.CallbackQuery.Data
	chatID := b.getChatID(update)
	userID := b.getUserID(update)

	b.answerCallback(ctx, update.CallbackQuery.ID)

	switch {
	case data == "sora" || data == "sora:start":
		b.handleSora2(ctx, tgBot, update)
	case data == "sora:prompt":
		b.showSoraPromptScreen(ctx, chatID, userID)
	case data == "sora:image":
		b.showSoraImageScreen(ctx, chatID, userID)
	case data == "sora:back":
		b.handleSoraBack(ctx, chatID, userID)
	case strings.HasPrefix(data, "sora:dur_"):
		b.handleSoraDuration(ctx, chatID, userID, data)
	case strings.HasPrefix(data, "sora:fmt_"):
		b.handleSoraFormat(ctx, chatID, userID, data)
	case data == "sora:hd":
		b.handleSoraHD(ctx, chatID, userID)
	case data == "sora:generate":
		b.handleSoraGenerate(ctx, chatID, userID)
	}
}

func (b *Bot) handleSoraBack(ctx context.Context, chatID, userID int64) {
	if clearErr := b.stateService.ClearWaiting(ctx, userID); clearErr != nil {
		b.handleError(ctx, chatID, clearErr, "Failed to clear waiting for Sora")
		return
	}
	b.showSoraMainScreen(ctx, chatID, userID)
}

func (b *Bot) handleSoraDuration(ctx context.Context, chatID, userID int64, data string) {
	durStr := strings.TrimPrefix(data, "sora:dur_")
	dur, parseErr := strconv.Atoi(durStr)
	if parseErr != nil {
		b.handleError(ctx, chatID, parseErr, "Invalid duration format")
		return
	}
	if setErr := b.stateService.SetSoraDuration(ctx, userID, dur); setErr != nil {
		b.handleError(ctx, chatID, setErr, "Failed to set Sora duration")
		return
	}
	b.showSoraMainScreen(ctx, chatID, userID)
}

func (b *Bot) handleSoraFormat(ctx context.Context, chatID, userID int64, data string) {
	format := strings.ReplaceAll(strings.TrimPrefix(data, "sora:fmt_"), "_", ":")
	if setErr := b.stateService.SetSoraFormat(ctx, userID, format); setErr != nil {
		b.handleError(ctx, chatID, setErr, "Failed to set Sora format")
		return
	}
	b.showSoraMainScreen(ctx, chatID, userID)
}

func (b *Bot) handleSoraHD(ctx context.Context, chatID, userID int64) {
	if _, toggleErr := b.stateService.ToggleSoraHD(ctx, userID); toggleErr != nil {
		b.handleError(ctx, chatID, toggleErr, "Failed to toggle Sora HD")
		return
	}
	b.showSoraMainScreen(ctx, chatID, userID)
}

func (b *Bot) handleSoraGenerate(ctx context.Context, chatID, userID int64) {
	hasCredits, creditsErr := b.service.HasCredits(ctx, userID, service.ModelSora)
	if creditsErr != nil {
		b.handleError(ctx, chatID, creditsErr, "Failed to check Sora credits")
		return
	}
	if !hasCredits {
		b.sendNoCreditsMessage(ctx, chatID)
		return
	}

	state, getErr := b.stateService.Get(ctx, userID)
	if getErr != nil {
		b.handleError(ctx, chatID, getErr, "Failed to get user state for Sora generation")
		return
	}

	if state.Sora.Prompt == "" {
		text := "Сначала укажите промпт для генерации."
		if state.Sora.ImageURL != "" {
			text = "Вы добавили изображение, теперь добавьте описание (промпт) для видео."
		}

		kb := &tgmodels.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
				{{Text: "➕ Добавить промпт", CallbackData: "sora:prompt"}},
				{{Text: "⬅️ Назад", CallbackData: "sora:back"}, {Text: "🏠 Меню", CallbackData: "menu"}},
			},
		}
		b.sendMessage(ctx, chatID, text, kb)
		return
	}

	b.sendTyping(ctx, chatID)

	systemPrompt := fmt.Sprintf(`Ты - AI ассистент для генерации описаний видео. 
Пользователь хочет создать видео со следующими параметрами:
- Промпт: %s
- Длительность: %d секунд
- Формат: %s
- HD качество: %v
- Изображение в основе: %s

Создай детальное и креативное описание того, как будет выглядеть это видео. 
Опиши сцены, движение камеры, атмосферу. Пиши на русском языке, 2-3 абзаца.`,
		state.Sora.Prompt,
		state.Sora.Duration,
		state.Sora.Format,
		state.Sora.HD,
		func() string {
			if state.Sora.ImageURL != "" {
				return "да"
			}
			return "нет"
		}())

	aiMessages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "Создай описание видео"},
	}

	response, aiErr := b.aiClient.Chat(ctx, aiMessages)
	if aiErr != nil {
		b.handleError(ctx, chatID, aiErr, "Failed to generate video description")
		return
	}

	text := fmt.Sprintf("✅ Генерация видео завершена!\n\n%s", response)

	kb := &tgmodels.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgmodels.InlineKeyboardButton{
			{{Text: "🎥 Sora 2", CallbackData: "sora:start"}, {Text: "🏠 Меню", CallbackData: "menu"}},
		},
	}

	b.sendMessage(ctx, chatID, text, kb)

	if usageErr := b.service.UseCredits(ctx, userID, service.ModelSora); usageErr != nil {
		slog.Error("Failed to deduct Sora credits", "error", usageErr, "user_id", userID)
	}

	if resetErr := b.stateService.ResetSora(ctx, userID); resetErr != nil {
		slog.Error("Failed to reset Sora state", "error", resetErr, "user_id", userID)
	}
}
