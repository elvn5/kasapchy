// Package domain holds the users service entities and business errors.
package domain

import (
	"errors"
	"time"
)

// ErrUserNotFound is returned when a user does not exist.
var ErrUserNotFound = errors.New("user not found")

// User is an end user of the product, identified by their Telegram account.
type User struct {
	ID         string
	TelegramID int64
	Username   *string
	FirstName  *string
	LastName   *string
	Phone      *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TelegramProfile is the public profile Telegram sends with every update.
type TelegramProfile struct {
	TelegramID int64
	Username   *string
	FirstName  *string
	LastName   *string
}
