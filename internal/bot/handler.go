package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"ivanSaichkin/language-bot/internal/domain"
	"ivanSaichkin/language-bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SimpleHandler struct {
	bot               *tgbotapi.BotAPI
	userService       service.UserService
	wordService       service.WordService
	reviewService     service.ReviewService
	statsService      service.StatsService
	sessionService    service.SessionService
	repetitionService service.SpacedRepetitionService
	sessions          map[int64]*domain.ReviewSession
}

func NewSimpleHandler(
	bot *tgbotapi.BotAPI,
	userService service.UserService,
	wordService service.WordService,
	reviewService service.ReviewService,
	statsService service.StatsService,
	sessionService service.SessionService,
	repetitionService service.SpacedRepetitionService,
) *SimpleHandler {
	return &SimpleHandler{
		bot:               bot,
		userService:       userService,
		wordService:       wordService,
		reviewService:     reviewService,
		statsService:      statsService,
		sessionService:    sessionService,
		repetitionService: repetitionService,
		sessions:          make(map[int64]*domain.ReviewSession),
	}
}

func (h *SimpleHandler) loadUserSessions(ctx context.Context, userID int64) {
	sessions, err := h.sessionService.LoadUserSessions(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to load user sessions: %v", err)
		return
	}

	for _, session := range sessions {
		if !session.IsCompleted {
			h.sessions[userID] = session
			log.Printf("🔄 Loaded active session for user %d: %s", userID, session.ID)
		}
	}
}

func (h *SimpleHandler) saveSession(ctx context.Context, session *domain.ReviewSession) {
	if err := h.sessionService.SaveSession(ctx, session); err != nil {
		log.Printf("⚠️ Failed to save session %s: %v", session.ID, err)
	} else {
		log.Printf("💾 Saved session %s to database", session.ID)
	}
}

func (h *SimpleHandler) deleteSession(ctx context.Context, sessionID string) {
	if err := h.sessionService.DeleteSession(ctx, sessionID); err != nil {
		log.Printf("⚠️ Failed to delete session %s: %v", sessionID, err)
	} else {
		log.Printf("🗑️ Deleted session %s from database", sessionID)
	}
}

func (h *SimpleHandler) HandleUpdate(update tgbotapi.Update) {
	ctx := context.Background()

	if update.Message == nil {
		return
	}

	userID := update.Message.Chat.ID

	if _, exists := h.sessions[userID]; !exists {
		h.loadUserSessions(ctx, userID)
	}

	log.Printf("📨 Received message from %s: %s",
		update.Message.From.UserName,
		update.Message.Text)

	user := domain.NewUser(
		update.Message.Chat.ID,
		update.Message.From.UserName,
		update.Message.From.FirstName,
		update.Message.From.LastName,
		update.Message.From.LanguageCode,
	)

	if err := h.userService.CreateOrUpdateUser(ctx, user); err != nil {
		log.Printf("❌ Failed to create/update user: %v", err)
		h.sendMessage(update.Message.Chat.ID, "❌ Произошла ошибка при обработке запроса")
		return
	}

	if update.Message.IsCommand() {
		h.handleCommand(ctx, update)
		return
	}

	h.handleMessage(ctx, update)
}

func (h *SimpleHandler) handleCommand(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	command := update.Message.Command()

	log.Printf("🔧 Processing command %s for user %d", command, chatID)

	switch command {
	case "start":
		h.handleStartCommand(ctx, chatID, update.Message.From)
	case "help":
		h.handleHelpCommand(chatID)
	case "add":
		h.handleAddCommand(ctx, chatID)
	case "review":
		h.handleReviewCommand(ctx, chatID)
	case "stats":
		h.handleStatsCommand(ctx, chatID)
	case "words":
		h.handleWordsCommand(ctx, chatID)
	case "test":
		h.handleTestCommand(ctx, chatID)
	case "debug":
		h.handleDebugCommand(ctx, chatID)
	case "leaderboard":
		h.handleLeaderboardCommand(ctx, chatID)
	case "goal":
		h.handleGoalCommand(ctx, chatID, update.Message.CommandArguments())
	default:
		h.sendMessage(chatID, "❌ Неизвестная команда. Используйте /help для списка команд.")
	}
}

