// Package usecase implements the users service business logic.
package usecase

import (
	"context"

	"github.com/your-org/app/users/internal/domain"
)

// UserRepository persists users.
type UserRepository interface {
	List(ctx context.Context) ([]domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	UpsertByTelegramID(ctx context.Context, profile domain.TelegramProfile) (domain.User, error)
}

// UserService contains the user use cases.
type UserService struct {
	repo UserRepository
}

// NewUserService creates a UserService.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// List returns all users, newest first.
func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.List(ctx)
}

// Get returns a user by ID or domain.ErrUserNotFound.
func (s *UserService) Get(ctx context.Context, id string) (domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// RegisterFromTelegram creates the user on first contact and refreshes the
// Telegram profile fields on subsequent ones.
func (s *UserService) RegisterFromTelegram(
	ctx context.Context,
	profile domain.TelegramProfile,
) (domain.User, error) {
	return s.repo.UpsertByTelegramID(ctx, profile)
}
