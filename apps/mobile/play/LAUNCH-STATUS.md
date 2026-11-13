# Запуск: проверяемое состояние на 13 ноября 2026

## Текущее состояние: отправлено на проверку (13 ноября, закрытый тест)

Этот раздел заменяет устаревшие статусы ниже; далее сохранён журнал подготовки.

- Отправлены 15 изменений на проверку Google: Alpha `0.1.0 — закрытый тест`,
  только Кыргызстан, карточка магазина и декларации. Console подтвердила
  «15 changes sent for review» и «Changes in review». На момент отправки идут
  предварительные автоматические проверки (показано до 14 минут); после их успеха
  изменения переходят к рассмотрению. Это не одобрение и не production-публикация.
  Managed publishing выключен. Отсчёт 14 дней тестирования ещё не начался.
- API доступен по HTTPS. Вход по email и паролю без кода, OTP/OAuth в первом
  релизе отключены. `/privacy` и `/delete-account` доступны публично. SMTP не нужен
  для этого способа входа; восстановление пароля через email пока отсутствует.
- Подготовлены отдельные аккаунты Buyer/Seller для Google Play с тестовым товаром,
  не предназначенным для покупки. Пароли находятся только в локальном защищённом
  хранилище. После явного разрешения владельца обе пары credentials и инструкции
  сохранены в Google Play Sign in details. Дополнительное использование данных
  для тестирования на устройствах Google и доверенных партнёров выключено.
- Добавлены жалобы на товар/продавца, блокировка и снятие блокировок, правила
  публикации и CLI оператора для рассмотрения жалоб. Нужна регулярная модерация
  владельцем; автоматические уведомления не подключены. Инструкция в OPERATIONS.md.
- Серверное обновление модерации развёрнуто: GitHub Actions run `43401021656`
  завершён успешно. Публичные `/content-rules`, `/privacy`, `/delete-account`,
  `/v1/products` отвечают 200; `/v1/me` без токена — 401; закрытые readyz/OTP — 404.
- Новый подписанный AAB: versionCode 1 / versionName 0.1.0 (подтверждено Play Console), Bundletool и jarsigner
  пройдены; SHA-256 `afd473d2bdaabe2d0b99324fdd8facc6b1e4e5e8d0e1ad674068fc1aaa6702e1`.
  JVM tests, release lint и Go race/integration на PostgreSQL прошли. Android API 30
  UI-тест с реальным Go API прошёл: регистрация, согласие с правилами, публикация,
  заказ, жалоба, блокировка, снятие блокировок и удаление аккаунта.
- IARC завершён после явного согласия владельца: статус Completed, IARC Generic
  12+, Users Interact / Parental Guidance Recommended. Это рейтинг содержания,
  отдельно от выбранной владельцем целевой аудитории 18+.
- К Alpha привязан список «Касапчы — закрытый тест» из 12 проверенных Google-аккаунтов.
  Из предоставленных 13 адресов Console отклонила 4; они не сохранены. Список адресов
  не включён в публичный репозиторий. Три дополнительных адреса владельца приняты
  Console и сохранены; таблица подтверждает 12 пользователей. Это список доступа,
  а не подтверждение вступления участников: отсчёт тестирования ещё не начался.
  Контакт обратной связи: a.ermek.kg@gmail.com. Активация Alpha включена в отправленные изменения.
- После переподключения расширения Chrome загрузка заработала. AAB принят;
  загружены иконка 512×512, баннер 1024×500 и 4 скриншота 1080×1920.
  Карточка сохранена и отправлена. Иконка и баннер помечены как созданные/изменённые
  с помощью ИИ; реальные скриншоты приложения не помечены.
- Декларации завершены: Sign in details, аудитория только 18+, Advertising ID — No,
  Data safety сохранена финальной кнопкой Save (не draft). В Data safety добавлена
  безопасность для Other user-generated content и функциональность для App interactions;
  URL удаления обновлён на `/delete-account`. App content показывает «You’re all caught up».
- Блокирующих ошибок при подтверждении релиза нет. Одно предупреждение: отсутствуют
  native debug symbols. ReTrace mapping file приложен автоматически.
- Следующий этап: дождаться успешной проверки Google и доступности Alpha, затем
  минимум 12 участников должны вступить в тест и оставаться в нём 14 дней подряд.
  Ссылка после открытия теста: https://play.google.com/apps/testing/kg.kasapchy.mobile
- Правила доставки/отмены/возврата ещё не утверждены; тест не предназначен для
  реальных покупок. До production нужны отдельные решения владельца.

## Журнал подготовки

