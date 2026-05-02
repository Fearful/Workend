# Workend scripts

## Backup

```sh
./scripts/backup.sh                # writes to ./backups/<timestamp>/
./scripts/backup.sh /mnt/backups   # custom output dir
```

Each run produces three files in a timestamped subdirectory:

| File | Contents |
|---|---|
| `db.sql.gz` | `pg_dump` of the workend database (gzipped) |
| `repos.tar.gz` | Contents of the `workend_repos` volume — all cloned repos |
| `logs.tar.gz` | Contents of the `workend_logs` volume — all run logs |

The compose stack must be running (specifically the `db` container) for `pg_dump` to succeed.

Cron a nightly backup with:

```cron
0 3 * * * cd /path/to/workend/v2 && ./scripts/backup.sh /mnt/nas/workend
```

## Restore

```sh
./scripts/restore.sh ./backups/20260501-120000
```

**Destructive.** Drops the workend database and replaces volume contents with the backup. Prompts for confirmation. The api and web containers are stopped during restore and restarted at the end; the dagger-engine and db containers stay up.

## Notes

- Volumes are restored byte-for-byte; UUIDs and paths from the backup work as-is on the new instance because they're independent of the host filesystem.
- Restoring across a Workend version boundary (different schema) is not supported. Backups are intended for same-version disaster recovery / migration to a new host.
- For point-in-time recovery you'd want WAL archiving on the postgres container — out of scope here.
