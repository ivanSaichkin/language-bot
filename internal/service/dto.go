package service

import "time"

type ReviewAnswerResult struct {
	IsCorrect       bool
	CorrectAnswer   string
	NextInterval    time.Duration
	OriginalWord    string
	SessionProgress *SessionProgress
}

type SessionProgress struct {
	Current    int
	Total      int
	Correct    int
	Accuracy   float64
	IsComplete bool
}

type WordProgress struct {
	TotalWords    int
	LearnedWords  int
	DueWords      int
	Progress      float64
	TodayReviewed int
}

type UserStatsRanking struct {
	UserID       int64
	Username     string
	FirstName    string
	TotalWords   int
	LearnedWords int
	StreakDays   int
	Rank         int
}

type StreakInfo struct {
	CurrentStreak    int
	MaxStreak        int
	LastReviewDate   time.Time
	IsTodayCompleted bool
}

type DailyProgress struct {
	DailyGoal      int
	TodayReviewed  int
	Remaining      int
	CompletionRate float64
	IsGoalAchieved bool
}
