# Marketplace API (Go / Gin / PostgreSQL)

Отдельный Go-модуль, независимый от шаблонной админской авторизации. Существующие
Go-сервисы не используются для мобильных сессий. Команды этого модуля выполняются
с `GOWORK=off`: корневой шаблон закрепляет другую версию Go.

## Локально

```sh
docker compose -f compose.local.yml up -d --build
# API: http://127.0.0.1:18080
# Почтовый ящик Mailpit: http://127.0.0.1:18025
```

В development письма доставляются только в локальный Mailpit. Production требует
SMTP с STARTTLS. В базе нет демонстрационных пользователей или товаров: зарегистрируйте
продавца, добавьте товар, затем зарегистрируйте покупателя.

```sh
GOWORK=off go test ./...
MARKET_TEST_DATABASE_URL='postgres://marketplace:local-development-only@127.0.0.1:55438/marketplace?sslmode=disable' \
  GOWORK=off go test -race ./...
```

Интеграционные тесты создают отдельные временные схемы и удаляют только их.

## API v1

- `POST /auth/email/request` — `{email, role}` (`Buyer` / `Seller`), возвращает `challengeId`.
- `POST /auth/email/verify` — `{challengeId, code}`, возвращает `{token, user}`.
- `POST /auth/oidc/challenge` — `{provider, role, proofHash?}`, возвращает challenge и nonce.
- `POST /auth/oidc/verify` — `{provider, challengeId, idToken}`. Проверка issuer, подписи,
  audience, срока, email_verified и одноразового nonce; токены не логируются.
- `POST /auth/apple/callback` — зарегистрированный HTTPS return URL Apple, `form_post`.
- `POST /auth/apple/poll` — `{challengeId, proof}`, получение сессии после callback;
  202 пока подтверждения нет. Proof — случайная строка клиента, в БД только SHA-256.
- `GET /products` — каталог, до 500 последних записей.
- `GET/PATCH/DELETE /me` — профиль / `{name, phone}` / удаление аккаунта.
- `POST /auth/logout` — отзыв текущей сессии.
- `POST /products` — только продавец: `{name, category, price, stock, description}`.
- `GET /orders` — только свои покупки либо заказы на свои товары, до 500 последних.
- `POST /orders` — `{productId, quantity, delivery, address, requestId}`. Цена и сумма
  рассчитываются сервером; остаток блокируется транзакцией. Повторный requestId не
  создаёт второй заказ; изменённое содержимое с тем же ключом получает 409.
- `POST /orders/:id/advance` — `{expectedStatus}`; только владелец товара.

Bearer-сессии действуют 30 дней, в базе только SHA-256 токена. OTP: криптографически
случайные 6 цифр, HMAC, 10 минут, 5 попыток, cooldown 60 секунд, квоты на email/IP.
Роль уже созданного аккаунта не меняется через повторный вход.

Удаление аккаунта требует входа в последние 15 минут; удаляет профиль, сессии,
товары и связанные заказы. Это текущая политика удаления: до публичного запуска
владелец должен определить допустимое хранение документов о сделках. Сервис не
проводит платежи. Доставка пока фиксирована (150 сом) и требует бизнес-подтверждения.

## Сервер и GitHub Actions

Workflow: `.github/workflows/marketplace-deploy.yml`. Деплой в `/opt/kasapchy`
через отдельный forced-command SSH-ключ; ключ не даёт произвольный SSH shell.
GitHub Secrets: `MARKET_DEPLOY_KEY`, `MARKET_KNOWN_HOSTS`.
GitHub Variable: `MARKET_DEPLOY_HOST`. Пароли БД/OTP находятся только в
`/opt/kasapchy/.env` (0600), не в git или GitHub логах.

Текущий bootstrap — staging, API слушает только `127.0.0.1:5700`; Mailpit — только
`127.0.0.1:5701`. Он не отправляет реальные письма и не опубликован в интернете.
Существующие nginx-конфигурации и чужие контейнеры не меняются.

После выбора домена:
1. DNS и HTTPS reverse proxy на `127.0.0.1:5700`; не публиковать БД/Mailpit наружу.
2. Указать реальный SMTP, sender, SPF/DKIM/DMARC; переключить `APP_ENV=production`.
3. Настроить Google server client ID и Android client c SHA-1 сертификата Play
   App Signing. Зарегистрировать Apple Services ID и return URL
   `https://<api-domain>/v1/auth/apple/callback`.
4. Задать `MARKET_GOOGLE_CLIENT_ID`, `MARKET_APPLE_CLIENT_ID`, перезапустить API.
5. Настроить резервное копирование PostgreSQL, мониторинг, проверку восстановления,
   публичные privacy/deletion страницы и контакт поддержки.

Для проверки через SSH: `ssh -L 18080:127.0.0.1:5700 -L 18025:127.0.0.1:5701 root@<server>`.
До подключения SMTP не использовать этот сервер как публичный production.

## Домены (12 ноября 2026)

DNS `kasapchy.api.ermekten.com` и `test-kasapchy.api.ermekten.com` направлен на
сервер. Установлен отдельный Nginx vhost из `deploy/nginx-bootstrap.conf`,
сертификат Let's Encrypt покрывает оба имени. HTTP перенаправляется на HTTPS.
Автопродление использует существующий certbot.timer и deploy-hook
`deploy/renew-nginx.sh` для проверки/reload Nginx после обновления сертификата.

Основной домен пока возвращает 503. На тестовом открыт только `/healthz` →
`127.0.0.1:5700/healthz`; остальные маршруты возвращают 503. Это намеренное
ограничение до подключения SMTP и настройки production, не неисправность DNS.
Mailpit и база по-прежнему недоступны извне. Это ещё не два независимых окружения:
есть один staging stack, production-домен зарезервирован maintenance-конфигурацией.
Существующий `api.ermekten.com` обслуживает прежний сервис без изменений.

Nginx и certbot — конфигурация хоста, не часть Docker deploy workflow: перед
установкой vhost нужен выпущенный сертификат, затем `nginx -t` и reload.
Открывать `/v1` в production следует после выполнения списка настроек выше и
настройки доверенного reverse proxy; сейчас SMTP доставляет письма в Mailpit.

## Резервное копирование и контроль

Установлены ежедневные encrypted backups (14 дней), backup перед деплоем,
проверка `/readyz` с PostgreSQL и systemd monitoring каждые 5 минут.
Восстановление первой копии проверено на отдельном локальном PostgreSQL.
Процедуры, пути ключей, ограничения offsite/уведомлений и откат:
[deploy/OPERATIONS.md](deploy/OPERATIONS.md).


### First-release password authentication

`POST /v1/auth/password/register` and `/v1/auth/password/login` take
`email`, `password` (12–72 UTF-8 bytes), and `role` (`Buyer`/`Seller`).
They return a session without sending email. Email ownership is not verified.
Duplicate registration never overwrites a password; verified identities are not
silently linked to password accounts. bcrypt cost 12; 10 attempts/email and 30/IP
per 15 minutes. Password reset is not available in this release.

Email OTP is disabled by default in the server process. Set
`MARKET_EMAIL_OTP_ENABLED=true` only with the intended mail configuration.
The local compose environment explicitly enables Mailpit for legacy OTP tests.
Deletion on the web uses the same password, without storing it in browser storage.
