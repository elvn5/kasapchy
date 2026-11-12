# CLAUDE.md

## General Rules & Code Style
- **Imports:** ALWAYS use absolute imports (`@/` or `~/`) instead of relative paths (`../`, `./`). In Go, import by full module path (`github.com/your-org/app/...`).
- **Architecture:**
  - **Backend:** Go microservices in `services/`, each following Clean Architecture (`internal/domain` -> `internal/usecase` -> `internal/adapter/*` -> `cmd/<name>`). Keep business logic strictly independent of Gin/pgx/Telegram. Services own their data (database-per-service) and talk over HTTP; only `services/gateway` is public. Shared helpers live in `services/pkg`.
  - **Frontend:** Feature-Sliced Design (FSD) (`app/`, `pages/`, `widgets/`, `features/`, `entities/`, `shared/`). Respect strict cross-layer import rules.
- **Formatting & Linting:** Use **Biome** for the frontend (`npx @biomejs/biome check --write`) and `gofmt` + `go vet` for Go.
- **Validation:** Use Gin binding tags / **go-playground/validator** (backend) and **VeeValidate** + Zod (frontend) for input/data validation.
- **Logging:** Use `log/slog` (via `services/pkg/logger`) in Go. Avoid `fmt.Println` / `console.log`.
- **Environment:** Manage environment variables via the root `.env` (loaded in Go by `services/pkg/config` with godotenv + caarlos0/env).
- **TypeScript:** Strict mode enabled. Always specify return types for public functions/methods.
- **Migrations:** SQL migrations per service in `services/<name>/migrations/` (goose format, embedded and applied on startup).

## Tech Stack
- **Backend:** Go 1.27, Gin (HTTP), pgx + goose (PostgreSQL), go-telegram/bot, golang-jwt, Docker
- **Frontend:** Nuxt 3, Pinia (+ persistedstate), @nuxt/ui, Tailwind CSS, Nuxt Icon, @nuxtjs/seo, @nuxtjs/i18n, @vueuse/nuxt, @tanstack/vue-query
- **Linter/Formatter:** Biome

## Commands
- **Dev (Go):** `make dev-gateway` / `dev-auth` / `dev-users` / `dev-bot` (or `cd services && go run ./<name>/cmd/<name>`)
- **Dev (Nuxt):** `pnpm dev:admin` / `pnpm dev:client`
- **Test (Go):** `make test` (or `cd services && go test ./pkg/... ./gateway/... ./auth/... ./users/... ./bot/...`)
- **Build:** `make build` (Go) / `pnpm build` (Nuxt)
- **Lint/Format:** `make lint` / `make fmt` / `npx @biomejs/biome check --write .`
- **Type Check:** `pnpm typecheck`
- **Docker:** `docker compose up -d --build` (hot-reload: `docker compose -f docker-compose.dev.yml up --build`)
