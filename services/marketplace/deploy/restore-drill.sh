#!/usr/bin/env bash
# Run on the recovery workstation, never on the production database host.
# Usage: restore-drill.sh /absolute/path/market-....dump.cms /absolute/path/key.pem /absolute/path/recipient.pem
set -euo pipefail
umask 077
archive=${1:?encrypted backup required}
key=${2:?recovery private key required}
cert=${3:?recipient certificate required}
test -s "$archive" && test -s "$key" && test -s "$cert"
python3 - "$archive" <<'PY'
import hashlib,pathlib,sys
p=pathlib.Path(sys.argv[1])
expected=pathlib.Path(str(p)+'.sha256').read_text().split()[0]
with p.open('rb') as f:
    digest=hashlib.sha256()
    for chunk in iter(lambda:f.read(1024*1024),b''): digest.update(chunk)
if digest.hexdigest()!=expected: raise SystemExit('Backup checksum mismatch')
PY
container="kasapchy-restore-drill-$(date +%s)-$$"
trap 'docker rm -f -v "$container" >/dev/null 2>&1 || true' EXIT
# No network, no published ports, no persistent volume, no production credentials.
docker run -d --name "$container" --network none --tmpfs /var/lib/postgresql/data:rw,size=512m \
  -e POSTGRES_HOST_AUTH_METHOD=trust postgres:17-alpine >/dev/null
ready=false
for attempt in $(seq 1 30); do
  if docker exec "$container" pg_isready -U postgres >/dev/null 2>&1; then ready=true; break; fi
  sleep 1
done
[[ "$ready" == true ]]
openssl cms -decrypt -binary -inform DER -in "$archive" -recip "$cert" -inkey "$key" | \
  docker exec -i "$container" pg_restore -U postgres -d postgres --no-owner --no-acl --exit-on-error
tables=$(docker exec "$container" psql -U postgres -d postgres -Atc "SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name LIKE 'market_%'")
[[ "$tables" -ge 7 ]]
docker exec "$container" psql -U postgres -d postgres -v ON_ERROR_STOP=1 -Atc \
  'SELECT count(*) FROM market_users; SELECT count(*) FROM market_orders; SELECT count(*) FROM market_products;' >/dev/null
printf 'Restore drill passed: %s marketplace tables; temporary database removed on exit.\n' "$tables"
