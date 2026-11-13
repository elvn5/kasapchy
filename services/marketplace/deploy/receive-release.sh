#!/usr/bin/env bash
# Installed on the host, invoked by a dedicated forced-command SSH key.
set -euo pipefail
if [[ ! "${SSH_ORIGINAL_COMMAND:-}" =~ ^deploy\ ([0-9a-f]{40})$ ]]; then
  echo 'Only deploy <commit-sha> is accepted.' >&2
  exit 1
fi
revision="${BASH_REMATCH[1]}"
base=/opt/kasapchy
exec 9>"$base/deploy.lock"
flock -n 9 || { echo 'Another deployment is running.' >&2; exit 1; }
release="$base/releases/$revision"
mkdir -p "$release"
archive=$(mktemp "$base/release.XXXXXX.tar.gz")
trap 'rm -f "$archive"' EXIT
cat > "$archive"
# Refuse traversal, links, device nodes and archives outside the expected source tree.
python3 - "$archive" "$release" <<'PY'
import sys,tarfile,pathlib
root=pathlib.Path(sys.argv[2]).resolve()
with tarfile.open(sys.argv[1]) as tar:
    for member in tar.getmembers():
        dest=(root/member.name).resolve()
        if (dest != root and root not in dest.parents) or not (member.isfile() or member.isdir()):
            raise SystemExit('Unsafe release archive')
    tar.extractall(root,filter='data') if sys.version_info >= (3,12) else tar.extractall(root)
PY
# Preserve a recoverable database snapshot before changing containers/migrations.
if [[ -x "$base/ops/backup.sh" && -L "$base/current" ]]; then "$base/ops/backup.sh"; fi
docker compose --project-name kasapchy --env-file "$base/.env" -f "$release/compose.server.yml" up -d --build --remove-orphans
for attempt in $(seq 1 30); do
  if curl -fsS --max-time 3 http://127.0.0.1:5700/readyz >/dev/null; then
    if [[ -f "$base/enable-public-api" ]]; then bash "$release/deploy/activate-public-api.sh" "$release"; fi
    if [[ -L "$base/current" ]]; then ln -sfn "$(readlink -f "$base/current")" "$base/previous"; fi
    ln -sfn "$release" "$base/current"
    printf 'Deployed %s\n' "$revision"
    exit 0
  fi
  sleep 2
done
echo 'Health check failed; inspect the kasapchy compose logs. No other application was modified.' >&2
exit 1
