package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"ivanSaichkin/language-bot/internal/domain"
	"ivanSaichkin/language-bot/internal/repository"
)

type sessionService struct {
	sessionRepo repository.SessionRepository
}

func NewSessionService(sessionRepo repository.SessionRepository) SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *sessionService) CleanupOldSessions(ctx context.Context, olderThan time.Duration) (int, error) {
	cleaned, err := s.sessionRepo.CleanupOldSessions(ctx, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old sessions: %w", err)
	}

	if cleaned > 0 {
		log.Printf("🧹 Cleaned up %d old sessions from database", cleaned)
	}

	return cleaned, nil
}

func (s *sessionService) GetActiveSessionsCount(ctx context.Context) int {
	sessions, err := s.sessionRepo.GetActiveSessions(ctx)
	if err != nil {
		log.Printf("⚠️ Failed to get active sessions: %v", err)
		return 0
	}

	return len(sessions)
}

func (s *sessionService) SaveSession(ctx context.Context, session *domain.ReviewSession) error {
	existing, err := s.sessionRepo.GetByID(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}

	if existing == nil {
		return s.sessionRepo.Create(ctx, session)
	} else {
		return s.sessionRepo.Update(ctx, session)
	}
}

func (s *sessionService) LoadSession(ctx context.Context, sessionID string) (*domain.ReviewSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

func (s *sessionService) LoadUserSessions(ctx context.Context, userID int64) ([]*domain.ReviewSession, error) {
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user sessions: %w", err)
	}

	return sessions, nil
}

func (s *sessionService) DeleteSession(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}
