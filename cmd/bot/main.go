package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ivanSaichkin/language-bot/internal/bot"
	"ivanSaichkin/language-bot/internal/config"
	"ivanSaichkin/language-bot/internal/repository"
	"ivanSaichkin/language-bot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("🚀 Starting Language Learning Bot with SQLite...")

	cfg := config.Load()
	if cfg.TelegramToken == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN is required. Please check your .env file")
	}

	log.Printf("🔧 Bot Workers: %d", cfg.BotWorkers)
	log.Printf("🔧 Bot Queue Size: %d", cfg.BotQueueSize)
	log.Printf("🤖 Bot Username: %s", cfg.TelegramToken[:10]+"...") // Логируем только часть токена для безопасности

	telegramBot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("❌ Failed to create bot: %v", err)
	}

	telegramBot.Debug = false
	log.Printf("✅ Authorized on account %s", telegramBot.Self.UserName)

	log.Println("🗄️ Initializing SQLite database...")
	db, err := initializeDatabase()
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("⚠️ Failed to close database: %v", err)
		} else {
			log.Println("✅ Database connection closed")
		}
	}()

	log.Println("📦 Initializing services...")
	serviceContainer := initializeServices(db)

	handler := bot.NewSimpleHandler(
		telegramBot,
		serviceContainer.UserService,
		serviceContainer.WordService,
		serviceContainer.ReviewService,
		serviceContainer.StatsService,
		serviceContainer.SessionService,
		serviceContainer.RepetitionService,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupGracefulShutdown(cancel)

	log.Println("🎉 Bot is starting...")
	runBot(ctx, telegramBot, handler, serviceContainer)
}

func initializeDatabase() (*sql.DB, error) {
	maxRetries := 3
	var db *sql.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		log.Printf("🔧 Attempt %d/%d to connect to database...", i+1, maxRetries)

		db, err = repository.NewSQLiteDB()
		if err == nil {
			log.Println("✅ Database connection established")
			return db, nil
		}

		log.Printf("⚠️ Connection failed: %v", err)
		if i < maxRetries-1 {
			log.Printf("🔄 Retrying in 3 seconds...")
			time.Sleep(3 * time.Second)
		}
	}

	return nil, err
}

func initializeServices(db *sql.DB) *service.ServiceContainer {
	log.Println("🔨 Creating repositories...")
	userRepo := repository.NewUserRepository(db)
	wordRepo := repository.NewWordRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	log.Println("🔨 Creating services...")
	return service.NewServiceContainer(userRepo, wordRepo, statsRepo, sessionRepo)
}

func runBot(ctx context.Context, botAPI *tgbotapi.BotAPI, handler *bot.SimpleHandler, services *service.ServiceContainer) {
	log.Println("📡 Setting up updates channel...")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.Limit = 100
	u.Offset = 0

	updates := botAPI.GetUpdatesChan(u)

	log.Println("🔄 Starting background tasks...")
	go startBackgroundTasks(ctx, botAPI, services.UserService, services.SessionService)

	log.Println("🎊 Bot is now running and listening for messages!")
	log.Println("💡 Send /start to begin your language learning journey")
	log.Println("🔔 Use /help to see all available commands")
	log.Println("===========================================================================================")

	messageCount := 0
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Shutting down bot...")
			log.Printf("📊 Processed %d messages in %v", messageCount, time.Since(startTime))
			return

		case update := <-updates:
			messageCount++
			go func(update tgbotapi.Update) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("⚠️ Recovered from panic in message processing: %v", r)
					}
				}()

				if update.Message != nil {
					user := update.Message.From
					messagePreview := update.Message.Text
					if len(messagePreview) > 50 {
						messagePreview = messagePreview[:50] + "..."
					}
					log.Printf("📨 Message from @%s (%s): %s",
						user.UserName, user.FirstName, messagePreview)
				}

				handler.HandleUpdate(update)
			}(update)

			if messageCount%100 == 0 {
				log.Printf("📊 Processed %d messages so far...", messageCount)
			}
		}
	}
}

func startBackgroundTasks(
	ctx context.Context,
	botAPI *tgbotapi.BotAPI,
	userService service.UserService,
	sessionService service.SessionService,
) {
	log.Println("⏰ Starting background tasks scheduler...")

	go startSessionCleanup(ctx, sessionService)

	go startActivityMonitor(ctx, botAPI, userService)

	go startReminderService(ctx, botAPI, userService)

	go startUsageStats(ctx, sessionService)

	log.Println("✅ All background tasks started successfully")
}

func startSessionCleanup(ctx context.Context, sessionService service.SessionService) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("🧹 Session cleanup task started")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Stopping session cleanup...")
			return
		case <-ticker.C:
			cleaned, err := sessionService.CleanupOldSessions(ctx, 24*time.Hour)
			if err != nil {
				log.Printf("⚠️ Failed to cleanup sessions: %v", err)
			} else if cleaned > 0 {
				log.Printf("🧹 Cleaned up %d old sessions from database", cleaned)
			}

			activeCount := sessionService.GetActiveSessionsCount(ctx)
			if activeCount > 0 {
				log.Printf("📊 Active sessions in database: %d", activeCount)
			}
		}
	}
}

