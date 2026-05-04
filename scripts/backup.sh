#!/usr/bin/env bash
# Workend v2 — backup
#
# Dumps the postgres database and tars the repos + logs volumes into a
# timestamped directory. Run from anywhere; pulls compose context from the
# v2/ directory the script lives in.
#
# Usage:  ./scripts/backup.sh [output_dir]
# Output: <output_dir>/<timestamp>/{db.sql.gz, repos.tar.gz, logs.tar.gz}

set -euo pipefail

cd "$(dirname "$0")/.."  # cd into v2/

OUT="${1:-./backups}"
STAMP="$(date -u +%Y%m%d-%H%M%S)"
DEST="$OUT/$STAMP"

mkdir -p "$DEST"
echo "→ backing up to $DEST"

# 1. Database — pg_dump from inside the running db container.
echo "  · postgres dump"
docker compose exec -T db pg_dump -U workend -d workend --no-owner --no-acl \
  | gzip -9 > "$DEST/db.sql.gz"

# 2. repos volume — tar from a throwaway alpine attached to the volume.
echo "  · repos volume"
docker run --rm -v workend_repos:/data:ro -v "$(pwd)/$DEST":/out \
  alpine:3.20 sh -c 'cd /data && tar -czf /out/repos.tar.gz . 2>/dev/null || true'

# 3. logs volume — same pattern.
echo "  · logs volume"
docker run --rm -v workend_logs:/data:ro -v "$(pwd)/$DEST":/out \
  alpine:3.20 sh -c 'cd /data && tar -czf /out/logs.tar.gz . 2>/dev/null || true'

ls -lh "$DEST"
echo "✔ backup complete: $DEST"
