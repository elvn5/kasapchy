# OAuth: подготовка и оставшиеся шаги

12 ноября 2026 создан Google Cloud проект **kasapchy-marketplace**.
[Google Auth Platform](https://console.cloud.google.com/auth/overview?project=kasapchy-marketplace).
Branding не завершён: требуется разрешённый владельцем публичный support email;
личный адрес Google-аккаунта автоматически не выбран. OAuth clients пока не созданы.
Биллинг не подключался, доступ IAM не расширялся.

## Google

- App name: Касапчы. Audience: External, сначала Testing.
- Scopes: только openid, email, profile; доступ к Gmail/Drive/контактам не нужен.
- Создать Web OAuth client для serverClientId. Один и тот же public client ID
  используется Android `kasapchy.googleClientId` и API `MARKET_GOOGLE_CLIENT_ID`.
  В текущем native ID-token flow redirect URI и web client secret не используются.
- Android OAuth client: package `kg.kasapchy.mobile`, SHA-1 сертификата **Play App
  Signing** из Console после настройки приложения. Upload key и Play signing key
  могут отличаться. Не заменять SHA-1 на SHA-256 в поле Android OAuth fingerprint.
- Для локального API debug: package `kg.kasapchy.mobile.apidebug`, SHA-1
  `48:E0:68:7C:22:E3:3C:67:CA:C6:D9:82:08:78:BE:55:E0:0C:F4:B9`.
- SHA-1 локального upload key для sideload release:
  `CE:CF:E2:0E:47:86:50:EF:01:38:E4:14:B5:09:F2:D3:83:2A:78:3F`.
  Это публичные отпечатки, не секреты. Другой компьютер может иметь другой debug key.
- Добавить согласованные тестовые Google-аккаунты, проверить вход/отмену/nonce replay,
  затем пройти требуемую Google проверку branding перед production.

[Официальная интеграция Credential Manager](https://developer.android.com/identity/sign-in/credential-manager-siwg).

## Apple

Developer portal не загрузил форму входа в текущем Chrome (пустой iframe после
перезагрузки). Статус членства и доступ к Identifiers не подтверждены. Владелец
должен войти в https://developer.apple.com/account/ и подтвердить доступ.

Предлагаемые значения, ещё НЕ зарегистрированные:

- Primary App ID / iOS bundle: `kg.kasapchy.mobile`, Sign in with Apple enabled.
- Services ID для Android/web: `kg.kasapchy.mobile.web`.
- Domain: `kasapchy.api.ermekten.com`.
- Return URL: `https://kasapchy.api.ermekten.com/v1/auth/apple/callback`.
- Для отдельного staging Service ID: `kg.kasapchy.mobile.web.staging`, domain
  `test-kasapchy.api.ermekten.com`, соответствующий `/v1/auth/apple/callback`.
  Сейчас callback закрыт Nginx до готовности сервиса.

Apple требует связать web Services ID с подходящим primary App ID. Нельзя считать
создание только Services ID достаточным. Текущая реализация проверяет подписанный
id_token, audience, nonce, email_verified; live-вход не проверен. Перед релизом
нужно также подтвердить требования к отзыву Apple-токенов при удалении аккаунта;
обмен кода и хранение/revocation refresh token сейчас не реализованы.

[Настройка Apple web authentication](https://developer.apple.com/help/account/capabilities/configure-sign-in-with-apple-for-the-web/).
[Удаление аккаунтов и токенов](https://developer.apple.com/documentation/signinwithapplerestapi/revoke-tokens).

Настройки Android перечислены в `config/oauth.properties.example`. Не включать
кнопки с фиктивными client IDs ради обхода release guard. Production readiness
остаётся заблокирован до фактических external end-to-end проверок.
