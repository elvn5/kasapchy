# Касапчы — Android / iOS

Мобильное приложение на **Kotlin Multiplatform + Compose Multiplatform**. Экраны,
палитра и переходы реализуют согласованный wireflow. UI и демо-логика общие для Android и iOS. Android также подключён к новому
Go/Gin API в `services/marketplace`; iOS пока запускает демо.

## Запуск

Требования: JDK 17, Android SDK 36 для Android. Для iOS нужен macOS с полным
Xcode (совместимая версия по матрице выбранного Kotlin: Xcode 16.4), поддерживаются iPhone
и iPad с iOS 16+, а также ARM64 Simulator. Откройте **эту папку** в Android Studio
как отдельный Gradle-проект.

```sh
cd apps/mobile
# Укажите SDK через ANDROID_HOME или локальный, игнорируемый local.properties:
# sdk.dir=/absolute/path/to/Android/sdk
./gradlew :androidApp:assembleDebug
./gradlew :androidApp:installDebug  # подключённый телефон / запущенный эмулятор
./gradlew :shared:jvmTest :androidApp:lintDebug
./gradlew :androidApp:connectedDebugAndroidTest # UI-тесты на запущенном эмуляторе
```

APK: `androidApp/build/outputs/apk/debug/androidApp-debug.apk`.

Для iOS откройте `iosApp/iosApp.xcodeproj`, выберите общую схему **Kasapchy** и
ARM64 Simulator. Build Phase автоматически запускает
`:shared:embedAndSignAppleFrameworkForXcode`. Для физического iPhone укажите свою
Signing Team в настройках target. Секреты и signing credentials не нужны для
демо-входа. Если Xcode не находит Java, задайте `JAVA_HOME` в окружении сборки.

```sh
xcodebuild -project iosApp/iosApp.xcodeproj -scheme Kasapchy \
  -configuration Debug -sdk iphonesimulator \
  -destination 'generic/platform=iOS Simulator' ARCHS=arm64 CODE_SIGNING_ALLOWED=NO build
```

## Что можно пройти в demo/debug

- Старт → покупатель / касапчы → регистрация или вход.
- Только email обязателен. Одноразовый демо-код — **1234**, пароль не нужен.
- Google / Apple → подтверждение демо-аккаунта → рынок; есть отмена входа.
- Рынок → поиск по товару/продавцу → категория, цена и сортировка → товар.
- Покупатель: товар → количество → самовывоз/доставка → подтверждение → заказ.
- Касапчы: добавление товара → публикация; входящие заказы → принят → готов → завершён.
- Три нижние вкладки: **Рынок**, **Мои заказы**, **Профиль**.
- Профиль: необязательные имя / название магазина и телефон; телефон можно удалить.
- Назад в заголовке и системный Back на Android. После подтверждения заказа
  возврат ведёт к списку заказов, а не к повторной отправке формы.

## Ограничения demo/debug

В demo/debug нет сетевых запросов, реальных писем, OAuth, платежей или серверной публикации.
Google/Apple показывают явно обозначенный экран имитации, не запрашивают пароли
и используют адреса на `example.com`. Ручной email также используется только в
оперативной памяти. Данные сохраняются при повороте Android-экрана, но сбрасываются
при завершении процесса приложения. Профили и покупки разделены по email; роль
выбирается перед входом. Все демо-продавцы управляют прилавком **Азамат эт**, для
которого заранее есть один входящий заказ. Товары «Береке» и «Айыл ферма» доступны
покупателю, но их заказы не попадают в управление другого продавца.

Цена фиксирована в целых сомах, вес — целые килограммы. Доставка — 150 сом;
самовывоз бесплатный. Заказ хранит снимок товара и суммы, резервирует остаток.
Иллюстрации категорий — emoji, как в wireflow; внешний вид зависит от ОС.

## Структура

- `shared/src/commonMain/.../entities` — модели, валидация, поиск, демо-каталог.
- `shared/src/commonMain/.../features` — состояние сессии и переходы (`DemoStore`).
- `shared/src/commonMain/.../pages` — экраны.
- `shared/src/commonMain/.../shared/ui` — компоненты и светлая/тёмная темы.
- `shared/src/commonMain/.../app` — сборка интерфейса и нижняя навигация.
- `androidApp` — Activity, Android Back, ViewModel для сохранения сессии при повороте.
- `shared/src/iosMain` + `iosApp` — ComposeUIViewController и SwiftUI entry point.
- `shared/src/commonTest` — тесты авторизации, профиля, фильтров, заказов и навигации.
- `androidApp/src/androidTest` — три UI-сценария: покупка и телефон, продавец, Apple и Back.

Палитра из `nsp_logo.svg`: синий `#007EC5`, серый `#A7AFB2`. Кнопки используют
более тёмный `#0076B8` для контраста, светлые поверхности — `#EDF5FA`.
Название приложения остаётся «Касапчы»; логотип NSP использован как источник цвета.

Версии закреплены: Kotlin 2.2.20, Compose Multiplatform 1.9.3, AGP 8.11.1,
Gradle 8.14, JDK 17. [Матрица совместимости Kotlin](https://kotlinlang.org/docs/multiplatform/multiplatform-compatibility-guide.html)
и [прямая интеграция с Xcode](https://kotlinlang.org/docs/multiplatform/multiplatform-direct-integration.html).

## Подготовка к Google Play

Пакет публикации, команды AAB, подпись и реальные блокеры: [play/RELEASE.md](play/RELEASE.md).
`apiDebug` и `release` используют NetworkStore: email OTP, каталог, заказы и профиль
через Go/Gin API. Телефон остаётся необязательным. Google/Apple требуют настройки
внешних client IDs. Production-выпуск заблокирован до конфигурации рабочих сервисов.
