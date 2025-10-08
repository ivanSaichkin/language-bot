package bot

import (
	"fmt"
	"ivanSaichkin/language-bot/internal/constants"
	"ivanSaichkin/language-bot/internal/domain"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SimpleHandler struct {
	bot *tgbotapi.BotAPI
}

func NewSimpleHandler(bot *tgbotapi.BotAPI) *SimpleHandler {
	return &SimpleHandler{
		bot: bot,
	}
}

func (h *SimpleHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	log.Printf("Received message from %s: %s",
		update.Message.From.UserName,
		update.Message.Text)

	// Обрабатываем команды
	if update.Message.IsCommand() {
		h.handleCommand(update)
		return
	}

	// Обрабатываем обычные сообщения
	h.handleMessage(update)
}

func (h *SimpleHandler) handleCommand(update tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	command := update.Message.Command()

	var resp string
	switch command {
	case "start":
		resp = `🇬🇧 Language Learning Bot 🇩🇪

Добро пожаловать! Я помогу вам изучать языки с помощью интервального повторения.

📚 Доступные команды:
/add - Добавить новое слово
/review - Повторять слова
/stats - Статистика обучения
/words - Список всех слов
/help - Помощь`

	case "help":
		resp = `📖 Помощь по командам:

/add - Добавить слово в формате: слово - перевод
Пример: hello - привет

/review - Начать сессию повторения слов
/stats - Посмотреть вашу статистику
/words - Показать все ваши слова`

	case "add":
		resp = "Введите слово и перевод через тире:\n\nПример: hello - привет"

	case "review":
		resp = "🔄 Запускаю сессию повторения..."
		// TODO: реализовать в след. частях

	case "stats":
		resp = "📊 Загружаю статистику..."
		// TODO: реализовать в след. частях

	case "words":
		resp = "📖 Ваши слова:\n(функциональность будет добавлена)"
		// TODO: реализовать в след. частях
	default:
		resp = "Неизвестная команда. Используйте /help для списка команд."
	}

	h.sendMessage(chatId, resp)
}

func (h *SimpleHandler) handleMessage(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	if h.isWordFormat(text) {
		h.handleWordAddition(chatID, text)
		return
	}

	response := "Я понимаю только команды и добавление слов в формате: слово - перевод\nИспользуйте /help для справки"
	h.sendMessage(chatID, response)
}

func (h *SimpleHandler) handleWordAddition(chatID int64, text string) {
	original, translation, ok := parseWordInput(text)
	if !ok {
		h.sendMessage(chatID, "Неверный формат. Используйте: слово - перевод")
		return
	}

	word := domain.NewWord(chatID, original, translation, constants.LanguageEnglish)

	response := fmt.Sprintf("✅ Слово добавлено: *%s* - %s\n\nЯзык: английский 🇬🇧",
		word.Original, word.Translation)

	h.sendMessage(chatID, response)
}

func (h *SimpleHandler) isWordFormat(text string) bool {
	for i, char := range text {
		if char == '-' || char == '—' {
			return i > 0 && i < len(text)-1
		}
	}
	return false
}

func (h *SimpleHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "markdown"

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func findSeparator(text string) int {
	for i, char := range text {
		if char == '-' || char == '—' {
			return i
		}
	}
	return -1
}

func cleanText(text string) string {
	return strings.TrimSpace(text)
}

func parseWordInput(text string) (original, translation string, ok bool) {
	separators := []string{" - ", " — ", "-", "—"}

	for _, sep := range separators {
		if idx := strings.Index(text, sep); idx != -1 {
			original = strings.TrimSpace(text[:idx])
			translation = strings.TrimSpace(text[idx+len(sep):])

			if original != "" && translation != "" {
				return original, translation, true
			}
		}
	}

	return "", "", false
}
