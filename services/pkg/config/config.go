// Package config loads service configuration from environment variables and .env files.
package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// envFiles are probed in order and the first existing one is loaded. This lets a service
// started from services/<name>, services/ or the repo root pick up the root .env file.
// Variables already present in the environment always win over values from the file.
var envFiles = []string{".env", "../.env", "../../.env"}

// Base holds settings shared by every service. Embed it into a service config struct.
type Base struct {
	AppEnv   string `env:"APP_ENV" envDefault:"development"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

// IsProduction reports whether the service runs in production mode.
func (b Base) IsProduction() bool {
	return b.AppEnv == "production"
}

// Validator is implemented by configs that need checks beyond struct tags.
type Validator interface {
	Validate() error
}

// Load reads the first .env file found (if any) and parses environment variables into T.
// If T implements Validator, Validate is called on the result.
func Load[T any]() (T, error) {
	var zero T
	for _, file := range envFiles {
		err := godotenv.Load(file)
		if err == nil {
			break
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return zero, fmt.Errorf("load %s: %w", file, err)
		}
	}

	cfg, err := env.ParseAs[T]()
	if err != nil {
		return zero, fmt.Errorf("parse env: %w", err)
	}

	if v, ok := any(cfg).(Validator); ok {
		if err := v.Validate(); err != nil {
			return zero, fmt.Errorf("invalid config: %w", err)
		}
	}
	return cfg, nil
}
