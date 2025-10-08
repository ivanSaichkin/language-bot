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

	cfg := config.Load()

	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	telegramBot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized on account %s", telegramBot.Self.UserName)

	handler := bot.NewSimpleHandler(telegramBot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := telegramBot.GetUpdatesChan(u)

	log.Println("Bot is running and listening for messages...")

	for update := range updates {
		go handler.HandleUpdate(update)
	}
}
