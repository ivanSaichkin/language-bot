package database

import (
	"database/sql"
	"fmt"
	"log"
)

func InitDatabase(db *sql.DB) error {
	log.Println("🗄️ Initializing database tables...")

	tables := []struct {
		name string
		sql  string
	}{
		{
			name: "users",
			sql: `
				CREATE TABLE IF NOT EXISTS users (
					id BIGINT PRIMARY KEY,
					username VARCHAR(255),
					first_name VARCHAR(255) NOT NULL,
					last_name VARCHAR(255),
					language_code VARCHAR(10) DEFAULT 'ru',
					state VARCHAR(50) DEFAULT '',
					daily_goal INTEGER DEFAULT 10,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				)
			`,
		},
		{
			name: "words",
			sql: `
				CREATE TABLE IF NOT EXISTS words (
					id SERIAL PRIMARY KEY,
					user_id BIGINT NOT NULL,
					original VARCHAR(500) NOT NULL,
					translation VARCHAR(500) NOT NULL,
					language VARCHAR(10) DEFAULT 'en',
					part_of_speech VARCHAR(20) DEFAULT '',
					example TEXT DEFAULT '',
					difficulty FLOAT DEFAULT 2.5,
					next_review TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					review_count INTEGER DEFAULT 0,
					correct_answers INTEGER DEFAULT 0,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				)
			`,
		},
		{
			name: "user_stats",
			sql: `
				CREATE TABLE IF NOT EXISTS user_stats (
					user_id BIGINT PRIMARY KEY,
					total_words INTEGER DEFAULT 0,
					learned_words INTEGER DEFAULT 0,
					total_reviews INTEGER DEFAULT 0,
					total_correct INTEGER DEFAULT 0,
					streak_days INTEGER DEFAULT 0,
					max_streak_days INTEGER DEFAULT 0,
					total_time BIGINT DEFAULT 0,
					last_review_date TIMESTAMP WITH TIME ZONE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				)
			`,
		},
	}

	for _, table := range tables {
		log.Printf("🔧 Creating table: %s", table.name)
		_, err := db.Exec(table.sql)
		if err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.name, err)
		}
		log.Printf("✅ Table %s created/verified", table.name)
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_words_user_id ON words(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_words_next_review ON words(next_review)",
		"CREATE INDEX IF NOT EXISTS idx_users_state ON users(state)",
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			log.Printf("⚠️ Failed to create index: %v", err)
		}
	}

	log.Println("🎉 Database initialization completed successfully!")
	return nil
}
