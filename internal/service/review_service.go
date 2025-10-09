package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"ivanSaichkin/language-bot/internal/domain"
	"ivanSaichkin/language-bot/internal/repository"
)

type reviewService struct {
	wordRepo   repository.WordRepository
	statsRepo  repository.StatsRepository
	repetition SpacedRepetitionService
}

func NewReviewService(
	wordRepo repository.WordRepository,
	statsRepo repository.StatsRepository,
	repetition SpacedRepetitionService,
) ReviewService {
	return &reviewService{
		wordRepo:   wordRepo,
		statsRepo:  statsRepo,
		repetition: repetition,
	}
}

func (s *reviewService) StartReviewSession(ctx context.Context, userID int64, limit int) (*domain.ReviewSession, error) {
	words, err := s.wordRepo.GetWordsForReview(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get words for review: %w", err)
	}

	if len(words) == 0 {
		return nil, fmt.Errorf("no words available for review")
	}

	session := domain.NewReviewSession(userID, words)

	log.Printf("🔄 Started review session for user %d with %d words", userID, len(words))
	return session, nil
}

func (s *reviewService) ProcessAnswer(ctx context.Context, session *domain.ReviewSession, answer string) (*ReviewAnswerResult, error) {
	currentWord := session.GetCurrentWord()
	if currentWord == nil {
		return nil, fmt.Errorf("no current word in session")
	}

	isCorrect := strings.EqualFold(strings.TrimSpace(answer), currentWord.Translation)

	result, err := s.repetition.CalculateNextReview(currentWord, isCorrect)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next review: %w", err)
	}

	currentWord.MarkReviewedWithResult(result, time.Now().Add(result.NextInterval))
	if err := s.wordRepo.Update(ctx, currentWord); err != nil {
		return nil, fmt.Errorf("failed to update word: %w", err)
	}

	startTime := session.StartTime
	session.Answer(isCorrect)

	duration := time.Since(startTime)
	if err := s.statsRepo.AddReview(ctx, session.UserID, isCorrect, duration); err != nil {
		log.Printf("⚠️ Failed to record review stats: %v", err)
	}

	current, total := session.GetProgress()
	progress := &SessionProgress{
		Current:    current,
		Total:      total,
		Correct:    session.CorrectAnswers,
		Accuracy:   session.GetAccuracy(),
		IsComplete: session.IsCompleted,
	}

	reviewResult := &ReviewAnswerResult{
		IsCorrect:       isCorrect,
		CorrectAnswer:   currentWord.Translation,
		NextInterval:    result.NextInterval,
		SessionProgress: progress,
	}

	log.Printf("📝 User %d answered: %s -> %s (correct: %v, quality: %d)",
		session.UserID, currentWord.Original, answer, isCorrect, result.Quality)

	return reviewResult, nil
}

func (s *reviewService) CompleteReviewSession(ctx context.Context, session *domain.ReviewSession) error {
	if !session.IsCompleted {
		session.Complete()
	}

	log.Printf("🏁 Completed review session for user %d: %d/%d correct",
		session.UserID, session.CorrectAnswers, session.TotalQuestions)

	return nil
}

// (заглушка)
func (s *reviewService) GetSession(ctx context.Context, sessionID string) (*domain.ReviewSession, error) {
	return nil, fmt.Errorf("not implemented")
}

// (заглушка)
func (s *reviewService) CleanupOldSessions(ctx context.Context, olderThan time.Duration) (int, error) {
	log.Printf("🧹 Would cleanup review sessions older than %v", olderThan)
	return 0, nil
}
