#!/usr/bin/env bash
# Workend v2 — restore
#
# Restores a backup created by ./scripts/backup.sh. DESTRUCTIVE: drops the
# existing workend database and replaces volume contents.
#
# Usage:  ./scripts/restore.sh <backup_dir>
# Where <backup_dir> is e.g. ./backups/20260501-120000

set -euo pipefail

cd "$(dirname "$0")/.."  # cd into v2/

if [ $# -ne 1 ]; then
  echo "usage: $0 <backup_dir>" >&2
  exit 1
fi

SRC="$1"
for f in db.sql.gz repos.tar.gz logs.tar.gz; do
  if [ ! -f "$SRC/$f" ]; then
    echo "missing: $SRC/$f" >&2
    exit 1
  fi
done

echo "‼  DESTRUCTIVE: this will replace the database and volume contents."
read -r -p "type 'restore' to confirm: " CONFIRM
[ "$CONFIRM" = "restore" ] || { echo "aborted"; exit 1; }

echo "→ stopping api and web (keeping db + dagger-engine up)"
docker compose stop api web

# 1. Database
echo "  · drop + recreate workend db"
docker compose exec -T db psql -U workend -d postgres -c \
  "DROP DATABASE IF EXISTS workend; CREATE DATABASE workend OWNER workend;"
echo "  · restoring db.sql.gz"
gunzip -c "$SRC/db.sql.gz" | docker compose exec -T db psql -U workend -d workend

# 2. repos volume — wipe + restore
echo "  · restoring repos volume"
docker run --rm -v workend_repos:/data \
  alpine:3.20 sh -c 'rm -rf /data/* /data/.[!.]* 2>/dev/null || true'
docker run --rm -v workend_repos:/data -v "$(pwd)/$SRC":/in:ro \
  alpine:3.20 sh -c 'cd /data && tar -xzf /in/repos.tar.gz'

# 3. logs volume — wipe + restore
echo "  · restoring logs volume"
docker run --rm -v workend_logs:/data \
  alpine:3.20 sh -c 'rm -rf /data/* /data/.[!.]* 2>/dev/null || true'
docker run --rm -v workend_logs:/data -v "$(pwd)/$SRC":/in:ro \
  alpine:3.20 sh -c 'cd /data && tar -xzf /in/logs.tar.gz'

echo "→ restarting api and web"
docker compose start api web

echo "✔ restore complete"
