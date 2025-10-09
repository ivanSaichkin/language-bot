package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ivanSaichkin/language-bot/internal/domain"
)

type statsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) Create(ctx context.Context, stats *domain.UserStats) error {
	query := `
        INSERT INTO user_stats (user_id, total_words, learned_words, total_reviews, total_correct,
                               streak_days, max_streak_days, total_time, last_review_date, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := r.db.ExecContext(ctx, query,
		stats.UserID,
		stats.TotalWords,
		stats.LearnedWords,
		stats.TotalReviews,
		stats.TotalCorrect,
		stats.StreakDays,
		stats.MaxStreakDays,
		stats.TotalTime,
		stats.LastReviewDate,
		stats.CreatedAt,
		stats.UpdatedAt,
	)

	return err
}

func (r *statsRepository) GetByUserID(ctx context.Context, userID int64) (*domain.UserStats, error) {
	query := `
        SELECT user_id, total_words, learned_words, total_reviews, total_correct,
               streak_days, max_streak_days, total_time, last_review_date, created_at, updated_at
        FROM user_stats WHERE user_id = ?
    `

	var stats domain.UserStats
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.UserID,
		&stats.TotalWords,
		&stats.LearnedWords,
		&stats.TotalReviews,
		&stats.TotalCorrect,
		&stats.StreakDays,
		&stats.MaxStreakDays,
		&stats.TotalTime,
		&stats.LastReviewDate,
		&stats.CreatedAt,
		&stats.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &stats, nil
}

func (r *statsRepository) Update(ctx context.Context, stats *domain.UserStats) error {
	query := `
        UPDATE user_stats
        SET total_words = ?, learned_words = ?, total_reviews = ?, total_correct = ?,
            streak_days = ?, max_streak_days = ?, total_time = ?, last_review_date = ?, updated_at = ?
        WHERE user_id = ?
    `

	result, err := r.db.ExecContext(ctx, query,
		stats.TotalWords,
		stats.LearnedWords,
		stats.TotalReviews,
		stats.TotalCorrect,
		stats.StreakDays,
		stats.MaxStreakDays,
		stats.TotalTime,
		stats.LastReviewDate,
		time.Now(),
		stats.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user stats not found")
	}

	return nil
}

func (r *statsRepository) AddReview(ctx context.Context, userID int64, isCorrect bool, duration time.Duration) error {
	stats, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if stats == nil {
		stats = domain.NewUserStats(userID)
		if err := r.Create(ctx, stats); err != nil {
			return err
		}
	}

	stats.AddReview(isCorrect, duration)

	return r.Update(ctx, stats)
}

func (r *statsRepository) UpdateWordCount(ctx context.Context, userID int64, total, learned int) error {
	stats, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if stats == nil {
		stats = domain.NewUserStats(userID)
	}

	stats.UpdateWordCount(total, learned)
	return r.Update(ctx, stats)
}
