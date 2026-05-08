# Migration Governance

## Environment Rules

- Development may use `DB_AUTO_MIGRATE=true` for local iteration, but versioned SQL migrations remain the source of record.
- Staging and production must run `cmd/migrate` with `DB_AUTO_MIGRATE=false`.
- Production should set `MIGRATION_STRICT_CHECKSUM=true`; an applied migration whose checksum differs from the local file must fail fast.

## Checksum Failure Handling

When checksum validation fails:

1. Stop deployment and keep the previous application version running.
2. Do not edit already-applied migration files to match production.
3. Create a new forward migration that repairs the schema or data.
4. If a migration was applied to the wrong environment, restore from backup or execute an explicitly reviewed compensating migration.
5. Re-run `go run ./cmd/migrate` and `go run ./cmd/verify-bootstrap`.

## Rollback And Compensation

Down migrations are provided for local and staging recovery. Production rollback should prefer backup restore or a new compensating forward migration because production data loss risk is higher than deployment speed.
