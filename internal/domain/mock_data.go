package domain

import "time"

// GetMockWords возвращает тестовые слова для демонстрации
func GetMockWords(userID int64) []*Word {
	return []*Word{
		{
			ID:             1,
			UserID:         userID,
			Original:       "hello",
			Translation:    "привет",
			Language:       "en",
			PartOfSpeech:   "noun",
			Difficulty:     2.5,
			CorrectAnswers: 3,
			NextReview:     time.Now().Add(-time.Hour), // Просрочено для тестирования
		},
		{
			ID:             2,
			UserID:         userID,
			Original:       "book",
			Translation:    "книга",
			Language:       "en",
			PartOfSpeech:   "noun",
			Difficulty:     2.8,
			CorrectAnswers: 1,
			NextReview:     time.Now().Add(-30 * time.Minute),
		},
	}
}