| Пункт | Сделано | Что блокирует завершение |
| --- | --- | --- |
| Google/Apple | Создан отдельный Google Cloud project kasapchy-marketplace; подготовлены package IDs, fingerprints и callback URLs в OAUTH.md | Google support email и consent, создание клиентов; доступ в Apple Developer, primary App ID, live OAuth и Apple token revocation |
| Владелец / privacy | Подтверждены ИП Ermek Cherikchiev, Кыргызстан, a.ermek.kg@gmail.com. Подготовлена полная политика с обработкой заказов, журналов и резервных копий; support email сохранён в Play | Политика опубликована на HTTPS /privacy, URL сохранён в Console; self-service deletion и SMTP ещё закрыты |
| Правила заказов | Текущее поведение и предложение записаны в ORDER-POLICY-DRAFT.md | Зона/стоимость доставки, отмена/возврат/спор, канал поддержки и соответствующая реализация |
| Инфраструктура | TLS, Actions deploy, encrypted daily/predeploy backups, успешный restore drill; readiness и мониторинг каждые 5 минут | Автоматическое offsite-хранилище, внешний мониторинг/канал уведомлений; SMTP отложен; публичные /v1 закрыты |
| Release QA | JVM/Android UI, настоящий local API flow, Go race/integration, подписанный demo AAB, lint, Bundletool, ELF/ZIP 16 KB и install smoke | Живой Google/Apple, реальный SMTP, production AAB и физические устройства/16 KB runtime |
| Play Console | Приложение kg.kasapchy.mobile уже создано, Draft. Сохранены русский текст карточки, Food & Drink и 4 анкеты: Ads / Government / Financial / Health. Console: 7 из 12 задач завершены, включая Trader status; Data safety сохранена как draft | Графика не загружена (пользователь отложил разрешение расширению); access, rating/audience, завершение Data safety, production AAB и тестирование ещё впереди |

