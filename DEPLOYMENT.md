# Деплой

CI/CD построен на GitHub Actions + Docker. Образы всех приложений
(Go-микросервисы `gateway`, `auth`, `users`, `bot` и фронтенды `admin`, `client`)
собираются и публикуются в GitHub Container Registry (GHCR), затем
разворачиваются на сервере по SSH через `docker compose`.

Проверки (gofmt, go vet, go test, Biome, typecheck, build) запускаются
отдельным workflow `.github/workflows/ci.yml` на каждый PR.

## Стратегия окружений

| Событие                     | Окружение   | Тег образа для деплоя |
| --------------------------- | ----------- | --------------------- |
| push / merge в `main`       | staging     | `staging`             |
| push тега `vX.Y.Z`          | production  | `X.Y.Z`               |
| «Run workflow» вручную      | по выбору   | `staging` / выбранное |

Каждый образ дополнительно тегируется неизменяемой версией
(`sha-<commit>` для staging, `X.Y.Z` для production) — для трассировки и
откатов. Workflow: `.github/workflows/deploy.yml`.

## Что нужно настроить в GitHub

Settings → Environments → создать **`staging`** и **`production`**.
Для production полезно включить *Required reviewers* — тогда прод-деплой
будет ждать ручного подтверждения.

### Секреты (для каждого окружения)

| Секрет             | Назначение                                             |
| ------------------ | ------------------------------------------------------ |
| `DEPLOY_SSH_HOST`  | адрес сервера                                          |
| `DEPLOY_SSH_USER`  | SSH-пользователь                                       |
| `DEPLOY_SSH_KEY`   | приватный SSH-ключ (весь PEM-файл)                     |
| `DEPLOY_SSH_PORT`  | порт SSH (необязательно, по умолчанию `22`)            |
| `POSTGRES_PASSWORD`, `JWT_SECRET`, `ADMIN_PASSWORD`, `BOT_TOKEN` | секреты приложений (обязательные) |
| `POSTGRES_USER`, `POSTGRES_DB`, `JWT_EXPIRES_IN`, `CORS_ORIGINS` | необязательные |

`GITHUB_TOKEN` создаётся автоматически и используется и для публикации
образов в GHCR, и для их скачивания на сервере — отдельно ничего не нужно.

### Переменные (Variables)

Передаются фронтендам как build-args и записываются в `.env` на сервере
(Nuxt читает `NUXT_PUBLIC_*` и в рантайме).

| Переменная                 | Назначение                          |
| -------------------------- | ----------------------------------- |
| `NUXT_PUBLIC_API_BASE_URL` | публичный URL API                   |
| `NUXT_PUBLIC_BOT_URL`      | ссылка на Telegram-бота (client)    |

## Что нужно на сервере

1. Установлены Docker и плагин Docker Compose (`docker compose version`).
2. Больше ничего: каталог деплоя (`/opt/app` или `/opt/app-staging`)
   создаётся автоматически, а `.env` (из секретов GitHub),
   `docker-compose.deploy.yml`, `deploy.sh` и `deploy/postgres/init-databases.sh`
   копируются туда при каждом деплое. Для фронтендов важно, чтобы
   `NUXT_PUBLIC_API_BASE_URL` указывал на **публичный** адрес API (gateway).

Портами наружу (на `127.0.0.1`, за nginx) проброшены: API/gateway `3001`,
admin `3000`, client `3002`. Сервисы `auth`, `users`, `bot` и Postgres
доступны только внутри docker-сети.

## Как выкатывать

- **Staging** — просто влить в `main`.
- **Production** — поставить тег и запушить его:
  ```bash
  git tag v1.0.0
  git push origin v1.0.0
  ```
- **Вручную** — вкладка Actions → Deploy → Run workflow → выбрать окружение.

## Откат

Задеплоить нужную версию вручную через Run workflow не выйдет по старому
коммиту напрямую, поэтому для отката перевыпустите предыдущий тег или на
сервере укажите конкретный неизменяемый тег:

```bash
cd /opt/app
IMAGE_PREFIX=ghcr.io/your-org/your-repo IMAGE_TAG=0.9.0 ./deploy.sh
```

## Миграции БД

Каждый сервис с базой (сейчас `users`) хранит SQL-миграции в
`services/<name>/migrations/` (формат goose), вшивает их в бинарник и
применяет при старте. Отдельный шаг миграций в пайплайне не нужен.

База каждого сервиса создаётся скриптом `deploy/postgres/init-databases.sh`
из списка `POSTGRES_MULTIPLE_DATABASES` — но только при **первой**
инициализации тома Postgres. Добавив новый сервис с базой на уже работающем
сервере, создайте базу вручную:

```bash
docker compose -p app -f docker-compose.deploy.yml exec postgres createdb -U app <name>
```

## Локальная сборка

`docker compose up -d --build` собирает все образы из исходников локально
(см. `docker-compose.yml`).

## Два окружения на одном сервере: production + staging (test.*)

Production и staging живут на одном VPS, но полностью изолированы через
разные compose-проекты, каталоги, тома БД и порты. `deploy.yml` сам
подставляет нужные значения по каналу:

| Параметр            | production                | staging                          |
|---------------------|---------------------------|----------------------------------|
| Триггер             | тег `vX.Y.Z`              | push в `main` / dispatch=staging |
| Каталог на сервере  | `/opt/app`                | `/opt/app-staging`               |
| compose-проект      | `app`                     | `app-staging`                    |
| Порты (127.0.0.1)   | api 3001, admin 3000, client 3002 | api 4001, admin 4000, client 4002 |
| Домены              | example.com, admin., api. | test.example.com, test.admin., test.api. |
| Образ (IMAGE_TAG)   | версия `X.Y.Z`            | `staging`                        |

Разные `COMPOSE_PROJECT_NAME` дают отдельные тома (`app_postgres_data`
vs `app-staging_postgres_data`), сети и контейнеры — БД не пересекаются.

### Разовая настройка сервера (для staging)

1. **DNS** — A-записи на IP сервера:
   `test`, `test.admin`, `test.api` → <IP>.

2. **nginx** — скопируйте `deploy/nginx/staging.example.conf` в
   `/etc/nginx/conf.d/app-staging.conf`, замените `example.com` на свой домен,
   затем `nginx -t && systemctl reload nginx`.

3. **TLS** — Let's Encrypt для staging-доменов:
   ```bash
   certbot --nginx -d test.example.com -d test.admin.example.com -d test.api.example.com
   ```

4. **GitHub Environments** (Settings → Environments):
   - **production**: variable `NUXT_PUBLIC_API_BASE_URL=https://api.example.com`,
     secret `CORS_ORIGINS=https://example.com,https://admin.example.com`,
     плюс все app-секреты (JWT_SECRET, ADMIN_PASSWORD, BOT_TOKEN, POSTGRES_PASSWORD, …).
   - **staging**: variable `NUXT_PUBLIC_API_BASE_URL=https://test.api.example.com`,
     secret `CORS_ORIGINS=https://test.example.com,https://test.admin.example.com`,
     плюс свои app-секреты.
   - `DEPLOY_SSH_*` — общие (repo-level), сервер один.
   - ⚠️ **BOT_TOKEN**: один Telegram-бот не может работать в двух стеках
     одновременно (конфликт polling). Заведите ОТДЕЛЬНОГО тест-бота и его
     токен положите в staging-environment (иначе прод-бот и staging-бот
     будут отбивать апдейты друг у друга).
   - `NUXT_PUBLIC_API_BASE_URL` вшивается в client/admin на этапе СБОРКИ,
     поэтому job `build` теперь тоже имеет `environment:` — менять домен без
     пересборки образа нельзя.

### Порядок первого запуска staging
DNS → nginx (http) → push в `main` (соберёт и задеплоит staging-стек,
каталог создастся автоматически) → certbot (после того как домены отвечают
по http) → при необходимости повторный push, чтобы образы пересобрались с
https-доменом API.
