package domain

import (
	"fmt"
	"time"
)

type ReviewSession struct {
	ID             string    `json:"id"`
	UserID         int64     `json:"user_id"`
	Words          []*Word   `json:"words"`
	CurrentIndex   int       `json:"current_index"`
	CorrectAnswers int       `json:"correct_answers"`
	TotalQuestions int       `json:"total_questions"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	IsCompleted    bool      `json:"is_completed"`
}

func NewReviewSession(userID int64, words []*Word) *ReviewSession {
	if len(words) > 20 {
		words = words[:20]
	}

	return &ReviewSession{
		ID:             generateSessionID(userID),
		UserID:         userID,
		Words:          words,
		CurrentIndex:   0,
		CorrectAnswers: 0,
		TotalQuestions: len(words),
		StartTime:      time.Now(),
		IsCompleted:    false,
	}
}

func (rs *ReviewSession) Complete() {
	rs.IsCompleted = true
	rs.EndTime = time.Now()
}

func (rs *ReviewSession) GetCurrentWord() *Word {
	if rs.CurrentIndex >= len(rs.Words) {
		return nil
	}

	return rs.Words[rs.CurrentIndex]
}

func (rs *ReviewSession) Answer(isCorrect bool) {
	if rs.IsCompleted {
		return
	}

	if isCorrect {
		rs.CorrectAnswers++
	}

	rs.CurrentIndex++
	if rs.CurrentIndex >= len(rs.Words) {
		rs.Complete()
	}
}

func (rs *ReviewSession) GetProgress() (current int, total int) {
	return rs.CurrentIndex + 1, len(rs.Words)
}

func (rs *ReviewSession) GetAccuracy() float64 {
	if rs.CurrentIndex == 0 {
		return 0
	}
	return float64(rs.CorrectAnswers) / float64(rs.CurrentIndex) * 100
}

func (rs *ReviewSession) GetDuration() time.Duration {
	endTime := rs.EndTime
	if endTime.IsZero() {
		endTime = time.Now()
	}
	return endTime.Sub(rs.StartTime)
}

func generateSessionID(userID int64) string {
	return fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
}
