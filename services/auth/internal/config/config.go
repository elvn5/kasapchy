// Package config describes the auth service settings.
package config

import (
	"errors"
	"time"

	"github.com/your-org/app/pkg/config"
)

// Config is loaded from the environment (see .env.example).
type Config struct {
	config.Base
	HTTPPort      int           `env:"HTTP_PORT" envDefault:"8081"`
	JWTSecret     string        `env:"JWT_SECRET,required"`
	JWTExpiresIn  time.Duration `env:"JWT_EXPIRES_IN" envDefault:"12h"`
	AdminPassword string        `env:"ADMIN_PASSWORD,required"`
}

// Validate implements config.Validator.
func (c Config) Validate() error {
	if len(c.JWTSecret) < 16 {
		return errors.New("JWT_SECRET must be at least 16 characters")
	}
	return nil
}