func (h *SimpleHandler) handleStartCommand(ctx context.Context, chatID int64, from *tgbotapi.User) {
	response := `🇬🇧 *Language Learning Bot* 🇩🇪

Добро пожаловать, *%s*! Я помогу вам изучать языки с помощью интервального повторения.

📚 *Доступные команды:*
/add - Добавить новое слово
/review - Повторять слова
/stats - Статистика обучения
/words - Список всех слов
/leaderboard - Таблица лидеров
/goal - Установить дневную цель
/help - Помощь

💡 *Быстрый старт:*
1. Добавьте слова: hello - привет
2. Повторяйте: /review
3. Следите за прогрессом: /stats`

	h.sendMessage(chatID, fmt.Sprintf(response, from.FirstName))
}

func (h *SimpleHandler) handleHelpCommand(chatID int64) {
	response := `📖 *Помощь по командам*

🎯 *Основные команды:*
/add - Добавить слово в формате: слово - перевод
Пример: hello - привет

/review - Начать сессию повторения слов
/stats - Посмотреть вашу статистику
/words - Показать все ваши слова

📊 *Дополнительные команды:*
/leaderboard - Таблица лидеров среди пользователей
/goal [число] - Установить дневную цель (например: /goal 15)
/debug - Отладочная информация

💡 *Советы:*
• Добавляйте слова с примерами: hello - привет | Hello world!
• Регулярно повторяйте слова с помощью /review
• Старайтесь достигать дневной цели`

	h.sendMessage(chatID, response)
}

func (h *SimpleHandler) handleAddCommand(ctx context.Context, chatID int64) {
	if err := h.userService.SetUserState(ctx, chatID, "awaiting_word"); err != nil {
		h.sendMessage(chatID, "❌ Ошибка при изменении состояния")
		return
	}

	response := `📝 *Добавление нового слова*

Введите слово и перевод через тире:
• Английский: hello - привет
• С примером: book - книга | I read a book

Поддерживаемые языки: английский (en), немецкий (de), французский (fr)`

	h.sendMessage(chatID, response)
}

func (h *SimpleHandler) handleReviewCommand(ctx context.Context, chatID int64) {
	if session, exists := h.sessions[chatID]; exists && !session.IsCompleted {
		h.sendMessage(chatID, "🔁 У вас уже есть активная сессия. Продолжайте отвечать на вопросы.")
		return
	}

	session, err := h.reviewService.StartReviewSession(ctx, chatID, 10)
	if err != nil {
		h.sendMessage(chatID, fmt.Sprintf("❌ Не удалось начать сессию: %v", err))
		return
	}

	h.sessions[chatID] = session
	h.saveSession(ctx, session)
	h.sendNextReviewQuestion(chatID, session)
}

func (h *SimpleHandler) handleStatsCommand(ctx context.Context, chatID int64) {
	stats, err := h.statsService.GetUserStats(ctx, chatID)
	if err != nil {
		h.sendMessage(chatID, "❌ Не удалось загрузить статистику")
		return
	}

	wordProgress, err := h.wordService.GetWordProgress(ctx, chatID)
	if err != nil {
		h.sendMessage(chatID, "❌ Не удалось загрузить прогресс слов")
		return
	}

	streakInfo, err := h.statsService.GetStreakInfo(ctx, chatID)
	if err != nil {
		h.sendMessage(chatID, "❌ Не удалось загрузить информацию о серии")
		return
	}

	response := fmt.Sprintf(`📊 *Ваша статистика*

🎯 *Слова:*
• Всего слов: %d
• Выучено: %d
• В процессе: %d
• Прогресс: %.1f%%

📈 *Эффективность:*
• Всего повторений: %d
• Правильных ответов: %d
• Точность: %.1f%%
• Среднее время: %.1f сек

🔥 *Серия:*
• Текущая серия: %d дней
• Рекорд: %d дней
• Сегодня: %s`,
		wordProgress.TotalWords,
		wordProgress.LearnedWords,
		wordProgress.TotalWords-wordProgress.LearnedWords,
		wordProgress.Progress,
		stats.TotalReviews,
		stats.TotalCorrect,
		stats.GetAccuracy(),
		stats.GetAverageTime(),
		streakInfo.CurrentStreak,
		streakInfo.MaxStreak,
		map[bool]string{true: "✅ выполнено", false: "⏳ осталось"}[streakInfo.IsTodayCompleted],
	)

	h.sendMessage(chatID, response)
}

