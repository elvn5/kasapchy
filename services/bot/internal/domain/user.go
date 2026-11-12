// Package domain holds the bot entities.
package domain

// TelegramUser is the sender of a Telegram update.
type TelegramUser struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
}

// DisplayName returns the best human-readable name available.
func (u TelegramUser) DisplayName() string {
	switch {
	case u.FirstName != "":
		return u.FirstName
	case u.Username != "":
		return "@" + u.Username
	default:
		return "гость"
	}
}
