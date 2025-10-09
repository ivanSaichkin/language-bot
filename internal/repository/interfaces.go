package repository

import (
	"context"
	"ivanSaichkin/language-bot/internal/domain"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, userID int64) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateState(ctx context.Context, userID int64, state string) error
	GetAll(ctx context.Context) ([]*domain.User, error)
}

type WordRepository interface {
	Create(ctx context.Context, word *domain.Word) error
	GetByID(ctx context.Context, wordID int) (*domain.Word, error)
	GetByUserID(ctx context.Context, userID int64) ([]*domain.Word, error)
	GetDueWords(ctx context.Context, userID int64) ([]*domain.Word, error)
	Update(ctx context.Context, word *domain.Word) error
	Delete(ctx context.Context, wordID int) error
	GetRandomTranslations(ctx context.Context, userID int64, exclude string, limit int) ([]string, error)
	GetWordsForReview(ctx context.Context, userID int64, limit int) ([]*domain.Word, error)
}

type StatsRepository interface {
	Create(ctx context.Context, stats *domain.UserStats) error
	GetByUserID(ctx context.Context, userID int64) (*domain.UserStats, error)
	Update(ctx context.Context, stats *domain.UserStats) error
	AddReview(ctx context.Context, userID int64, isCorrect bool, duration time.Duration) error
	UpdateWordCount(ctx context.Context, userID int64, total, learned int) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.ReviewSession) error
	GetByID(ctx context.Context, sessionID string) (*domain.ReviewSession, error)
	GetByUserID(ctx context.Context, userID int64) ([]*domain.ReviewSession, error)
	Update(ctx context.Context, session *domain.ReviewSession) error
	Delete(ctx context.Context, sessionID string) error
	DeleteByUserID(ctx context.Context, userID int64) error
	CleanupOldSessions(ctx context.Context, olderThan time.Duration) (int, error)
	GetActiveSessions(ctx context.Context) ([]*domain.ReviewSession, error)
}