func (h *SimpleHandler) handleWordsCommand(ctx context.Context, chatID int64) {
	words, err := h.wordService.GetUserWords(ctx, chatID)
	if err != nil {
		h.sendMessage(chatID, "❌ Не удалось загрузить слова")
		return
	}

	if len(words) == 0 {
		h.sendMessage(chatID, "📝 У вас пока нет слов для изучения. Используйте /add чтобы добавить первые слова!")
		return
	}

	var response strings.Builder
	response.WriteString("📖 *Ваши слова*\n\n")

	for i, word := range words {
		if i >= 15 {
			response.WriteString(fmt.Sprintf("\n... и ещё %d слов", len(words)-i))
			break
		}

		status := "🟢"
		if word.IsDueForReview() {
			status = "🟡"
		}
		if word.IsLearned() {
			status = "🔵"
		}

		progress := word.GetProgress()
		response.WriteString(fmt.Sprintf("%s *%s* - %s\n", status, word.Original, word.Translation))
		response.WriteString(fmt.Sprintf("   📊 Прогресс: %.0f%%, Повторений: %d\n\n", progress, word.ReviewCount))
	}

	dueWords, _ := h.wordService.GetDueWords(ctx, chatID)
	if len(dueWords) > 0 {
		response.WriteString(fmt.Sprintf("⏰ *Слов для повторения: %d* /review", len(dueWords)))
	}

	h.sendMessage(chatID, response.String())
}

func (h *SimpleHandler) handleTestCommand(ctx context.Context, chatID int64) {
	response := `🧪 *Тестовый режим*

Эта команда предназначена для тестирования функциональности бота.

Доступные тестовые команды:
/debug - отладочная информация
/stats - просмотр статистики
/words - список слов

Для начала работы добавьте несколько слов с помощью /add`

	h.sendMessage(chatID, response)
}

func (h *SimpleHandler) handleDebugCommand(ctx context.Context, chatID int64) {
	words, err := h.wordService.GetUserWords(ctx, chatID)
	if err != nil {
		h.sendMessage(chatID, "❌ Не удалось загрузить слова для отладки")
		return
	}

	var response strings.Builder
	response.WriteString("🐛 *Отладочная информация*\n\n")

	response.WriteString("👤 *Пользователь:*\n")
	response.WriteString(fmt.Sprintf("ID: %d\n", chatID))

	response.WriteString(fmt.Sprintf("\n📚 *Слова (%d):*\n", len(words)))

	for i, word := range words {
		if i >= 10 {
			response.WriteString(fmt.Sprintf("\n... и ещё %d слов", len(words)-i))
			break
		}

		status := "✅"
		if word.IsDueForReview() {
			status = "⏰"
		}

		response.WriteString(fmt.Sprintf("%s %d. %s - %s\n", status, i+1, word.Original, word.Translation))
		response.WriteString(fmt.Sprintf("   Сложность: %.2f, Повторений: %d, Правильно: %d\n",
			word.Difficulty, word.ReviewCount, word.CorrectAnswers))
		response.WriteString(fmt.Sprintf("   След. повтор: %s\n\n",
			word.NextReview.Format("02.01.2006 15:04")))
	}

	if session, exists := h.sessions[chatID]; exists {
		response.WriteString("🔄 *Активная сессия:*\n")
		response.WriteString(fmt.Sprintf("Слов: %d, Прогресс: %d/%d\n",
			session.TotalQuestions, session.CurrentIndex+1, session.TotalQuestions))
	}

	h.sendMessage(chatID, response.String())
}

