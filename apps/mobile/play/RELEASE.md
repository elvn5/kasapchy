# Google Play: пакет подготовки

Состояние на 12 ноября 2026: Android подключён к настоящему API; API развёрнут
через GitHub Actions в закрытом staging. Публичный production и карточка приложения
в Play Console ещё не созданы. Подписанный `demoRelease` — отдельное приложение
`kg.kasapchy.mobile.demo`, его нельзя выдавать за production `kg.kasapchy.mobile`.

iOS использует общий интерфейс и DemoStore. Подключение платформенного OAuth,
Keychain и рабочего NetworkStore в iOS ещё не выполнено; сборка на полном Xcode
в этой среде не проверена. Этот пакет предназначен для Android / Google Play.

## Готовые материалы

- `metadata/ru-RU`: название, краткое и полное описание на русском.
- `assets/store-icon.png`: 512 × 512.
- `assets/feature-graphic.png`: 1024 × 500.
- `assets/screenshots`: 4 настоящих снимка Compose UI, 1080 × 1920,
  полученных тестом StoreScreenshotsTest с демонстрационным каталогом.
- Векторные исходники графики рядом с PNG; палитра из предоставленного NSP logo.
- Android target/compile SDK 36, min SDK 24, versionCode 1 / versionName 0.1.0.
- R8, сокращение ресурсов, отдельная upload-подпись, запрет HTTP в release,
  шифрование токена Android Keystore, удаление аккаунта из профиля.

Описания и снимки — черновик карточки. Перед публикацией проверьте соответствие
фактически доступным товарам, зоне обслуживания, цене доставки и способу оплаты.
Google/Apple нельзя объявлять работающими до проверки с настоящими client IDs.

## Сборки и подпись

Из `apps/mobile`, JDK 17 и Android SDK 36:

```sh
./gradlew :androidApp:assembleDebug                 # офлайн-демо, .debug
./gradlew :androidApp:assembleApiDebug              # реальный API, .apidebug
# API эмулятора по умолчанию http://10.0.2.2:18080
./gradlew :shared:jvmTest :androidApp:lintDemoRelease

# Только локально; файл не хранится в репозитории:
source "$HOME/.config/kasapchy/signing/signing.env"
./gradlew :androidApp:bundleDemoRelease

# После фактической настройки всех сервисов:
./gradlew :androidApp:bundleRelease \
  -Pkasapchy.apiBaseUrl=https://API_DOMAIN \
  -Pkasapchy.googleClientId=GOOGLE_WEB_CLIENT_ID \
  -Pkasapchy.appleClientId=APPLE_SERVICES_ID \
  -Pkasapchy.privacyUrl=https://API_DOMAIN/privacy \
  -Pkasapchy.deletionUrl=https://API_DOMAIN/delete-account \
  -Pkasapchy.versionCode=1 -Pkasapchy.versionName=0.1.0
```

AAB: `androidApp/build/outputs/bundle/<variant>/androidApp-<variant>.aab`.
Сохраняйте R8 mapping соответствующего релиза из `build/outputs/mapping`.
Production build проверяет наличие HTTPS URL, OAuth IDs и четырёх переменных
`KASAPCHY_UPLOAD_*`; эта проверка конфигурации не заменяет проверку живых сервисов.

На машине владельца создан `$HOME/.config/kasapchy/signing/upload.jks`, alias
`upload`, и `signing.env` с паролями (права 0600). Сделайте зашифрованную резервную
копию каталога вне этого компьютера. Не отправляйте пароли в чат или git.
Публичный сертификат `upload-certificate.pem` находится рядом. При подключении
Play App Signing Google хранит ключ подписи приложения; локальный ключ нужен
для загрузки. В OAuth Android укажите отпечаток сертификата **Play App Signing**,
а для локальных тестов — соответствующий локальный сертификат.

## Что проверено

- 9 общих JVM-тестов модели/навигации и 3 UI-сценария демо.
- Настоящий API UI-сценарий: email OTP → продавец публикует товар → покупатель
  оформляет заказ → удаляет аккаунт. PostgreSQL и Mailpit, без имитации API.
- Go vet и race/integration tests: изоляция пользователей, остатки, повторный
  запрос заказа, статусы, OTP, отзыв сессий и удаление.
- Подписанный demoRelease AAB собран; jarsigner подтвердил подпись; lint прошёл.
- AAB BundleConfig требует PAGE_ALIGNMENT_16K; все 4 ELF-библиотеки имеют
  PT_LOAD alignment 16384. Bundletool validate прошёл. Скрипт проверки — в tools.
