// Package domain holds the auth service entities and business errors.
package domain

import (
	"errors"
	"time"
)

var (
	// ErrInvalidCredentials is returned when the supplied password does not match.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUnauthenticated is returned when a session token is missing or invalid.
	ErrUnauthenticated = errors.New("unauthenticated")
)

// Token is a signed session token handed to the client.
type Token struct {
	Value     string
	ExpiresAt time.Time
}

// Session describes an authenticated principal.
type Session struct {
	Role string
}