func (h *SimpleHandler) handleLeaderboardCommand(ctx context.Context, chatID int64) {
	leaderboard, err := h.statsService.GetLeaderboard(ctx, 10)
	if err != nil {
		h.sendMessage(chatID, "❌ Не удалось загрузить таблицу лидеров")
		return
	}

	var response strings.Builder
	response.WriteString("🏆 *Таблица лидеров*\n\n")

	if len(leaderboard) == 0 {
		response.WriteString("Пока здесь пусто. Будьте первым! 🎯")
		h.sendMessage(chatID, response.String())
		return
	}

	for i, user := range leaderboard {
		medal := ""
		switch i {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("%d.", i+1)
		}

		username := user.FirstName
		if user.Username != "" {
			username = "@" + user.Username
		}

		response.WriteString(fmt.Sprintf("%s %s\n", medal, username))
		response.WriteString(fmt.Sprintf("   📚 Слов: %d, Выучено: %d, Серия: %d дн.\n\n",
			user.TotalWords, user.LearnedWords, user.StreakDays))
	}

	h.sendMessage(chatID, response.String())
}

func (h *SimpleHandler) handleGoalCommand(ctx context.Context, chatID int64, args string) {
	if args == "" {
		user, err := h.userService.GetUser(ctx, chatID)
		if err != nil {
			h.sendMessage(chatID, "❌ Не удалось получить информацию о пользователе")
			return
		}

		h.sendMessage(chatID, fmt.Sprintf("🎯 Ваша текущая дневная цель: *%d слов*", user.DailyGoal))
		return
	}

	var goal int
	if _, err := fmt.Sscanf(args, "%d", &goal); err != nil || goal < 1 || goal > 100 {
		h.sendMessage(chatID, "❌ Неверный формат. Используйте: /goal [число от 1 до 100]")
		return
	}

	if err := h.userService.UpdateDailyGoal(ctx, chatID, goal); err != nil {
		h.sendMessage(chatID, "❌ Не удалось обновить дневную цель")
		return
	}

	h.sendMessage(chatID, fmt.Sprintf("✅ Дневная цель установлена: *%d слов*", goal))
}

func (h *SimpleHandler) handleMessage(ctx context.Context, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	if session, exists := h.sessions[chatID]; exists && !session.IsCompleted {
		h.handleReviewAnswer(ctx, chatID, text, session)
		return
	}

	state, err := h.userService.GetUserState(ctx, chatID)
	if err != nil {
		h.sendMessage(chatID, "❌ Ошибка при получении состояния")
		return
	}

	switch state {
	case "awaiting_word":
		h.handleWordAddition(ctx, chatID, text)
	default:
		h.sendMessage(chatID, "💡 Используйте команды для взаимодействия с ботом. /help - список команд")
	}
}

func (h *SimpleHandler) handleWordAddition(ctx context.Context, chatID int64, text string) {
	original, translation, example, ok := parseWordInput(text)
	if !ok {
		h.sendMessage(chatID, "❌ Неверный формат. Используйте: слово - перевод | пример")
		return
	}

	word := domain.NewWord(chatID, original, translation, "en")
	if example != "" {
		word.WithExample(example)
	}

	if err := h.wordService.AddWord(ctx, word); err != nil {
		h.sendMessage(chatID, "❌ Не удалось добавить слово")
		return
	}

	if err := h.userService.SetUserState(ctx, chatID, ""); err != nil {
		log.Printf("⚠️ Failed to reset user state: %v", err)
	}

	response := fmt.Sprintf("✅ Слово добавлено:\n\n*%s* - %s", word.Original, word.Translation)
	if word.Example != "" {
		response += fmt.Sprintf("\n📝 Пример: %s", word.Example)
	}

	h.sendMessage(chatID, response)
}

