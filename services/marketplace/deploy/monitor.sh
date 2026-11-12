#!/usr/bin/env bash
set -euo pipefail
umask 077
base=/opt/kasapchy
state_dir=/var/lib/kasapchy
mkdir -p "$state_dir"
failures=()
curl -fsS --max-time 5 http://127.0.0.1:5700/readyz >/dev/null || failures+=(database-readiness)
curl -fsS --max-time 10 --resolve test-kasapchy.api.ermekten.com:443:127.0.0.1 \
  https://test-kasapchy.api.ermekten.com/healthz >/dev/null || failures+=(https-health)
used=$(df -P "$base" | awk 'NR==2 {gsub(/%/,"",$5); print $5}')
[[ "$used" -lt 85 ]] || failures+=(disk-space)
if [[ ! -f "$base/backups/last-success" ]] || \
   [[ $(( $(date +%s) - $(stat -c %Y "$base/backups/last-success" 2>/dev/null || echo 0) )) -gt 129600 ]]; then
  failures+=(backup-older-than-36-hours)
fi
openssl x509 -checkend 1209600 -noout -in /etc/letsencrypt/live/kasapchy.api.ermekten.com/fullchain.pem >/dev/null || failures+=(certificate-expiry)
status=healthy
if [[ ${#failures[@]} -gt 0 ]]; then status="unhealthy: ${failures[*]}"; fi
previous=$(cat "$state_dir/monitor.status" 2>/dev/null || true)
printf '%s\n' "$status" > "$state_dir/monitor.status"
date -u +%FT%TZ > "$state_dir/monitor.checked-at"
if [[ "$previous" != "$status" ]]; then logger -t kasapchy-monitor -- "$status"; fi
printf '%s\n' "$status"
[[ ${#failures[@]} -eq 0 ]]
