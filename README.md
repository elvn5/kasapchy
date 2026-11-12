# Fullstack App Template

**Репозиторий-шаблон — чистый лист для новых проектов.** В нём нет никакой
предметной логики: только инфраструктура, на которой можно сразу строить продукт —
Go-микросервисы, админка и публичный сайт на Nuxt, Telegram-бот, Docker и CI/CD.

Создайте новый репозиторий кнопкой **Use this template** на GitHub и начинайте
писать свою бизнес-логику.

## Что внутри

| Каталог | Стек | Назначение |
| --- | --- | --- |
| [`services/gateway`](services/gateway) | Go + Gin | API Gateway — единственная публичная точка входа: CORS, проверка сессии, проксирование в сервисы |
| [`services/auth`](services/auth) | Go + Gin | Вход в админку (пароль → JWT в httpOnly-куке), `/auth/me`, logout |
| [`services/users`](services/users) | Go + Gin + pgx | Пользователи, своя БД `users`, миграции goose |
| [`services/bot`](services/bot) | Go + go-telegram/bot | Telegram-бот: `/start` регистрирует пользователя через `users` |
| [`services/pkg`](services/pkg) | Go | Общий код: конфиг, логгер (slog), Gin-сервер, JWT, Postgres |
| [`apps/admin`](apps/admin) | Nuxt 4 + @nuxt/ui | Админ-панель: логин, главная, список пользователей |
| [`apps/client`](apps/client) | Nuxt 4 + i18n | Публичный сайт-заглушка (ru / ky / en), SEO, sitemap |
| [`packages/theme`](packages/theme) | Tailwind v4 | Общая цветовая схема фронтендов |

`auth`, `users` и `bot` — это стартовые примеры: переименуйте, расширьте или
удалите их под свой домен.

## Архитектура

```
                 ┌──────────────┐   ┌──────────────┐
   браузер ────▶ │ apps/client  │   │  apps/admin  │
                 └──────┬───────┘   └──────┬───────┘
                        │  HTTP (cookie)   │
                        ▼                  ▼
                 ┌─────────────────────────────────┐
                 │   services/gateway  (Gin, :3001) │  CORS, проверка JWT
                 └───────┬───────────────┬─────────┘
               /auth/*   │               │  /users/* (только admin)
                         ▼               ▼
              ┌────────────────┐  ┌────────────────┐     ┌──────────────┐
              │ services/auth  │  │ services/users │ ◀── │ services/bot │ ◀── Telegram
              │     :8081      │  │     :8082      │     └──────────────┘
              └────────────────┘  └───────┬────────┘   POST /internal/users/telegram
                                          ▼
                                  ┌────────────────┐
                                  │  PostgreSQL    │  база на сервис (users, …)
                                  └────────────────┘
```

- **Микросервисы.** Каждый сервис — отдельный Go-модуль и отдельный Docker-образ,
  владеет своими данными (database-per-service на общем Postgres-сервере) и общается
  с другими по HTTP. Наружу публикуется только gateway.
- **Clean Architecture внутри сервиса:**
  ```
  services/<name>/
    cmd/<name>/main.go          сборка зависимостей и запуск (frameworks & drivers)
    internal/domain/            сущности и доменные ошибки — без зависимостей
    internal/usecase/           бизнес-логика; зависит только от интерфейсов
    internal/adapter/httpapi/   HTTP-хендлеры (Gin)
    internal/adapter/repository/ реализация хранилища (pgx)
    internal/config/            настройки из env
    migrations/                 SQL-миграции (goose), вшиваются в бинарник
  ```
- **Внутренние маршруты** (`/internal/*`) gateway не проксирует — они доступны
  только внутри docker-сети (например, для бота).

## Требования

- Go 1.27+
- Node.js 24+ и pnpm 9 (`corepack enable`)
- Docker + Docker Compose

## Быстрый старт

```bash
corepack enable
pnpm install
cp .env.example .env
```

