#!/usr/bin/env bash
set -euo pipefail
umask 077
base=/opt/kasapchy
backup_dir="$base/backups"
mkdir -p "$backup_dir"
exec 8>"$base/backup.lock"
flock -w 60 8 || { echo "Another backup is still running" >&2; exit 1; }
stamp=$(date -u +%Y%m%dT%H%M%SZ)
file="$backup_dir/market-$stamp.dump.cms"
test -s "$base/backup-recipient.pem"
trap 'rm -f "$file.partial"' EXIT
docker compose --project-name kasapchy --env-file "$base/.env" -f "$base/current/compose.server.yml" \
  exec -T db pg_dump -U marketplace -d marketplace --format=custom --no-owner --no-acl | \
  openssl cms -encrypt -binary -aes-256-cbc -outform DER -out "$file.partial" "$base/backup-recipient.pem"
test -s "$file.partial"
mv "$file.partial" "$file"
sha256sum "$file" > "$file.sha256"
# Only files created by this backup job are subject to its 14-day retention.
find "$backup_dir" -maxdepth 1 -type f -name 'market-*.dump.cms*' -mtime +14 -delete
printf '%s\n' "$stamp" > "$backup_dir/last-success"
printf 'Encrypted database backup created: %s\n' "$file"
