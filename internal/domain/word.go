package domain

import (
	"ivanSaichkin/language-bot/internal/constants"
	"time"
)

type Word struct {
	ID             int       `json:"id"`
	UserID         int64     `json:"user_id"`
	Original       string    `json:"original"`
	Translation    string    `json:"translation"`
	Language       string    `json:"language"`
	PartOfSpeech   string    `json:"part_of_speech"`
	Example        string    `json:"example"`
	Difficulty     float64   `json:"difficulty"`
	NextReview     time.Time `json:"next_review"`
	ReviewCount    int       `json:"review_count"`
	CorrectAnswers int       `json:"correct_answers"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func NewWord(userID int64, original, translation, language string) *Word {
	now := time.Now()

	return &Word{
		UserID:         userID,
		Original:       original,
		Translation:    translation,
		Language:       language,
		PartOfSpeech:   constants.PartOfSpeechNoun,
		Difficulty:     2.5,
		NextReview:     now,
		ReviewCount:    0,
		CorrectAnswers: 0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (w *Word) WithPartOfSpeech(pos string) *Word {
	w.PartOfSpeech = pos
	return w
}

func (w *Word) WithExample(example string) *Word {
	w.Example = example
	return w
}

func (w *Word) IsDueForReview() bool {
	return time.Now().After(w.NextReview) || time.Now().Equal(w.NextReview)
}

func (w *Word) MarkReviewed(isCorrect bool, nextReview time.Time, newDifficulty float64) {
	w.ReviewCount++
	if isCorrect {
		w.CorrectAnswers++
	} else {
		w.CorrectAnswers = 0
	}
	w.Difficulty = newDifficulty
	w.NextReview = nextReview
	w.UpdatedAt = time.Now()
}

func (w *Word) IsLearned() bool {
	return w.CorrectAnswers >= 5
}

func (w *Word) GetProgress() float64 {
	if w.IsLearned() {
		return 100.0
	}

	return float64(w.CorrectAnswers) / 5.0 * 100.0
}

func (w *Word) MarkReviewedWithResult(result *ReviewResult, nextReview time.Time) {
	w.ReviewCount++
	if result.IsCorrect {
		w.CorrectAnswers++
	} else {
		w.CorrectAnswers = 0
	}
	w.Difficulty = result.NewDifficulty
	w.NextReview = nextReview
	w.UpdatedAt = time.Now()
}