Заполните в `.env` как минимум `JWT_SECRET` (≥ 16 символов), `ADMIN_PASSWORD`,
`POSTGRES_PASSWORD` и `BOT_TOKEN` (от [@BotFather](https://t.me/BotFather)).

### Всё в Docker

```bash
docker compose up -d --build
```

- сайт — http://localhost:3002
- админка — http://localhost:3000 (пароль — `ADMIN_PASSWORD`)
- API (gateway) — http://localhost:3001, health-check — `/healthz`

Режим разработки с hot-reload (air для Go, Nuxt dev-server для фронтендов):

```bash
docker compose -f docker-compose.dev.yml up --build
```

### Локально без Docker

Поднимите только Postgres (`docker compose up -d postgres`), затем в отдельных
терминалах:

```bash
make dev-users      # :8082, применит миграции
make dev-auth       # :8081
make dev-gateway    # :3001
make dev-bot
pnpm dev:admin      # :3000
pnpm dev:client     # :3002
```

Без `make`: `cd services && go run ./users/cmd/users` и т.д.

## Команды

| Команда | Описание |
| --- | --- |
| `make help` | Список Go-команд |
| `make dev-<service>` | Запустить Go-сервис локально |
| `make test` / `make vet` / `make fmt` | Тесты, go vet, gofmt |
| `make lint` | gofmt + go vet + Biome |
| `make build` | Собрать все Go-сервисы в `services/tmp/` |
| `make tidy` | `go mod tidy` во всех модулях |
| `pnpm dev:admin` / `pnpm dev:client` | Nuxt в режиме разработки |
| `pnpm lint` / `pnpm lint:fix` | Biome |
| `pnpm typecheck` / `pnpm build` | Проверка типов / сборка фронтендов |

## Структура

```
services/            Go-микросервисы (go.work объединяет модули для IDE и локальной сборки)
  pkg/               общий модуль
  gateway/ auth/ users/ bot/
  Dockerfile         общий prod-Dockerfile (--build-arg SERVICE=<name>)
  Dockerfile.dev     образ с air для hot-reload
apps/
  admin/             Nuxt-админка
  client/            Nuxt-сайт
packages/
  theme/             общая тема Tailwind / Nuxt UI
deploy/
  postgres/          init-скрипт: база на каждый сервис
  nginx/             пример reverse proxy
.github/workflows/   ci.yml (проверки), deploy.yml (сборка образов и деплой)
```

## Как начать новый проект из шаблона

1. **Use this template** → новый репозиторий.
2. Замените модульный путь Go `github.com/your-org/app` на свой:
   ```bash
   grep -rl 'github.com/your-org/app' services | xargs sed -i '' 's#github.com/your-org/app#github.com/<you>/<project>#g'
   ```
   (на Linux — `sed -i` без `''`).
3. Переименуйте проект в [package.json](package.json), `MyApp` / `App Admin` во
   фронтендах и тексты в `apps/client/i18n/locales/*.json`; цвета — в
   [packages/theme/theme.css](packages/theme/theme.css).
4. Скопируйте `.env.example` в `.env` и заполните секреты — никогда не коммитьте `.env`.
5. Добавляйте свою предметную область (см. ниже).

### Добавить новый микросервис

1. Скопируйте `services/users` в `services/<name>` и поменяйте `module` в `go.mod`.
2. Добавьте модуль в `services/go.work` и `SERVICES` в [Makefile](Makefile).
3. Если нужна БД — добавьте имя в `POSTGRES_MULTIPLE_DATABASES`.
4. Добавьте сервис в `docker-compose*.yml`, в матрицу `build` в
   `.github/workflows/deploy.yml` и в `.github/dependabot.yml`.
5. Пропишите публичные маршруты в `services/gateway/internal/router/router.go`.

## Деплой

GitHub Actions собирает образы в GHCR и выкатывает их по SSH: push в `main` →
staging, тег `vX.Y.Z` → production. Подробности — в [DEPLOYMENT.md](DEPLOYMENT.md).

## Соглашения

См. [CLAUDE.md](CLAUDE.md) и [CONTRIBUTING.md](CONTRIBUTING.md).

## Лицензия

[MIT](LICENSE)
