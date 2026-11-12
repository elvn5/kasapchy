// Package telegram maps Telegram updates onto bot use cases.
package telegram

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/your-org/app/bot/internal/domain"
	"github.com/your-org/app/bot/internal/usecase"
)

// Handlers holds the Telegram update handlers.
type Handlers struct {
	onboarding *usecase.Onboarding
	log        *slog.Logger
}

// NewHandlers creates Handlers.
func NewHandlers(onboarding *usecase.Onboarding, log *slog.Logger) *Handlers {
	return &Handlers{onboarding: onboarding, log: log}
}

// Register attaches command handlers to b.
func (h *Handlers) Register(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand, h.start)
	b.RegisterHandler(bot.HandlerTypeMessageText, "help", bot.MatchTypeCommand, h.help)
}

// Default handles every update that no command handler matched.
func (h *Handlers) Default(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	h.reply(ctx, b, update.Message.Chat.ID, h.onboarding.Help())
}

func (h *Handlers) start(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	if msg == nil || msg.From == nil {
		return
	}

	greeting, err := h.onboarding.Start(ctx, domain.TelegramUser{
		ID:        msg.From.ID,
		Username:  msg.From.Username,
		FirstName: msg.From.FirstName,
		LastName:  msg.From.LastName,
	})
	if err != nil {
		h.log.ErrorContext(ctx, "register user", slog.Int64("telegram_id", msg.From.ID), slog.Any("error", err))
	}
	h.reply(ctx, b, msg.Chat.ID, greeting)
}

func (h *Handlers) help(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	h.reply(ctx, b, update.Message.Chat.ID, h.onboarding.Help())
}

func (h *Handlers) reply(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text}); err != nil {
		h.log.ErrorContext(ctx, "send message", slog.Int64("chat_id", chatID), slog.Any("error", err))
	}
}
