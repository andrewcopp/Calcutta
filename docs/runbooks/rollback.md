# Rollback Runbook

Procedures for rolling back bad deploys, migrations, and data corruption during the tournament.

## Pre-Deploy Checklist

Before every production deploy:

1. **Record the current state**
   ```bash
   # Current ECS task definition revision
   aws ecs describe-services --cluster $ECS_CLUSTER --services $ECS_SERVICE \
     --query 'services[0].taskDefinition' --output text

   # Current migration version
   make query SQL="SELECT version, dirty FROM schema_migrations"
   ```

2. **Verify staging is green** — same image tag must pass staging first

3. **Check migration backward-compatibility** — can the current backend code still work if the new migration is applied? (See safety guidelines below.)

4. **Create a manual RDS snapshot before risky deploys**
   ```bash
   aws rds create-db-snapshot \
     --db-instance-identifier $RDS_INSTANCE \
     --db-snapshot-identifier "pre-deploy-$(date +%Y%m%d-%H%M%S)"
   ```

---

## Scenario A: Bad Backend Deploy (Code Only)

The new backend image has a bug, but no migration was run.

**Option 1: Re-trigger deploy with previous image tag**
- Go to GitHub Actions > Backend Deploy > Run workflow
- Select `prod` environment, enter the previous image tag (git SHA)

**Option 2: CLI rollback to previous task definition**
```bash
# List recent task definition revisions
aws ecs list-task-definitions --family-prefix $ECS_TASK_DEFINITION \
  --sort DESC --max-items 5

# Update service to previous revision
aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE \
  --task-definition "$ECS_TASK_DEFINITION:$PREVIOUS_REVISION" \
  --force-new-deployment

# Wait for stability
aws ecs wait services-stable --cluster $ECS_CLUSTER --services $ECS_SERVICE
```

**Verify:** `curl -s https://api.example.com/healthz | jq .`

---

## Scenario B: Bad Frontend Deploy (S3 + CloudFront)

**Option 1: Re-deploy previous build**
```bash
# Rebuild from the last known-good commit
git checkout $PREVIOUS_SHA
cd frontend && npm ci && npm run build

# Upload to S3
aws s3 sync dist/ s3://$S3_BUCKET --delete

# Invalidate CloudFront cache
aws cloudfront create-invalidation --distribution-id $CF_DIST_ID --paths "/*"
```

**Option 2:** Re-run the frontend deploy workflow at the previous commit.

---

## Scenario C: Bad Migration

### Understanding the migrate CLI

- `migrate -up` runs ALL pending up migrations
- `migrate -down` runs ALL down migrations — **never use in production**
- `migrate -steps=N` runs N steps (positive = up, negative = down)
- `migrate -force=V` sets the version to V and clears the dirty flag (does NOT run SQL)

### Rolling back one migration

```bash
# 1. Run the specific .down.sql file manually
psql $DATABASE_URL < migrations/schema/YYYYMMDDHHMMSS_description.down.sql

# 2. Set the migration version to the previous version
go run ./cmd/migrate -force=$PREVIOUS_VERSION
```

### Dirty state recovery

If a migration failed halfway and the schema_migrations table is dirty:

```bash
# 1. Manually fix the database state (undo partial changes)
# 2. Force the version back to the last clean version
go run ./cmd/migrate -force=$LAST_CLEAN_VERSION
```

---

## Scenario D: Data Corruption (RDS Point-in-Time Recovery)

When data is corrupted and you need to restore from backup.

```bash
# 1. Restore to a new instance via PITR
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier $RDS_INSTANCE \
  --target-db-instance-identifier "$RDS_INSTANCE-restore-$(date +%Y%m%d)" \
  --restore-time "2026-03-15T10:30:00Z"

# 2. Wait for the new instance to become available
aws rds wait db-instance-available \
  --db-instance-identifier "$RDS_INSTANCE-restore-$(date +%Y%m%d)"

# 3. Update DATABASE_URL in ECS task definition to point to restored instance

# 4. Restart the ECS service
aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE \
  --force-new-deployment

# 5. Verify data integrity
make query SQL="SELECT count(*) FROM core.pools"
make query SQL="SELECT count(*) FROM core.portfolios"
```

---

## Scenario E: Combined Bad Deploy + Bad Migration

Undo in **reverse order** of application:

1. Roll back the code (Scenario A) — so the old code isn't trying to use new schema
2. Roll back the migration (Scenario C)
3. Verify the application is healthy

---

## Migration Safety Guidelines

### Safe operations (no rollback risk)
- Add a new table
- Add a nullable column
- Add a column with a default value
- Add an index (use `CONCURRENTLY` for large tables)

### Unsafe operations (require multi-step deploy)
- **Drop column:** First deploy code that stops reading the column, then drop it
- **Rename column:** Add new column, deploy code to write both, backfill, deploy code to read new, drop old
- **Add NOT NULL:** Add column as nullable with default, backfill existing rows, then add constraint
- **Change column type:** Add new column with new type, backfill, swap reads, drop old

### General rules
- Every migration must have a working `.down.sql`
- Test the down migration locally before deploying
- Never modify an existing migration file — create a new one

---

## RDS Backup Verification

### Check automated backups are enabled
```bash
aws rds describe-db-instances --db-instance-identifier $RDS_INSTANCE \
  --query 'DBInstances[0].{BackupRetentionPeriod:BackupRetentionPeriod,DeletionProtection:DeletionProtection,LatestRestorableTime:LatestRestorableTime}'
```

**Recommended settings:**
- Backup retention: 14 days
- Deletion protection: enabled

### List recent snapshots
```bash
aws rds describe-db-snapshots --db-instance-identifier $RDS_INSTANCE \
  --query 'DBSnapshots[*].{ID:DBSnapshotIdentifier,Created:SnapshotCreateTime,Status:Status}' \
  --output table
```

### Restore drill (test quarterly)
```bash
# 1. Restore to a temporary instance
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier "restore-drill-$(date +%Y%m%d)" \
  --db-snapshot-identifier $LATEST_SNAPSHOT

# 2. Verify data integrity on the temp instance

# 3. Delete the temp instance when done
aws rds delete-db-instance \
  --db-instance-identifier "restore-drill-$(date +%Y%m%d)" \
  --skip-final-snapshot
```
