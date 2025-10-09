package service

import (
	"ivanSaichkin/language-bot/internal/repository"
)

type ServiceContainer struct {
	UserService       UserService
	WordService       WordService
	ReviewService     ReviewService
	StatsService      StatsService
	SessionService    SessionService
	RepetitionService SpacedRepetitionService
}

func NewServiceContainer(
	userRepo repository.UserRepository,
	wordRepo repository.WordRepository,
	statsRepo repository.StatsRepository,
	sessionRepo repository.SessionRepository,
) *ServiceContainer {
	// Создаем сервис повторений
	repetitionService := NewSpacedRepetitionService()

	// Создаем основные сервисы
	userService := NewUserService(userRepo, wordRepo, statsRepo)
	wordService := NewWordService(wordRepo, statsRepo)
	statsService := NewStatsService(userRepo, wordRepo, statsRepo)
	reviewService := NewReviewService(wordRepo, statsRepo, repetitionService)
	sessionService := NewSessionService(sessionRepo)

	return &ServiceContainer{
		UserService:       userService,
		WordService:       wordService,
		ReviewService:     reviewService,
		StatsService:      statsService,
		SessionService:    sessionService,
		RepetitionService: repetitionService,
	}
}
