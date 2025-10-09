package service

import (
	"ivanSaichkin/language-bot/internal/repository"
)

type ServiceContainer struct {
	UserService       UserService
	WordService       WordService
	ReviewService     ReviewService
	StatsService      StatsService
	RepetitionService SpacedRepetitionService
}

func NewServiceContainer(
	userRepo repository.UserRepository,
	wordRepo repository.WordRepository,
	statsRepo repository.StatsRepository,
) *ServiceContainer {
	repetitionService := NewSpacedRepetitionService()

	userService := NewUserService(userRepo, wordRepo, statsRepo)
	wordService := NewWordService(wordRepo, statsRepo)
	statsService := NewStatsService(userRepo, wordRepo, statsRepo)
	reviewService := NewReviewService(wordRepo, statsRepo, repetitionService)

	return &ServiceContainer{
		UserService:       userService,
		WordService:       wordService,
		ReviewService:     reviewService,
		StatsService:      statsService,
		RepetitionService: repetitionService,
	}
}
