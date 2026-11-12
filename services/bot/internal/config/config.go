// Package config describes the bot settings.
package config

import "github.com/your-org/app/pkg/config"

// Config is loaded from the environment (see .env.example).
type Config struct {
	config.Base
	BotToken        string `env:"BOT_TOKEN,required"`
	UsersServiceURL string `env:"USERS_SERVICE_URL" envDefault:"http://localhost:8082"`
}
