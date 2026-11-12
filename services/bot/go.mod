module github.com/your-org/app/bot

go 1.27.1

require (
	github.com/go-telegram/bot v1.27.0
	github.com/your-org/app/pkg v0.0.0
)

require (
	github.com/caarlos0/env/v11 v11.4.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
)

replace github.com/your-org/app/pkg => ../pkg
