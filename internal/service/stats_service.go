package service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"ivanSaichkin/language-bot/internal/domain"
	"ivanSaichkin/language-bot/internal/repository"
)

type statsService struct {
	userRepo  repository.UserRepository
	wordRepo  repository.WordRepository
	statsRepo repository.StatsRepository
}

func NewStatsService(
	userRepo repository.UserRepository,
	wordRepo repository.WordRepository,
	statsRepo repository.StatsRepository,
) StatsService {
	return &statsService{
		userRepo:  userRepo,
		wordRepo:  wordRepo,
		statsRepo: statsRepo,
	}
}

func (s *statsService) GetUserStats(ctx context.Context, userID int64) (*domain.UserStats, error) {
	stats, err := s.statsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	if stats == nil {
		stats = domain.NewUserStats(userID)
		if err := s.statsRepo.Create(ctx, stats); err != nil {
			return nil, fmt.Errorf("failed to create user stats: %w", err)
		}
	}

	return stats, nil
}

func (s *statsService) AddReviewRecord(ctx context.Context, userID int64, isCorrect bool, duration time.Duration) error {
	if err := s.statsRepo.AddReview(ctx, userID, isCorrect, duration); err != nil {
		return fmt.Errorf("failed to add review record: %w", err)
	}

	return nil
}

func (s *statsService) GetLeaderboard(ctx context.Context, limit int) ([]*UserStatsRanking, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for leaderboard: %w", err)
	}

	var rankings []*UserStatsRanking

	for _, user := range users {
		stats, err := s.GetUserStats(ctx, user.ID)
		if err != nil {
			log.Printf("⚠️ Failed to get stats for user %d: %v", user.ID, err)
			continue
		}

		ranking := &UserStatsRanking{
			UserID:       user.ID,
			Username:     user.Username,
			FirstName:    user.FirstName,
			TotalWords:   stats.TotalWords,
			LearnedWords: stats.LearnedWords,
			StreakDays:   stats.StreakDays,
		}

		rankings = append(rankings, ranking)
	}

	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].LearnedWords > rankings[j].LearnedWords
	})

	for i, ranking := range rankings {
		if i >= limit {
			break
		}
		ranking.Rank = i + 1
	}

	if len(rankings) > limit {
		rankings = rankings[:limit]
	}

	return rankings, nil
}

func (s *statsService) GetStreakInfo(ctx context.Context, userID int64) (*StreakInfo, error) {
	stats, err := s.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	isTodayCompleted := false
	if !stats.LastReviewDate.IsZero() {
		isTodayCompleted = isSameDay(stats.LastReviewDate, time.Now())
	}

	return &StreakInfo{
		CurrentStreak:    stats.StreakDays,
		MaxStreak:        stats.MaxStreakDays,
		LastReviewDate:   stats.LastReviewDate,
		IsTodayCompleted: isTodayCompleted,
	}, nil
}

func (s *statsService) GetDailyProgress(ctx context.Context, userID int64) (*DailyProgress, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for daily progress: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found: %d", userID)
	}

	todayReviewed := 0 // Временная заглушка

	remaining := user.DailyGoal - todayReviewed
	if remaining < 0 {
		remaining = 0
	}

	completionRate := float64(todayReviewed) / float64(user.DailyGoal) * 100
	if user.DailyGoal == 0 {
		completionRate = 0
	}

	return &DailyProgress{
		DailyGoal:      user.DailyGoal,
		TodayReviewed:  todayReviewed,
		Remaining:      remaining,
		CompletionRate: completionRate,
		IsGoalAchieved: todayReviewed >= user.DailyGoal,
	}, nil
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
