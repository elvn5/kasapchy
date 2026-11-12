#!/bin/sh
# Создаёт отдельную базу для каждого микросервиса (database-per-service).
# Список баз — в POSTGRES_MULTIPLE_DATABASES через запятую, например: users,orders
# Скрипт выполняется образом postgres только при ПЕРВОЙ инициализации тома.
# Для уже существующего тома создайте базу вручную:
#   docker compose exec postgres createdb -U "$POSTGRES_USER" <name>
set -eu

for db in $(echo "${POSTGRES_MULTIPLE_DATABASES:-}" | tr ',' ' '); do
  echo "Creating database: $db"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    SELECT 'CREATE DATABASE "$db"'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db')\gexec
EOSQL
done
