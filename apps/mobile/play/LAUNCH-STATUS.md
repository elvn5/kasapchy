# Запуск: проверяемое состояние на 13 ноября 2026

| Пункт | Сделано | Что блокирует завершение |
| --- | --- | --- |
| Google/Apple | Создан отдельный Google Cloud project kasapchy-marketplace; подготовлены package IDs, fingerprints и callback URLs в OAUTH.md | Google support email и consent, создание клиентов; доступ в Apple Developer, primary App ID, live OAuth и Apple token revocation |
| Владелец / privacy | Серверная policy template и external deletion page реализованы; не публикуются с выдуманными реквизитами | Имя оператора/компании, support email, финальные сроки хранения и согласование текста |
| Правила заказов | Текущее поведение и предложение записаны в ORDER-POLICY-DRAFT.md | Зона/стоимость доставки, отмена/возврат/спор, канал поддержки и соответствующая реализация |
| Инфраструктура | TLS, Actions deploy, encrypted daily/predeploy backups, успешный restore drill; readiness и мониторинг каждые 5 минут | Автоматическое offsite-хранилище, внешний мониторинг/канал уведомлений; SMTP отложен; публичные /v1 закрыты |
| Release QA | JVM/Android UI, настоящий local API flow, Go race/integration, подписанный demo AAB, lint, Bundletool, ELF/ZIP 16 KB и install smoke | Живой Google/Apple, реальный SMTP, production AAB и физические устройства/16 KB runtime |
| Play Console | Приложение kg.kasapchy.mobile уже создано, Draft. Сохранены русский текст карточки, Food & Drink и 4 анкеты: Ads / Government / Financial / Health. Console: 4 из 12 задач | Графика не загружена (пользователь отложил разрешение расширению); contact/privacy, access, rating/audience, Data safety, trader status, production AAB и тестирование ещё впереди |

13 ноября Console показала уже созданное приложение — повторно Create app не выполнялся.
[Dashboard](https://play.google.com/console/u/0/developers/8403504172845266761/app/4972055557310372525/app-dashboard).
Сохранённые анкеты: рекламы нет; не государственное; финансовых и медицинских функций нет.
Изменения находятся в Publishing overview, на review не отправлены. Тексты карточки
сохранены именно как draft. Категория Food & Drink сохранена отдельно; контакт не указан.

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
