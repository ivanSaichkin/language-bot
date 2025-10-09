package service

import (
	"context"
	"fmt"
	"log"

	"ivanSaichkin/language-bot/internal/domain"
	"ivanSaichkin/language-bot/internal/repository"
)

type wordService struct {
	wordRepo  repository.WordRepository
	statsRepo repository.StatsRepository
}

func NewWordService(
	wordRepo repository.WordRepository,
	statsRepo repository.StatsRepository,
) WordService {
	return &wordService{
		wordRepo:  wordRepo,
		statsRepo: statsRepo,
	}
}

func (s *wordService) AddWord(ctx context.Context, word *domain.Word) error {
	if err := s.validateWord(word); err != nil {
		return fmt.Errorf("word validation failed: %w", err)
	}

	if err := s.wordRepo.Create(ctx, word); err != nil {
		return fmt.Errorf("failed to create word: %w", err)
	}

	if err := s.updateWordStats(ctx, word.UserID); err != nil {
		log.Printf("⚠️ Failed to update word stats: %v", err)
	}

	log.Printf("✅ Added word: %s - %s for user %d",
		word.Original, word.Translation, word.UserID)

	return nil
}

func (s *wordService) GetUserWords(ctx context.Context, userID int64) ([]*domain.Word, error) {
	words, err := s.wordRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user words: %w", err)
	}

	return words, nil
}

func (s *wordService) GetDueWords(ctx context.Context, userID int64) ([]*domain.Word, error) {
	words, err := s.wordRepo.GetDueWords(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get due words: %w", err)
	}

	return words, nil
}

func (s *wordService) GetWordsForReview(ctx context.Context, userID int64, limit int) ([]*domain.Word, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	words, err := s.wordRepo.GetWordsForReview(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get words for review: %w", err)
	}

	return words, nil
}

func (s *wordService) UpdateWord(ctx context.Context, word *domain.Word) error {
	if err := s.wordRepo.Update(ctx, word); err != nil {
		return fmt.Errorf("failed to update word: %w", err)
	}

	return nil
}

func (s *wordService) DeleteWord(ctx context.Context, wordID int) error {
	if err := s.wordRepo.Delete(ctx, wordID); err != nil {
		return fmt.Errorf("failed to delete word: %w", err)
	}

	log.Printf("🗑️ Deleted word: %d", wordID)
	return nil
}

func (s *wordService) GetRandomTranslations(ctx context.Context, userID int64, exclude string, limit int) ([]string, error) {
	translations, err := s.wordRepo.GetRandomTranslations(ctx, userID, exclude, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get random translations: %w", err)
	}

	return translations, nil
}

func (s *wordService) GetWordProgress(ctx context.Context, userID int64) (*WordProgress, error) {
	words, err := s.GetUserWords(ctx, userID)
	if err != nil {
		return nil, err
	}

	dueWords, err := s.GetDueWords(ctx, userID)
	if err != nil {
		return nil, err
	}

	var learnedCount int
	for _, word := range words {
		if word.IsLearned() {
			learnedCount++
		}
	}

	progress := float64(learnedCount) / float64(len(words)) * 100
	if len(words) == 0 {
		progress = 0
	}

	return &WordProgress{
		TotalWords:    len(words),
		LearnedWords:  learnedCount,
		DueWords:      len(dueWords),
		Progress:      progress,
		TodayReviewed: 0,
	}, nil
}

func (s *wordService) validateWord(word *domain.Word) error {
	if word.Original == "" {
		return fmt.Errorf("original word cannot be empty")
	}

	if word.Translation == "" {
		return fmt.Errorf("translation cannot be empty")
	}

	if word.UserID == 0 {
		return fmt.Errorf("user ID cannot be zero")
	}

	if len(word.Original) > 500 {
		return fmt.Errorf("original word too long")
	}

	if len(word.Translation) > 500 {
		return fmt.Errorf("translation too long")
	}

	return nil
}

func (s *wordService) updateWordStats(ctx context.Context, userID int64) error {
	words, err := s.GetUserWords(ctx, userID)
	if err != nil {
		return err
	}

	var learnedCount int
	for _, word := range words {
		if word.IsLearned() {
			learnedCount++
		}
	}

	if err := s.statsRepo.UpdateWordCount(ctx, userID, len(words), learnedCount); err != nil {
		return fmt.Errorf("failed to update word count: %w", err)
	}

	return nil
}
