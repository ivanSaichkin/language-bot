package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	BotWorkers    int
	BotQueueSize  int
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		TelegramToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		BotWorkers:    getEnvAsInt("BOT_WORKERS", 5),
		BotQueueSize:  getEnvAsInt("BOT_QUEUE_SIZE", 1000),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}

	return defaultValue
}
