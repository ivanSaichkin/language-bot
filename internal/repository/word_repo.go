package repository

import (
	"context"
	"database/sql"
	"fmt"
	"ivanSaichkin/language-bot/internal/domain"
	"time"
)

type wordRepository struct {
	db *sql.DB
}

func NewWordRepository(db *sql.DB) WordRepository {
	return &wordRepository{db: db}
}

func (r *wordRepository) Create(ctx context.Context, word *domain.Word) error {
	query := `
        INSERT INTO words (user_id, original, translation, language, part_of_speech, example,
                          difficulty, next_review, review_count, correct_answers, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING id
    `

	err := r.db.QueryRowContext(ctx, query,
		word.UserID,
		word.Original,
		word.Translation,
		word.Language,
		word.PartOfSpeech,
		word.Example,
		word.Difficulty,
		word.NextReview,
		word.ReviewCount,
		word.CorrectAnswers,
		word.CreatedAt,
		word.UpdatedAt,
	).Scan(&word.ID)

	if err != nil {
		return fmt.Errorf("failed to create word: %w", err)
	}

	// Обновляем статистику пользователя
	return r.updateUserWordStats(ctx, word.UserID)
}

func (r *wordRepository) GetByID(ctx context.Context, wordID int) (*domain.Word, error) {
	query := `
        SELECT id, user_id, original, translation, language, part_of_speech, example,
               difficulty, next_review, review_count, correct_answers, created_at, updated_at
        FROM words WHERE id = $1
    `

	var word domain.Word
	err := r.db.QueryRowContext(ctx, query, wordID).Scan(
		&word.ID,
		&word.UserID,
		&word.Original,
		&word.Translation,
		&word.Language,
		&word.PartOfSpeech,
		&word.Example,
		&word.Difficulty,
		&word.NextReview,
		&word.ReviewCount,
		&word.CorrectAnswers,
		&word.CreatedAt,
		&word.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get word: %w", err)
	}

	return &word, nil
}

func (r *wordRepository) GetByUserID(ctx context.Context, userID int64) ([]*domain.Word, error) {
	query := `
        SELECT id, user_id, original, translation, language, part_of_speech, example,
               difficulty, next_review, review_count, correct_answers, created_at, updated_at
        FROM words WHERE user_id = $1
        ORDER BY next_review ASC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user words: %w", err)
	}
	defer rows.Close()

	var words []*domain.Word
	for rows.Next() {
		var word domain.Word
		err := rows.Scan(
			&word.ID,
			&word.UserID,
			&word.Original,
			&word.Translation,
			&word.Language,
			&word.PartOfSpeech,
			&word.Example,
			&word.Difficulty,
			&word.NextReview,
			&word.ReviewCount,
			&word.CorrectAnswers,
			&word.CreatedAt,
			&word.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan word: %w", err)
		}
		words = append(words, &word)
	}

	return words, nil
}

func (r *wordRepository) GetDueWords(ctx context.Context, userID int64) ([]*domain.Word, error) {
	query := `
        SELECT id, user_id, original, translation, language, part_of_speech, example,
               difficulty, next_review, review_count, correct_answers, created_at, updated_at
        FROM words
        WHERE user_id = $1 AND next_review <= $2
        ORDER BY next_review ASC
        LIMIT 50
    `

	rows, err := r.db.QueryContext(ctx, query, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get due words: %w", err)
	}
	defer rows.Close()

	var words []*domain.Word
	for rows.Next() {
		var word domain.Word
		err := rows.Scan(
			&word.ID,
			&word.UserID,
			&word.Original,
			&word.Translation,
			&word.Language,
			&word.PartOfSpeech,
			&word.Example,
			&word.Difficulty,
			&word.NextReview,
			&word.ReviewCount,
			&word.CorrectAnswers,
			&word.CreatedAt,
			&word.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan word: %w", err)
		}
		words = append(words, &word)
	}

	return words, nil
}

func (r *wordRepository) GetWordsForReview(ctx context.Context, userID int64, limit int) ([]*domain.Word, error) {
	query := `
        SELECT id, user_id, original, translation, language, part_of_speech, example,
               difficulty, next_review, review_count, correct_answers, created_at, updated_at
        FROM words
        WHERE user_id = $1 AND next_review <= $2
        ORDER BY next_review ASC
        LIMIT $3
    `

	rows, err := r.db.QueryContext(ctx, query, userID, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get words for review: %w", err)
	}
	defer rows.Close()

	var words []*domain.Word
	for rows.Next() {
		var word domain.Word
		err := rows.Scan(
			&word.ID,
			&word.UserID,
			&word.Original,
			&word.Translation,
			&word.Language,
			&word.PartOfSpeech,
			&word.Example,
			&word.Difficulty,
			&word.NextReview,
			&word.ReviewCount,
			&word.CorrectAnswers,
			&word.CreatedAt,
			&word.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan word: %w", err)
		}
		words = append(words, &word)
	}

	return words, nil
}

func (r *wordRepository) Update(ctx context.Context, word *domain.Word) error {
	query := `
        UPDATE words
        SET original = $1, translation = $2, language = $3, part_of_speech = $4, example = $5,
            difficulty = $6, next_review = $7, review_count = $8, correct_answers = $9, updated_at = $10
        WHERE id = $11
    `

	result, err := r.db.ExecContext(ctx, query,
		word.Original,
		word.Translation,
		word.Language,
		word.PartOfSpeech,
		word.Example,
		word.Difficulty,
		word.NextReview,
		word.ReviewCount,
		word.CorrectAnswers,
		time.Now(),
		word.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update word: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("word not found")
	}

	// Обновляем статистику пользователя
	return r.updateUserWordStats(ctx, word.UserID)
}

func (r *wordRepository) Delete(ctx context.Context, wordID int) error {
	word, err := r.GetByID(ctx, wordID)
	if err != nil {
		return err
	}
	if word == nil {
		return fmt.Errorf("word not found")
	}

	query := `DELETE FROM words WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, wordID)
	if err != nil {
		return fmt.Errorf("failed to delete word: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("word not found")
	}

	// Обновляем статистику пользователя
	return r.updateUserWordStats(ctx, word.UserID)
}

func (r *wordRepository) GetRandomTranslations(ctx context.Context, userID int64, exclude string, limit int) ([]string, error) {
	query := `
        SELECT translation FROM words
        WHERE user_id = $1 AND translation != $2
        ORDER BY RANDOM()
        LIMIT $3
    `

	rows, err := r.db.QueryContext(ctx, query, userID, exclude, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get random translations: %w", err)
	}
	defer rows.Close()

	var translations []string
	for rows.Next() {
		var translation string
		if err := rows.Scan(&translation); err != nil {
			return nil, fmt.Errorf("failed to scan translation: %w", err)
		}
		translations = append(translations, translation)
	}

	return translations, nil
}

func (r *wordRepository) updateUserWordStats(ctx context.Context, userID int64) error {
	query := `
        UPDATE user_stats
        SET total_words = (SELECT COUNT(*) FROM words WHERE user_id = $1),
            learned_words = (SELECT COUNT(*) FROM words WHERE user_id = $1 AND correct_answers >= 5),
            updated_at = $2
        WHERE user_id = $1
    `

	_, err := r.db.ExecContext(ctx, query, userID, time.Now())
	return err
}
