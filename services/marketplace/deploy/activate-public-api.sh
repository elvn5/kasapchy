#!/usr/bin/env bash
# Called by the GitHub Actions receiver after the new API passes readiness.
set -euo pipefail
release=${1:?release directory required}
config=/etc/nginx/conf.d/kasapchy.conf
backup=$(mktemp /opt/kasapchy/nginx-before-release.XXXXXX)
cp -p "$config" "$backup"
install -m 644 "$release/deploy/nginx-public.conf" "$config"
if ! nginx -t || ! systemctl reload nginx; then
  cp -p "$backup" "$config"
  nginx -t && systemctl reload nginx
  echo 'Public API activation failed; restored previous nginx configuration.' >&2
  exit 1
fi
rm -f "$backup"
