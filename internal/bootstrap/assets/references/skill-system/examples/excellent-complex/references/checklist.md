# Migration Review Checklist

Checklist for migration-safety-review. Read this reference when a migration must be reviewed, not at skill load.

| Item | Check | Fail when |
|---|---|---|
| Rollback | Down migration restores schema and data | No down migration, or it drops data |
| Lock risk | Statements fit the release window | Long-running ALTER without batching |
| Defaults | New NOT NULL columns have defaults or backfill | Existing rows would violate |
| Idempotency | Re-run is safe | Second run errors |
| Data loss | Destructive changes are approved and recorded | DROP or TRUNCATE without review note |
