// Package config describes the API gateway settings.
package config

import (
	"errors"

	"github.com/your-org/app/pkg/config"
)

// Config is loaded from the environment (see .env.example).
type Config struct {
	config.Base
	HTTPPort        int      `env:"HTTP_PORT" envDefault:"3001"`
	CORSOrigins     []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://localhost:3002"`
	JWTSecret       string   `env:"JWT_SECRET,required"`
	AuthServiceURL  string   `env:"AUTH_SERVICE_URL" envDefault:"http://localhost:8081"`
	UsersServiceURL string   `env:"USERS_SERVICE_URL" envDefault:"http://localhost:8082"`
}

// Validate implements config.Validator.
func (c Config) Validate() error {
	if len(c.JWTSecret) < 16 {
		return errors.New("JWT_SECRET must be at least 16 characters")
	}
	return nil
}