13 ноября Console показала уже созданное приложение — повторно Create app не выполнялся.
[Dashboard](https://play.google.com/console/u/0/developers/8403504172845266761/app/4972055557310372525/app-dashboard).
Сохранённые анкеты: рекламы нет; не государственное; финансовых и медицинских функций нет.
Изменения находятся в Publishing overview, на review не отправлены. Тексты карточки
сохранены именно как draft. Категория Food & Drink и публичный email a.ermek.kg@gmail.com сохранены.

Для этого приложения Console явно требует закрытый тест с минимум 12 участниками,
непрерывно зарегистрированными в нём не менее 14 дней; сейчас участников 0,
кнопка Apply for production недоступна. Internal test этот этап не заменяет.
Графику пользователь попросил пока не загружать: расширение Chrome требует
Allow access to file URLs. Иконка, баннер и 4 скриншота остаются в assets.

SMTP намеренно отложен пользователем. Без него не открывать публичную регистрацию
на существующий Mailpit. iOS остаётся demo; полноценная iOS release-подготовка
не входит в текущий пакет Google Play.

## Минимальная финальная проверка перед production

1. Google: sign-in/cancel, новый и существующий email, неправильная audience и replay.
2. Apple: actual browser callback, nonce/proof, Hide My Email, удаление и отзыв токенов.
3. Email: внешняя доставка, срок/одноразовость кода, лимиты и retry после SMTP failure.
4. Buyer/seller: реальные товары, параллельные покупки последнего остатка, повтор
   POST после потери сети, недоступность чужих заказов, согласованные отмены/возвраты.
5. Account deletion: свежая авторизация, logout/revocation, web flow, backups retention.
6. Release из Play internal track: Google Play signing SHA-1, minified API code,
   supported Android versions, поворот/процесс death, TalkBack/font scaling, offline,
   16 KB runtime, screenshots и тексты совпадают с опубликованной функциональностью.

Физическое Android-устройство сейчас не подключено — доступен только эмулятор.
Не подменять недоступные live-проверки демо-сценариями или фиктивными client IDs.

13 ноября: публичный email поддержки сохранён в Store settings; политика
https://kasapchy.api.ermekten.com/privacy отвечает HTTP 200 без входа, URL сохранён
в Privacy Policy как изменение для последующего review. Основной /v1 остаётся 503.
Оператор: ИП Ermek Cherikchiev, Кыргызская Республика (подтверждено владельцем).


13 ноября, продолжение анкет:
- Пользователь подтвердил аудиторию только 18+ и первоначальный рынок только Кыргызстан.
  Это решения владельца; возраст и страны ещё не сохранены в соответствующих формах Console.
- Trader status сохранён: статус предпринимателя применяется к приложению (наследуется
  от developer account). Необязательный Terms of Service URL оставлен пустым, поскольку
  правила заказов ещё не утверждены. Dashboard подтверждает 7/12 выполненных задач.
- Пользователь попросил пока пропустить IARC; согласие с условиями не принято,
  возрастной рейтинг не получен.
- Target audience заблокирована Console до завершения Sign in details. Рабочих
  reviewer credentials пока нет; фиктивные данные и отсутствие ограничений не заявлялись.
- Data safety заполнена и сохранена кнопкой Save as draft. Финальное сохранение
  заблокировано незавершённой Target audience. На review не отправлена.
  Выбраны HTTPS, email + OTP и OAuth (для предполагаемой релизной конфигурации),
  URL запроса удаления аккаунта https://kasapchy.api.ermekten.com/privacy.
  Отдельный путь удаления части данных без аккаунта не заявлен.
- Data safety draft: Name, Email address, User IDs, Address, Phone number,
  Purchase history, App interactions, Other user-generated content, Device or other IDs.
  Все сохраняются, не ephemeral. Имя/телефон/адрес/заказы/объявления — optional;
  email/ID аккаунта/технические запросы/IP — required. Цели: функциональность и
  управление аккаунтом по типу данных; безопасность для email/ID/IP/запросов;
  Analytics для диагностики по серверным запросам (без аналитического SDK).
  Передачи поставщикам обработки и публикации/заказы по инициативе пользователя
  учтены по исключениям Google из Shared, поэтому preview показывает No data shared.
  Перед отправкой сверить это с финальной сборкой, OAuth SDK и фактическими поставщиками;
  отсутствие Shared в форме не означает отсутствия передачи продавцу в privacy policy.
- Требования Google к удалению допускают страницу с email поддержки и инструкцией:
  https://support.google.com/googleplay/android-developer/answer/13327111?hl=en
  Определения сбора/передачи:
  https://support.google.com/googleplay/android-developer/answer/10787469?hl=en


13 ноября, изменение первого релиза по просьбе владельца:
- Live Android использует email + пароль без подтверждения email; SMTP не нужен для
  регистрации и входа. Не путать это с входом по одному email: пароль обязателен.
- Добавлены register/login API, bcrypt cost 12, лимиты IP/email, запрет замены пароля
  повторной регистрацией и автоматического объединения unverified аккаунта с OTP/OAuth.
- Google/Apple в live UI скрыты до отдельного включения. Демо продолжает прежний wireflow.
- Email OTP в серверном процессе выключен по умолчанию; включается только через
  MARKET_EMAIL_OTP_ENABLED=true. SMTP-параметры нужны только в этом режиме.
- Восстановление пароля пока отсутствует; UI предупреждает об этом при регистрации.
  Нельзя сбрасывать пароль или удалять аккаунт лишь по письму с неподтверждённого email.
- Веб-удаление переведено на пароль; публичный nginx API остаётся закрыт до подготовки
  реального тестового релиза. Сам факт сборки AAB не означает доступности сервиса.
- Черновик Console Data safety обновлён и сохранён: Username and password вместо
  Username and other authentication / OAuth. Описание карточки обновлено локально;
  его текст в Console ещё нужно синхронизировать перед отправкой.

- Проверки этой версии: Go vet и полный race/integration suite с PostgreSQL прошли;
  новые тесты проверяют неправильный пароль, повторы регистрации, роль, существующие
  OTP-аккаунты, лимиты и отзыв сессий после удаления. JVM tests и assembleApiDebug прошли.
- Android API 30: реальный UI flow password registration → seller product → buyer order
  → account deletion прошёл с локальным API без SMTP. iOS native allTests остановлен
  во время первичной загрузки toolchain; iOS проверку не считать завершённой.
- Подписан production-package AAB kg.kasapchy.mobile, versionCode 1/versionName 1.0.0:
  androidApp/build/outputs/bundle/release/androidApp-release.aab. Release lint/R8,
  bundletool validate и jarsigner verify прошли. Это кандидат: публичный API всё ещё 503,
  в Play AAB не загружен, IARC/графика/доступ для review/closed test ещё не завершены.

- GitHub Actions run 43396664484 успешно развернул b17339a. На сервере readyz=ready,
  password/login принимает запросы (400 на пустой JSON), OTP request отключён (403).
  Публичная /privacy уже содержит bcrypt; /v1/products снаружи всё ещё 503.


13 ноября, подготовка закрытого релиза по запросу «публиковать»:
- В Console сохранены обновлённое полное описание (email + пароль), страна Kyrgyzstan
  для Closed testing - Alpha (track 4699056520390756582) и draft release 1 с именем
  «1.0.0 — закрытый тест» и русскими release notes. AAB ещё не загружен.
- Run 43398178368 успешно развернул публичный HTTPS API. В production включён
  APP_ENV=production, MARKET_EMAIL_OTP_ENABLED=false. Receiver применяет nginx-public.conf
  после readiness только при наличии /opt/kasapchy/enable-public-api; nginx -t и reload
  с восстановлением предыдущей конфигурации при ошибке. Прочие приложения не менялись.
- Снаружи проверены: /v1/products, /privacy, /delete-account и deletion JS = 200;
  /v1/me без токена = 401; /readyz и legacy OTP routes = 404. Внешний IP проксируется
  доверенным nginx, добавлен общий rate limit. Postgres и readiness не открыты наружу.
- Подготовлены и через HTTPS проверены отдельные Buyer/Seller review accounts.
  Пароли только в локальном ~/.config/kasapchy/play-review/accounts.json (0600),
  в git и чат не выводились. Опубликован один явно помеченный тестовый товар,
  не предназначенный для реальной продажи; инструкции Google объясняют этот статус.
- В Sign in details открыта подготовленная форма Buyer без credentials. Отдельное
  согласие на передачу тестовых credentials Google запрошено, пока не получено.
- Также ожидаются: согласие IARC, разрешение расширению на file URLs или ручная загрузка
  AAB/графики, список минимум 12 тестировщиков. Общая команда публиковать не подменяет
  отдельное принятие юридических условий. Публичный выпуск недоступен до closed test
  (12 opted-in непрерывно 14 дней) и одобрения production access Google.