func (h *SimpleHandler) handleReviewAnswer(ctx context.Context, chatID int64, answer string, session *domain.ReviewSession) {
	currentWordBefore := session.GetCurrentWord()
	if currentWordBefore == nil {
		h.sendMessage(chatID, "❌ Ошибка: не найдено текущее слово")
		return
	}

	originalWord := currentWordBefore.Original

	result, err := h.reviewService.ProcessAnswer(ctx, session, answer)
	if err != nil {
		h.sendMessage(chatID, "❌ Ошибка при обработке ответа")
		log.Printf("❌ Error processing answer: %v", err)
		return
	}

	var response string
	if result.IsCorrect {
		response = fmt.Sprintf("✅ *%s* - %s", originalWord, result.CorrectAnswer)
	} else {
		response = fmt.Sprintf("❌ *%s* - %s\nПравильный ответ: *%s*",
			originalWord, answer, result.CorrectAnswer)
	}

	h.sendMessage(chatID, response)

	time.Sleep(1 * time.Second)

	if result.SessionProgress.IsComplete {
		h.reviewService.CompleteReviewSession(ctx, session)
		delete(h.sessions, chatID)

		h.saveSession(ctx, session)

		h.showSessionResults(chatID, session)
	} else {
		h.sendNextReviewQuestion(chatID, session)
	}
}

func (h *SimpleHandler) sendNextReviewQuestion(chatID int64, session *domain.ReviewSession) {
	currentWord := session.GetCurrentWord()
	if currentWord == nil {
		h.sendMessage(chatID, "🎉 Все слова пройдены!")
		return
	}

	current, total := session.GetProgress()
	question := fmt.Sprintf("📚 Слово %d/%d\n\n*%s*", current, total, currentWord.Original)

	if currentWord.Example != "" {
		question += fmt.Sprintf("\n\n📝 %s", currentWord.Example)
	}

	log.Printf("🔍 Showing word: %s (correct: %s) to user %d",
		currentWord.Original, currentWord.Translation, chatID)

	h.sendMessage(chatID, question)
}

func (h *SimpleHandler) showSessionResults(chatID int64, session *domain.ReviewSession) {
	duration := session.GetDuration()
	accuracy := session.GetAccuracy()

	response := fmt.Sprintf(`🏁 *Сессия завершена!*

📊 Результаты:
• Правильных ответов: %d/%d
• Точность: %.1f%%
• Время: %.0f сек

🎯 Рекомендации:
%s`,
		session.CorrectAnswers,
		session.TotalQuestions,
		accuracy,
		duration.Seconds(),
		h.getSessionRecommendation(accuracy),
	)

	h.sendMessage(chatID, response)
}

func (h *SimpleHandler) getSessionRecommendation(accuracy float64) string {
	switch {
	case accuracy >= 90:
		return "Отличный результат! Можете увеличить интервалы повторения."
	case accuracy >= 70:
		return "Хороший результат! Продолжайте в том же духе."
	case accuracy >= 50:
		return "Неплохо! Рекомендуем чаще повторять слова."
	default:
		return "Стоит уделить больше времени изучению слов. Попробуйте добавить примеры использования."
	}
}

func (h *SimpleHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("❌ Error sending message: %v", err)
	}
}

func parseWordInput(text string) (original, translation, example string, ok bool) {
	separators := []string{" | ", " - ", " — ", "|", "-", "—"}

	for _, sep := range separators {
		if parts := strings.Split(text, sep); len(parts) >= 2 {
			original = strings.TrimSpace(parts[0])
			translation = strings.TrimSpace(parts[1])

			if len(parts) > 2 {
				example = strings.TrimSpace(parts[2])
			}

			if original != "" && translation != "" {
				return original, translation, example, true
			}
		}
	}

	return "", "", "", false
}
