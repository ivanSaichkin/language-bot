// cmd/bot/main.go
package main

import (
	"log"

	"ivanSaichkin/language-bot/internal/bot"
	"ivanSaichkin/language-bot/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	log.Println("Starting Language Learning Bot...")

	// Загружаем конфигурацию
	cfg := config.Load()

	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	// Инициализируем бота
	telegramBot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized on account %s", telegramBot.Self.UserName)

	// Инициализируем обработчик
	handler := bot.NewSimpleHandler(telegramBot)

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := telegramBot.GetUpdatesChan(u)

	log.Println("Bot is running and listening for messages...")

	// Основной цикл обработки сообщений
	for update := range updates {
		// Обрабатываем каждое обновление в отдельной горутине
		// для простоты и параллелизма
		go handler.HandleUpdate(update)
	}
}
