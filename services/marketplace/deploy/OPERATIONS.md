# Эксплуатация Касапчы

## Установлено на сервере

- `/opt/kasapchy/ops/backup.sh` и systemd `kasapchy-backup.timer`:
  ежедневно в 02:15 UTC (08:15 Asia/Bishkek), случайная задержка до 15 минут,
  Persistent=true. Также backup выполняется перед каждым деплоем.
- PostgreSQL custom dump передаётся сразу в OpenSSL CMS AES-256 encryption;
  plaintext dump на диск сервера не записывается. Сервер имеет только публичный
  сертификат `/opt/kasapchy/backup-recipient.pem`, приватного ключа на сервере нет.
- Копии и SHA-256 в `/opt/kasapchy/backups`, права каталога 0700 / файлов 0600,
  срок хранения 14 дней. Новые файлы публикуются атомарно после успешного dump.
- `kasapchy-monitor.timer`: каждые 5 минут проверяет readyz/PostgreSQL, HTTPS health,
  заполнение диска (<85%), возраст последней копии (<=36 часов), TLS (>14 дней).
  Статус в `/var/lib/kasapchy/monitor.status`, время в `monitor.checked-at`.
  При проблеме service завершается ненулевым кодом, изменение записывается в journal.
  Это локальный мониторинг: внешняя проверка отказа всего сервера и уведомления
  владельцу ещё не подключены, канал уведомлений не выбран.
- `/healthz` — жив процесс; `/readyz` — PostgreSQL отвечает за 2 секунды.
  readyz доступен только локально. Публично открыт только staging healthz.
- `MARKET_TRUSTED_PROXIES=172.22.0.1`: точный bridge gateway текущего Compose network.
  Nginx перезаписывает X-Forwarded-For адресом клиента, не дополняет входящий заголовок.
  При пересоздании Docker network проверить gateway и обновить значение. Доверять
  всем адресам, 0.0.0.0/0 или всем private ranges нельзя.

```sh
systemctl list-timers 'kasapchy-*'
systemctl start kasapchy-backup.service
systemctl start kasapchy-monitor.service
cat /var/lib/kasapchy/monitor.status
journalctl -u kasapchy-monitor.service --since '1 day ago'
journalctl -u kasapchy-backup.service --since '1 day ago'
```

## Восстановление и хранение ключа

На компьютере владельца: `$HOME/.config/kasapchy/backups/restore-key.pem` (0600),
`recipient.pem`. Каталог не входит в git. Контрольная копия, восстановленная
12 ноября, удалена с workstation 13 ноября после завершения проверки; ключ сохранён.
Потеря private key означает невозможность прочитать архивы. Сохранить его отдельную
зашифрованную резервную копию в надёжном месте; ключ не отправлять в чат.

12 ноября 2026 проверен backup `market-20261112T093951Z.dump.cms`: checksum,
расшифровка и pg_restore --exit-on-error в отдельном PostgreSQL 17 без сети, портов
и persistent volume. Восстановлены 7 таблиц; users/orders на момент снимка пусты.
Тестовый контейнер удалён. Рабочая БД не затрагивалась.

Для повторной проверки скопировать `.dump.cms` и `.sha256` по SSH на workstation:

```sh
./restore-drill.sh /path/market-TIMESTAMP.dump.cms /path/restore-key.pem /path/recipient.pem
```

Скрипт использует tmpfs до 512 MB и предназначен для текущей небольшой БД. По мере
роста базы увеличить лимит либо использовать выделенный диск восстановления.
Скопированный на workstation snapshot переживёт отказ сервера, но **регулярная
автоматическая offsite-копия пока не настроена**. Нужен выбранный владельцем
внешний storage и доступ к нему. Не считать 14-дневный архив на том же диске
защитой от потери всего сервера. Резервные копии могут содержать удалённые аккаунты
до окончания retention; это должно быть отражено в публичной privacy policy.

Перед восстановлением production: закрыть запись/вход в Nginx, сохранить текущую
копию, восстановить в НОВУЮ БД, сверить таблицы/строки/заказы и миграции, затем
переключить API. Не выполнять pg_restore --clean в рабочей базе. Перед открытием
сервиса обработать удаления аккаунтов, произошедшие после выбранного snapshot.

## Деплой и откат

GitHub Actions выполняет gofmt/vet/race/integration tests, отправляет tar по
forced-command SSH. Receiver делает encrypted backup перед заменой контейнеров,
проверяет `/readyz`, сохраняет ссылки `current` и `previous` только после успеха.
Health failure останавливает workflow; он не откатывает изменения БД автоматически.

Для отката кода: проверить совместимость текущей схемы с предыдущей версией,
закрыть внешний трафик если необходимо, создать Git revert проблемного изменения
и выполнить deployment через Actions. Не переключать старый API при несовместимой
миграции. Схема пока создаётся идемпотентным initial SQL; будущие изменения схемы
требуют версионируемых миграций и отдельной проверки upgrade/rollback.

Установку host scripts/systemd units выполняют отдельно от деплоя API:

```sh
# Из каталога deploy, после проверки diff:
scp backup.sh monitor.sh root@SERVER:/opt/kasapchy/ops/
scp systemd/kasapchy-* root@SERVER:/etc/systemd/system/
ssh root@SERVER 'systemctl daemon-reload && systemctl enable --now kasapchy-backup.timer kasapchy-monitor.timer'
```

Не заменять существующие сервисы других проектов или их Nginx-конфигурации.

## Жалобы и модерация объявлений

Оператору необходимо регулярно просматривать очередь и своевременно разбирать
жалобы: автоматическая оценка содержания и уведомления по SMTP не подключены.
Доступ к инструменту только через SSH администратора; публичного admin API нет.
Из `/opt/kasapchy` используйте тот же Compose-файл и env, что при развёртывании:

```sh
docker compose -p kasapchy --env-file /opt/kasapchy/.env -f /opt/kasapchy/current/compose.server.yml exec -T api /moderate list
docker compose -p kasapchy --env-file /opt/kasapchy/.env -f /opt/kasapchy/current/compose.server.yml exec -T api /moderate hide-product PRODUCT_ID
docker compose -p kasapchy --env-file /opt/kasapchy/.env -f /opt/kasapchy/current/compose.server.yml exec -T api /moderate suspend-seller SELLER_UUID
docker compose -p kasapchy --env-file /opt/kasapchy/.env -f /opt/kasapchy/current/compose.server.yml exec -T api /moderate resolve-report REPORT_ID
```

Перед действием сопоставьте ID и содержание жалобы с объявлением. Наличие жалобы
само по себе не доказывает нарушение. `hide-product` исключает товар из каталога
и новых заказов; `suspend-seller` также запрещает доступ продавцу. Уже оформленные
заказы сохраняются. После проверки отметьте жалобу рассмотренной. Для исправления
ошибочного решения доступны `show-product` и `restore-seller`. Обращения по
решениям принимает a.ermek.kg@gmail.com. Причины жалоб могут содержать личные
данные: не копируйте очередь в публичные логи, GitHub или чат.

Пользователь может пожаловаться из карточки товара, заблокировать продавца
и снять все свои блокировки в профиле. Перед каждой публикацией продавец
принимает `/content-rules`; API сохраняет дату согласия. Жалобы и блокировки
удаляются вместе с соответствующим аккаунтом.