func startActivityMonitor(ctx context.Context, botAPI *tgbotapi.BotAPI, userService service.UserService) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	log.Println("📊 Activity monitor started")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Stopping activity monitor...")
			return
		case <-ticker.C:
			monitorUserActivity(ctx, botAPI, userService)
		}
	}
}

func startReminderService(ctx context.Context, botAPI *tgbotapi.BotAPI, userService service.UserService) {
	ticker := time.NewTicker(24 * time.Hour) // Раз в день
	defer ticker.Stop()

	log.Println("⏰ Reminder service started")

	time.Sleep(1 * time.Minute)
	sendDailyReminders(ctx, botAPI, userService)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Stopping reminder service...")
			return
		case <-ticker.C:
			sendDailyReminders(ctx, botAPI, userService)
		}
	}
}

func startUsageStats(ctx context.Context, sessionService service.SessionService) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	log.Println("📈 Usage statistics collector started")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Stopping usage statistics...")
			return
		case <-ticker.C:
			activeSessions := sessionService.GetActiveSessionsCount(ctx)
			if activeSessions > 0 {
				log.Printf("👥 Currently %d active learning sessions", activeSessions)
			}
		}
	}
}

func sendDailyReminders(ctx context.Context, botAPI *tgbotapi.BotAPI, userService service.UserService) {
	log.Println("💌 Sending daily reminders...")

	users, err := userService.GetAllUsers(ctx)
	if err != nil {
		log.Printf("⚠️ Failed to get users for reminders: %v", err)
		return
	}

	log.Printf("👥 Found %d users in database", len(users))

	remindersSent := 0
	inactiveThreshold := 48 * time.Hour

	for _, user := range users {
		if time.Since(user.UpdatedAt) > inactiveThreshold {
			sendReminder(botAPI, user.ID)
			remindersSent++
			time.Sleep(100 * time.Millisecond)
		}
	}

	if remindersSent > 0 {
		log.Printf("✅ Sent reminders to %d inactive users", remindersSent)
	} else {
		log.Println("💤 No inactive users found for reminders")
	}
}

func sendReminder(botAPI *tgbotapi.BotAPI, userID int64) {
	msg := tgbotapi.NewMessage(userID,
		"📚 *Привет! Давно не виделись* 👋\n\n"+
			"Готовы продолжить изучение языков? Не забывайте, что регулярные повторения — ключ к успеху! 🗝️\n\n"+
			"✨ *Что можно сделать сейчас:*\n"+
			"• /review - повторить слова\n"+
			"• /add - добавить новые слова\n"+
			"• /stats - посмотреть прогресс\n\n"+
			"*Уделите всего 5 минут в день для заметных результатов!* ⏱️")

	msg.ParseMode = "Markdown"

	if _, err := botAPI.Send(msg); err != nil {
		log.Printf("⚠️ Failed to send reminder to user %d: %v", userID, err)
	} else {
		log.Printf("✅ Sent reminder to user %d", userID)
	}
}

func monitorUserActivity(ctx context.Context, botAPI *tgbotapi.BotAPI, userService service.UserService) {
	users, err := userService.GetAllUsers(ctx)
	if err != nil {
		log.Printf("⚠️ Failed to get users for activity monitoring: %v", err)
		return
	}

	now := time.Now()
	inactiveThreshold := 7 * 24 * time.Hour

	var activeUsers, inactiveUsers int
	for _, user := range users {
		if now.Sub(user.UpdatedAt) > inactiveThreshold {
			inactiveUsers++
		} else {
			activeUsers++
		}
	}

	log.Printf("📊 User activity: %d active, %d inactive (>7 days)", activeUsers, inactiveUsers)

	if len(users) > 0 {
		activePercentage := float64(activeUsers) / float64(len(users)) * 100
		log.Printf("📈 Active users: %.1f%%", activePercentage)
	}
}

func setupGracefulShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		log.Printf("🛑 Received signal: %v. Shutting down gracefully...", sig)

		cancel()

		shutdownTimeout := 5 * time.Second
		log.Printf("⏳ Waiting %v for cleanup operations...", shutdownTimeout)

		go func() {
			for i := 0; i < int(shutdownTimeout.Seconds()); i++ {
				log.Printf("⏰ Shutting down in %d seconds...", int(shutdownTimeout.Seconds())-i)
				time.Sleep(1 * time.Second)
			}
		}()

		time.Sleep(shutdownTimeout)

		log.Println("✅ Cleanup completed")
		log.Println("👋 Bot shutdown successfully")
		os.Exit(0)
	}()
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
