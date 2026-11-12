// Package config describes the users service settings.
package config

import (
	"github.com/your-org/app/pkg/config"
	"github.com/your-org/app/pkg/postgres"
)

// Config is loaded from the environment (see .env.example).
type Config struct {
	config.Base
	HTTPPort int    `env:"HTTP_PORT" envDefault:"8082"`
	DBName   string `env:"USERS_DB_NAME" envDefault:"users"`
	Postgres postgres.Config
}