- Живые Google/Apple, SMTP-доставка внешним адресатам, HTTPS-домен и тест на
  физическом устройстве с 16 KB страницами ещё не проверены.

Повтор UI-тестов требует запущенного эмулятора; API-тест — также compose.local.yml:

```sh
./gradlew :androidApp:connectedDebugAndroidTest
./gradlew -Pkasapchy.testBuildType=apiDebug :androidApp:connectedApiDebugAndroidTest
./gradlew :androidApp:connectedDebugAndroidTest \
  -Pandroid.testInstrumentationRunnerArguments.captureStore=true
# Скриншоты: /sdcard/Pictures/Kasapchy
```

## Развёртывание

[Успешный запуск Actions](https://github.com/elvn5/kasapchy/actions/runs/43186080857).
Сервис из `services/marketplace`, собственные PostgreSQL и Mailpit; файлы в
`/opt/kasapchy`. API слушает только `127.0.0.1:5700`, Mailpit — `127.0.0.1:5701`.
Существующие приложения сервера не изменены. Workflow использует отдельный SSH-ключ
с forced command, а не произвольный root shell. Секреты БД остаются на сервере.

Для публичного запуска нужны:

1. Домен/DNS, TLS и reverse proxy на API; настроенные trusted proxies для корректных
   лимитов по IP. Не открывать PostgreSQL или Mailpit в интернет.
2. SMTP с STARTTLS, sender и SPF/DKIM/DMARC, APP_ENV=production; убрать Mailpit из
   цепочки доставки реальных писем. Проверить доставку, в том числе Apple relay.
3. Google OAuth web/Android clients и Apple Services ID/return URL
   `https://API_DOMAIN/v1/auth/apple/callback`; IDs в API и Android build.
4. MARKET_OPERATOR_NAME и MARKET_SUPPORT_EMAIL: настоящее имя оператора и контакт.
   Пока они отсутствуют, /privacy возвращает 503. Проверить текст политики,
   сроки хранения и законные обязательства по документам о сделках.
5. Резервные копии БД, пробное восстановление, мониторинг и план отката релиза.
   Текущий workflow проверяет health, но автоматически не откатывает миграции.
6. Реальные продавцы/товары и подтверждённые правила доставки, оплаты, отмены,
   возврата и рассмотрения обращений. Сейчас оплата вне приложения,
   доставка фиксирована 150 сом; отмена/возврат через поддержку ещё не описаны.

## Анкеты Play Console — заполнять по фактам

Публичный support email, сайт, юридические сведения, страны распространения,
целевая аудитория, рейтинг и доступ для ревьюера требуют данных владельца.
Не использовать личный email из Play Console как публичный без его решения.
Ревьюеру нужен действующий способ входа в рабочий сервис; не добавлять постоянный
OTP или обход авторизации. Google может потребовать закрытое тестирование —
условия проверяются именно для этого приложения/аккаунта в Console.

Черновик Data safety (не готовая декларация):

| Данные | Назначение / доступ |
| --- | --- |
| Email, ID пользователя, роль | Вход и работа аккаунта; email покупателя видит продавец его заказа |
| Имя / магазин, телефон (необязательные) | Профиль; имя магазина публично в каталоге |
| Адрес доставки | Выполнение заказа; видит соответствующий продавец |
| Товары, заказы, история статусов | Работа рынка и история заказов |
| IP, попытки входа, хеши токенов | Ограничение злоупотреблений и управление сессиями |
| Google/Apple ID token | Проверка входа у выбранного провайдера, без логирования токена |

В текущем мобильном коде нет рекламы, аналитических SDK, доступа к контактам,
GPS, камере или платёжным картам. Хостинг, SMTP, Google/Apple участвуют в обработке
данных; классификацию sharing/service providers сверить с их условиями и анкетой.
HTTPS обеспечивает шифрование передачи только после настройки production TLS.
Удаление в приложении и /delete-account требует свежего подтверждённого входа;
удаляет профиль, сессии, товары и связанные заказы. Ограниченные по сроку OTP/rate
записи очищаются фоново; сроки удаления резервных копий ещё нужно определить.

Официальные требования:
[Target API](https://support.google.com/googleplay/android-developer/answer/11926878?hl=en),
[графика](https://support.google.com/googleplay/android-developer/answer/9866151?hl=en),
[удаление аккаунта](https://support.google.com/googleplay/android-developer/answer/13327111?hl=en),
[подпись](https://developer.android.com/studio/publish/app-signing),
[16 KB](https://developer.android.com/guide/practices/page-sizes).
