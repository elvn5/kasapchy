// Package usecase implements the bot business logic, independent of Telegram.
package usecase

import (
	"context"

	"github.com/your-org/app/bot/internal/domain"
)

// UserRegistry registers Telegram users in the users service.
type UserRegistry interface {
	Register(ctx context.Context, user domain.TelegramUser) error
}

// Onboarding handles a user's first contact with the bot.
type Onboarding struct {
	registry UserRegistry
}

// NewOnboarding creates an Onboarding use case.
func NewOnboarding(registry UserRegistry) *Onboarding {
	return &Onboarding{registry: registry}
}

// Start registers (or refreshes) the user and returns the greeting to send back.
func (o *Onboarding) Start(ctx context.Context, user domain.TelegramUser) (string, error) {
	greeting := "Привет, " + user.DisplayName() + "! 👋\n\n" +
		"Это стартовый бот шаблона. Добавьте свои команды в services/bot/internal/adapter/telegram.\n\n" +
		"/help — список команд"
	if err := o.registry.Register(ctx, user); err != nil {
		return greeting, err
	}
	return greeting, nil
}

// Help returns the list of supported commands.
func (o *Onboarding) Help() string {
	return "Доступные команды:\n/start — начать\n/help — помощь"
}
