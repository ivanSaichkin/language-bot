package main

import (
	"log"

	"ivanSaichkin/language-bot/internal/config"
	"ivanSaichkin/language-bot/internal/database/migrations"
	"ivanSaichkin/language-bot/internal/repository"
)

func main() {
	log.Println("🚀 Starting migration utility...")

	cfg := config.Load()

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	migrator := migrations.NewMigrator(db)

	version, err := migrator.GetDatabaseVersion()
	if err != nil {
		log.Fatalf("❌ Failed to get database version: %v", err)
	}
	log.Printf("📊 Current database version: %d", version)

	if err := migrator.RunMigrations(); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Migration utility completed successfully!")
}
