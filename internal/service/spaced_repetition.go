package service

import (
	"ivanSaichkin/language-bot/internal/domain"
	"math"
	"time"
)

type spacedRepetitionService struct {
	minInterval time.Duration
}

func NewSpacedRepetitionService() SpacedRepetitionService {
	return &spacedRepetitionService{
		minInterval: time.Hour * 24,
	}
}

func (s *spacedRepetitionService) CalculateNextReview(word *domain.Word, isCorrect bool) (*domain.ReviewResult, error) {
	quality := s.calculateQuality(isCorrect, word.Difficulty)

	result := &domain.ReviewResult{
		WordID:    word.ID,
		IsCorrect: isCorrect,
		Quality:   quality,
	}

	if isCorrect {
		newEaseFactor := s.CalculateEaseFactor(word, quality)

		if word.ReviewCount == 0 {
			result.NextInterval = 1 * s.minInterval
		} else if word.ReviewCount == 1 {
			result.NextInterval = 3 * s.minInterval
		} else {
			previousInterval := float64(s.getPreviousInterval(word))
			result.NextInterval = time.Duration(previousInterval * newEaseFactor)
		}

		result.NewDifficulty = newEaseFactor
	} else {
		result.NewDifficulty = math.Max(1.3, word.Difficulty-0.2)
		result.NextInterval = s.minInterval
	}

	return result, nil
}

func (s *spacedRepetitionService) GetWordsForReview(words []*domain.Word) []*domain.Word {
	var dueWords []*domain.Word

	for _, word := range words {
		if word.IsDueForReview() {
			dueWords = append(dueWords, word)
		}
	}
	return dueWords
}

func (s *spacedRepetitionService) CalculateEaseFactor(word *domain.Word, quality int) float64 {
	easeFactor := word.Difficulty
	easeFactor = easeFactor + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))

	if easeFactor < 1.3 {
		easeFactor = 1.3
	}

	return easeFactor
}

func (s *spacedRepetitionService) calculateQuality(isCorrect bool, difficulty float64) int {
	if !isCorrect {
		return 0
	}

	if difficulty < 2.0 {
		return 5
	} else if difficulty < 2.5 {
		return 4
	} else if difficulty < 3.0 {
		return 3
	} else if difficulty < 3.5 {
		return 2
	} else {
		return 1
	}
}

func (s *spacedRepetitionService) getPreviousInterval(word *domain.Word) float64 {
	if word.ReviewCount <= 1 {
		return 1.0
	}

	intervals := []float64{1.0, 3.0, 7.0, 14.0, 30.0}
	if word.ReviewCount-2 < len(intervals) {
		return intervals[word.ReviewCount-2]
	}

	return intervals[len(intervals)-1] * math.Pow(word.Difficulty, float64(word.ReviewCount-len(intervals)-1))
}
