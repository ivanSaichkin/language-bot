package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"ivanSaichkin/language-bot/internal/domain"
	"ivanSaichkin/language-bot/internal/repository"
)

type userService struct {
	userRepo  repository.UserRepository
	wordRepo  repository.WordRepository
	statsRepo repository.StatsRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	wordRepo repository.WordRepository,
	statsRepo repository.StatsRepository,
) UserService {
	return &userService{
		userRepo:  userRepo,
		wordRepo:  wordRepo,
		statsRepo: statsRepo,
	}
}

func (s *userService) CreateOrUpdateUser(ctx context.Context, user *domain.User) error {
	existingUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if existingUser == nil {
		if err := s.userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		log.Printf("✅ Created new user: %d %s", user.ID, user.Username)
	} else {
		existingUser.Username = user.Username
		existingUser.FirstName = user.FirstName
		existingUser.LastName = user.LastName
		existingUser.LanguageCode = user.LanguageCode
		existingUser.UpdatedAt = time.Now()

		if err := s.userRepo.Update(ctx, existingUser); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		log.Printf("✅ Updated user: %d %s", user.ID, user.Username)
	}

	return nil
}

func (s *userService) GetUser(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found: %d", userID)
	}

	return user, nil
}

func (s *userService) SetUserState(ctx context.Context, userID int64, state string) error {
	if err := s.userRepo.UpdateState(ctx, userID, state); err != nil {
		return fmt.Errorf("failed to set user state: %w", err)
	}

	log.Printf("🔧 User %d state changed to: %s", userID, state)
	return nil
}

func (s *userService) GetUserState(ctx context.Context, userID int64) (string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user state: %w", err)
	}

	if user == nil {
		return "", fmt.Errorf("user not found: %d", userID)
	}

	return string(user.State), nil
}

func (s *userService) UpdateDailyGoal(ctx context.Context, userID int64, goal int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user for goal update: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found: %d", userID)
	}

	user.SetDailyGoal(goal)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update daily goal: %w", err)
	}

	log.Printf("🎯 User %d daily goal updated to: %d", userID, goal)
	return nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	users, err := s.userRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	return users, nil
}

func (s *userService) CleanupOldSessions(ctx context.Context, olderThan time.Duration) (int, error) {
	log.Printf("🧹 Would cleanup sessions older than %v", olderThan)
	return 0, nil
}
