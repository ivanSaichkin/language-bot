package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ivanSaichkin/language-bot/internal/constants"
	"ivanSaichkin/language-bot/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, username, first_name, last_name, language_code, state, daily_goal, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.LanguageCode,
		string(user.State),
		user.DailyGoal,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Создаем запись статистики для пользователя
	stats := domain.NewUserStats(user.ID)
	return r.createUserStats(ctx, stats)
}

func (r *userRepository) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	query := `
        SELECT id, username, first_name, last_name, language_code, state, daily_goal, created_at, updated_at
        FROM users WHERE id = ?
    `

	var user domain.User
	var state string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.LanguageCode,
		&state,
		&user.DailyGoal,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.State = constants.UserState(state)
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
        UPDATE users
        SET username = ?, first_name = ?, last_name = ?, language_code = ?,
            state = ?, daily_goal = ?, updated_at = ?
        WHERE id = ?
    `

	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.LanguageCode,
		string(user.State),
		user.DailyGoal,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) UpdateState(ctx context.Context, userID int64, state string) error {
	query := `UPDATE users SET state = ?, updated_at = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, state, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user state: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	query := `
        SELECT id, username, first_name, last_name, language_code, state, daily_goal, created_at
        FROM users
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		var state string

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.LanguageCode,
			&state,
			&user.DailyGoal,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.State = constants.UserState(state)
		users = append(users, &user)
	}

	return users, nil
}

func (r *userRepository) createUserStats(ctx context.Context, stats *domain.UserStats) error {
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
