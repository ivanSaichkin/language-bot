package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ivanSaichkin/language-bot/internal/domain"
)

type sessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.ReviewSession) error {
	wordsJSON, err := json.Marshal(session.Words)
	if err != nil {
		return fmt.Errorf("failed to marshal words: %w", err)
	}

	query := `
        INSERT INTO review_sessions (id, user_id, correct_answers, total_questions, start_time, end_time, is_completed, words_data)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err = r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.CorrectAnswers,
		session.TotalQuestions,
		session.StartTime,
		session.EndTime,
		session.IsCompleted,
		wordsJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	log.Printf("✅ Created session: %s for user %d", session.ID, session.UserID)
	return nil
}

func (r *sessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.ReviewSession, error) {
	query := `
        SELECT id, user_id, correct_answers, total_questions, start_time, end_time, is_completed, words_data
        FROM review_sessions WHERE id = ?
    `

	var session domain.ReviewSession
	var wordsJSON string
	var endTime sql.NullTime

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.CorrectAnswers,
		&session.TotalQuestions,
		&session.StartTime,
		&endTime,
		&session.IsCompleted,
		&wordsJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var words []*domain.Word
	if err := json.Unmarshal([]byte(wordsJSON), &words); err != nil {
		return nil, fmt.Errorf("failed to unmarshal words: %w", err)
	}
	session.Words = words

	session.CurrentIndex = len(words) - session.TotalQuestions + session.CorrectAnswers
	if session.CurrentIndex < 0 {
		session.CurrentIndex = 0
	}

	if endTime.Valid {
		session.EndTime = endTime.Time
	}

	return &session, nil
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID int64) ([]*domain.ReviewSession, error) {
	query := `
        SELECT id, user_id, correct_answers, total_questions, start_time, end_time, is_completed, words_data
        FROM review_sessions WHERE user_id = ?
        ORDER BY start_time DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.ReviewSession
	for rows.Next() {
		var session domain.ReviewSession
		var wordsJSON string
		var endTime sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.CorrectAnswers,
			&session.TotalQuestions,
			&session.StartTime,
			&endTime,
			&session.IsCompleted,
			&wordsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		var words []*domain.Word
		if err := json.Unmarshal([]byte(wordsJSON), &words); err != nil {
			log.Printf("⚠️ Failed to unmarshal words for session %s: %v", session.ID, err)
			continue
		}
		session.Words = words

		session.CurrentIndex = len(words) - session.TotalQuestions + session.CorrectAnswers
		if session.CurrentIndex < 0 {
			session.CurrentIndex = 0
		}

		if endTime.Valid {
			session.EndTime = endTime.Time
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

func (r *sessionRepository) Update(ctx context.Context, session *domain.ReviewSession) error {
	wordsJSON, err := json.Marshal(session.Words)
	if err != nil {
		return fmt.Errorf("failed to marshal words: %w", err)
	}

	query := `
        UPDATE review_sessions
        SET correct_answers = ?, total_questions = ?, end_time = ?, is_completed = ?, words_data = ?
        WHERE id = ?
    `

	result, err := r.db.ExecContext(ctx, query,
		session.CorrectAnswers,
		session.TotalQuestions,
		session.EndTime,
		session.IsCompleted,
		wordsJSON,
		session.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID string) error {
	query := `DELETE FROM review_sessions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	log.Printf("🗑️ Deleted session: %s", sessionID)
	return nil
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM review_sessions WHERE user_id = ?`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	log.Printf("🗑️ Deleted %d sessions for user %d", rows, userID)
	return nil
}

func (r *sessionRepository) CleanupOldSessions(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-olderThan)

	query := `DELETE FROM review_sessions WHERE (is_completed = 1 AND start_time < ?) OR start_time < ?`

	result, err := r.db.ExecContext(ctx, query, cutoffTime, cutoffTime.Add(-24*time.Hour))
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old sessions: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows > 0 {
		log.Printf("🧹 Cleaned up %d old sessions", rows)
	}

	return int(rows), nil
}

func (r *sessionRepository) GetActiveSessions(ctx context.Context) ([]*domain.ReviewSession, error) {
	query := `
        SELECT id, user_id, correct_answers, total_questions, start_time, end_time, is_completed, words_data
        FROM review_sessions WHERE is_completed = 0
        ORDER BY start_time ASC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.ReviewSession
	for rows.Next() {
		var session domain.ReviewSession
		var wordsJSON string
		var endTime sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.CorrectAnswers,
			&session.TotalQuestions,
			&session.StartTime,
			&endTime,
			&session.IsCompleted,
			&wordsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		var words []*domain.Word
		if err := json.Unmarshal([]byte(wordsJSON), &words); err != nil {
			log.Printf("⚠️ Failed to unmarshal words for session %s: %v", session.ID, err)
			continue
		}
		session.Words = words

		session.CurrentIndex = len(words) - session.TotalQuestions + session.CorrectAnswers
		if session.CurrentIndex < 0 {
			session.CurrentIndex = 0
		}

		if endTime.Valid {
			session.EndTime = endTime.Time
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}
