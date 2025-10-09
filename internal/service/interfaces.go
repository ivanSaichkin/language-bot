package service

import (
	"context"
	"ivanSaichkin/language-bot/internal/domain"
	"time"
)

type UserService interface {
	CreateOrUpdateUser(ctx context.Context, user *domain.User) error
	GetUser(ctx context.Context, userID int64) (*domain.User, error)
	SetUserState(ctx context.Context, userID int64, state string) error
	GetUserState(ctx context.Context, userID int64) (string, error)
	UpdateDailyGoal(ctx context.Context, userID int64, goal int) error
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
}

type WordService interface {
	AddWord(ctx context.Context, word *domain.Word) error
	GetUserWords(ctx context.Context, userID int64) ([]*domain.Word, error)
	GetDueWords(ctx context.Context, userID int64) ([]*domain.Word, error)
	GetWordsForReview(ctx context.Context, userID int64, limit int) ([]*domain.Word, error)
	UpdateWord(ctx context.Context, word *domain.Word) error
	DeleteWord(ctx context.Context, wordID int) error
	GetRandomTranslations(ctx context.Context, userID int64, exclude string, limit int) ([]string, error)
	GetWordProgress(ctx context.Context, userID int64) (*WordProgress, error)
}

type ReviewService interface {
	StartReviewSession(ctx context.Context, userID int64, limit int) (*domain.ReviewSession, error)
	ProcessAnswer(ctx context.Context, session *domain.ReviewSession, answer string) (*ReviewAnswerResult, error)
	CompleteReviewSession(ctx context.Context, session *domain.ReviewSession) error
	GetSession(ctx context.Context, sessionID string) (*domain.ReviewSession, error)
	CleanupOldSessions(ctx context.Context, olderThan time.Duration) (int, error)
}

type StatsService interface {
	GetUserStats(ctx context.Context, userID int64) (*domain.UserStats, error)
	AddReviewRecord(ctx context.Context, userID int64, isCorrect bool, duration time.Duration) error
	GetLeaderboard(ctx context.Context, limit int) ([]*UserStatsRanking, error)
	GetStreakInfo(ctx context.Context, userID int64) (*StreakInfo, error)
	GetDailyProgress(ctx context.Context, userID int64) (*DailyProgress, error)
}

type SpacedRepetitionService interface {
	CalculateNextReview(word *domain.Word, isCorrect bool) (*domain.ReviewResult, error)
	GetWordsForReview(words []*domain.Word) []*domain.Word
	CalculateEaseFactor(word *domain.Word, quality int) float64
}
