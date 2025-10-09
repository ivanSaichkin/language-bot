package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func NewSQLiteDB() (*sql.DB, error) {
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join("data", "language_bot.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		log.Printf("⚠️ Failed to enable WAL mode: %v", err)
	}

	log.Printf("✅ Connected to SQLite database: %s", dbPath)

	if err := initSQLiteSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func initSQLiteSchema(db *sql.DB) error {
	log.Println("🗄️ Initializing SQLite schema...")

	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            username TEXT,
            first_name TEXT NOT NULL,
            last_name TEXT,
            language_code TEXT DEFAULT 'ru',
            state TEXT DEFAULT '',
            daily_goal INTEGER DEFAULT 10,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,

		`CREATE TABLE IF NOT EXISTS words (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            original TEXT NOT NULL,
            translation TEXT NOT NULL,
            language TEXT DEFAULT 'en',
            part_of_speech TEXT DEFAULT '',
            example TEXT DEFAULT '',
            difficulty REAL DEFAULT 2.5,
            next_review DATETIME DEFAULT CURRENT_TIMESTAMP,
            review_count INTEGER DEFAULT 0,
            correct_answers INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
        )`,

		`CREATE TABLE IF NOT EXISTS user_stats (
            user_id INTEGER PRIMARY KEY,
            total_words INTEGER DEFAULT 0,
            learned_words INTEGER DEFAULT 0,
            total_reviews INTEGER DEFAULT 0,
            total_correct INTEGER DEFAULT 0,
            streak_days INTEGER DEFAULT 0,
            max_streak_days INTEGER DEFAULT 0,
            total_time INTEGER DEFAULT 0,
            last_review_date DATETIME,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
        )`,

		`CREATE TABLE IF NOT EXISTS review_sessions (
            id TEXT PRIMARY KEY,
            user_id INTEGER NOT NULL,
            correct_answers INTEGER DEFAULT 0,
            total_questions INTEGER DEFAULT 0,
            start_time DATETIME DEFAULT CURRENT_TIMESTAMP,
            end_time DATETIME,
            is_completed BOOLEAN DEFAULT FALSE,
            words_data TEXT NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
        )`,
	}

	for i, tableSQL := range tables {
		if _, err := db.Exec(tableSQL); err != nil {
			return fmt.Errorf("failed to create table %d: %w", i+1, err)
		}
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_words_user_id ON words(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_words_next_review ON words(next_review)",
		"CREATE INDEX IF NOT EXISTS idx_users_state ON users(state)",
		"CREATE INDEX IF NOT EXISTS idx_review_sessions_user_id ON review_sessions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_review_sessions_completed ON review_sessions(is_completed)",
		"CREATE INDEX IF NOT EXISTS idx_review_sessions_time ON review_sessions(start_time)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			log.Printf("⚠️ Failed to create index: %v", err)
		}
	}

	log.Println("✅ SQLite schema initialized successfully")
	return nil
}
