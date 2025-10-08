package domain

import "time"

type UserStats struct {
	UserID         int64     `json:"user_id"`
	TotalWords     int       `json:"total_words"`
	LearnedWords   int       `json:"learned_words"`
	TotalReviews   int       `json:"total_reviews"`
	TotalCorrect   int       `json:"total_correct"`
	StreakDays     int       `json:"streak_days"`
	MaxStreakDays  int       `json:"max_streak_days"`
	LastReviewDate time.Time `json:"last_review_date"`
	TotalTime      int64     `json:"total_time"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func NewUserStats(userID int64) *UserStats {
	now := time.Now()
	return &UserStats{
		UserID:        userID,
		TotalWords:    0,
		LearnedWords:  0,
		TotalReviews:  0,
		TotalCorrect:  0,
		StreakDays:    0,
		MaxStreakDays: 0,
		TotalTime:     0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (us *UserStats) AddReview(isCorrect bool, duration time.Duration) {
	us.TotalReviews++
	if isCorrect {
		us.TotalCorrect++
	}
	us.TotalTime += int64(duration.Seconds())
	us.LastReviewDate = time.Now()
	us.UpdatedAt = time.Now()

	// Обновляем серию дней
	us.updateStreak()
}

func (us *UserStats) UpdateWordCount(total, learned int) {
	us.TotalWords = total
	us.LearnedWords = learned
	us.UpdatedAt = time.Now()
}

func (us *UserStats) GetAccuracy() float64 {
	if us.TotalReviews == 0 {
		return 0
	}
	return float64(us.TotalCorrect) / float64(us.TotalReviews) * 100
}

func (us *UserStats) GetAverageTime() float64 {
	if us.TotalReviews == 0 {
		return 0
	}
	return float64(us.TotalTime) / float64(us.TotalReviews)
}

func (us *UserStats) updateStreak() {
	now := time.Now()

	if us.LastReviewDate.IsZero() ||
		isSameDay(us.LastReviewDate, now) ||
		isYesterday(us.LastReviewDate, now) {

		if isYesterday(us.LastReviewDate, now) {
			us.StreakDays++
		} else if us.LastReviewDate.IsZero() || !isSameDay(us.LastReviewDate, now) {
			us.StreakDays = 1
		}

		if us.StreakDays > us.MaxStreakDays {
			us.MaxStreakDays = us.StreakDays
		}
	} else {
		us.StreakDays = 0
	}
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func isYesterday(t1, t2 time.Time) bool {
	yesterday := t2.AddDate(0, 0, -1)
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := yesterday.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
