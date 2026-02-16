# Dry Run Deployment Assessment

**Date**: 2026-02-15
**Status**: Infrastructure gap analysis for go-live rehearsal
**Critical deadline**: 4 weeks to March Madness launch

## Executive Summary

You want to rehearse the complete go-live lifecycle locally before staging and prod deploys. This is excellent practice, but the current infrastructure is **not ready** for even a staging deployment.

**Current state**: Local development works. Production infrastructure does not exist.

**What you need to build**: An entire deployment pipeline or switch to a managed platform.

## Your Desired Dry Run Flow

1. Create a new environment
2. Get admin access to that new environment
3. Seed the database with schools, historical tournaments, and historical calcuttas
4. Invite users to join the platform and claim accounts associated with their old entries
5. Create a new tournament and configure the teams and the bracket
6. Create a new calcutta off of the new tournament
7. Invite users to create entries for the new calcutta
8. Lock the tournament and simulate it playing out

## Critical Questions

### 1. What does "create a new environment" mean practically?

**Current setup:**
- Local dev: `docker-compose.yml` with dev mode containers (hot reload, source mounts)
- Local prod simulation: `docker-compose.local-prod.yml` with production builds
- Real staging/prod: **DOES NOT EXIST**

**What exists for Docker/infra:**

Local development works great:
```bash
make dev              # Start local dev (hot reload)
make prod-up          # Test production builds locally
make prod-ops-migrate # Run migrations on local prod
```

**What does NOT exist:**

1. **No AWS infrastructure** - The Terraform modules are 7-line stubs:
   ```
   /Users/andrewcopp/Developer/Calcutta/infra/terraform/modules/app/main.tf
   ```
   Contains only:
   ```terraform
   variable "env" {
     type = string
   }

   output "env" {
     value = var.env
   }
   ```

2. **No managed database** - Only Docker Postgres exists

3. **No CloudWatch/monitoring** - No metrics, logs, or alarms configured

4. **No secrets management** - Only `.env` files (which should NOT be used in prod)

5. **No DNS/SSL** - No Route53, ACM certificates, or CloudFront distribution

6. **No load balancer** - No ALB configuration

**To create a new environment, you would need:**

Option A: **AWS (2-3 weeks of DevOps work)**
- Write real Terraform modules for VPC, RDS, ECS, ALB, CloudFront, Route53, ACM, IAM roles
- Set up GitHub OIDC provider for CI/CD
- Configure CloudWatch logs and alarms
- Set up SSM Parameter Store or Secrets Manager
- Test migration pipeline
- Write deployment runbooks
- Practice incident response

Option B: **Managed Platform (3-5 days to staging, 1-2 weeks to prod-ready)**
- Render.com or Fly.io for backend (ECS alternative)
- Managed Postgres (Render/Fly/Supabase)
- CloudFlare for frontend static hosting
- Environment variables via platform UI
- Built-in logs/metrics/alerts
- DNS and SSL handled automatically

**Recommendation for 4-week timeline**: Option B (managed platform) is the only realistic path to March Madness launch.

### 2. How should we handle database seeding?

**Current seeding tools (EXCELLENT - already built):**

You have a complete bundle-based seeding system:

```bash
/Users/andrewcopp/Developer/Calcutta/backend/cmd/tools/import-bundles
/Users/andrewcopp/Developer/Calcutta/backend/cmd/tools/export-bundles
/Users/andrewcopp/Developer/Calcutta/backend/cmd/tools/verify-bundles
```

**Existing bundle data:**
```
/Users/andrewcopp/Developer/Calcutta/backend/exports/bundles/
├── schools.json                    # All NCAA schools
├── tournaments/
│   └── ncaa-tournament-*.json     # Historical tournaments (2017-2025)
└── calcuttas/
    └── ncaa-tournament-*/         # Historical calcuttas with entries
```

**How to seed a new environment:**

```bash
# Export from local (if needed)
go run ./cmd/tools/export-bundles -out ./exports/bundles

# Import to new environment
go run ./cmd/tools/import-bundles \
  -in ./exports/bundles \
  -dry-run=false
```

The bundle system includes:
- Schools (slug + name)
- Tournaments (season, bracket structure, game results)
- Calcuttas (entries, bids, ownership, payouts)
- Full referential integrity verification before import

**This is production-ready.** No changes needed.

**What's missing:**
- CLI tool to seed simulations for new environment (exists but not wired into Makefile)
- CLI tool to create first admin user (needs to be built)
- Documentation on seeding order and verification steps

**Recommended seeding workflow:**

```bash
# 1. Run migrations
make ops-migrate

# 2. Import bundles (schools, tournaments, calcuttas)
go run ./cmd/tools/import-bundles -in ./exports/bundles -dry-run=false

# 3. Seed simulation data (optional, for ML features)
go run ./cmd/tools/seed-simulations -n-sims=10000

# 4. Create admin user (needs to be built)
go run ./cmd/tools/create-admin -email=admin@example.com

# 5. Verify
make api-test ENDPOINT=/api/health
make query SQL="SELECT COUNT(*) FROM core.schools"
make query SQL="SELECT COUNT(*) FROM core.tournaments"
```

### 3. What's the current CI/CD situation and how should dry-run testing integrate?

**Current CI/CD (GitHub Actions):**

File: `/Users/andrewcopp/Developer/Calcutta/.github/workflows/backend_ci.yml`
- Runs on PR and push to main
- Executes Go tests
- Enforces code guardrails (no legacy imports, no forbidden patterns)
- **Works great, no changes needed**

File: `/Users/andrewcopp/Developer/Calcutta/.github/workflows/backend_deploy.yml`
- **EXPECTS AWS RESOURCES THAT DO NOT EXIST**
- Assumes ECR, ECS, ALB, task definitions
- Uses OIDC role assumption (good pattern, but roles don't exist)
- Auto-deploys to staging on push to main
- Manual promotion to prod via workflow_dispatch

**GitHub secrets it expects (NONE ARE SET):**
```
AWS_REGION
ECR_REPOSITORY
AWS_ROLE_ARN_STAGING
ECS_CLUSTER_STAGING
ECS_SERVICE_STAGING
ECS_TASK_DEFINITION_STAGING
AWS_ROLE_ARN_PROD
ECS_CLUSTER_PROD
ECS_SERVICE_PROD
ECS_TASK_DEFINITION_PROD
```

**Dry-run testing integration:**

For local dry-run (what you can do NOW):
```bash
# 1. Test production builds locally
make prod-up
make prod-ops-migrate

# 2. Import seed data
go run ./cmd/tools/import-bundles -in ./exports/bundles -dry-run=false

# 3. Test backend health
make api-test ENDPOINT=/api/health

# 4. Test frontend build
cd frontend && npm run build

# 5. Simulate entire deployment locally
make prod-reset  # Fresh start
make prod-up
make prod-ops-migrate
# ... run through your 8-step flow ...
make prod-down
```

For staging dry-run (what you CANNOT do yet):
- Requires a real staging environment (doesn't exist)
- Would need infrastructure provisioned first

**What needs to be built for staging dry-run:**
1. Staging infrastructure (AWS or managed platform)
2. CI/CD secrets configured
3. Migration smoke tests in CI
4. E2E tests that can run against staging
5. Deployment verification script (health checks + DB row counts)

### 4. How do we handle secrets, env vars, and configuration differences?

**Current approach (LOCAL ONLY):**

Environment variables defined in:
```
/Users/andrewcopp/Developer/Calcutta/.env           # Gitignored
/Users/andrewcopp/Developer/Calcutta/.env.example   # Checked in (template)
/Users/andrewcopp/Developer/Calcutta/.env.local     # Gitignored (localhost overrides)
```

Docker Compose loads these via:
```makefile
ENV_DOCKER = set -a; [ -f .env ] && . ./.env; set +a;
```

**Environment-specific config:**

The application properly loads all config from environment variables:
```
DATABASE_URL
AUTH_MODE (legacy/dev/cognito)
SMTP_HOST, SMTP_PORT, SMTP_FROM_EMAIL
ALLOWED_ORIGINS (CORS)
JWT_SECRET
COGNITO_REGION, COGNITO_USER_POOL_ID, COGNITO_APP_CLIENT_ID
HTTP timeouts, DB pool limits, rate limiting
```

This is **correct** - the app is 12-factor compliant.

**What's missing for staging/prod:**

1. **No secrets management** - Prod should use AWS SSM Parameter Store or Secrets Manager
2. **No secret rotation** - Current JWT_SECRET and DB passwords are in git history
3. **No environment-specific configs** - Need separate configs for local/staging/prod

**Recommended approach:**

Local (current):
```bash
.env              # Defaults
.env.local        # Localhost overrides
```

Staging (needs to be built):
```
AWS SSM Parameter Store:
  /calcutta/staging/db_password
  /calcutta/staging/jwt_secret
  /calcutta/staging/smtp_password
  /calcutta/staging/cognito_user_pool_id
```

Prod (needs to be built):
```
AWS SSM Parameter Store:
  /calcutta/prod/db_password
  /calcutta/prod/jwt_secret
  /calcutta/prod/smtp_password
  /calcutta/prod/cognito_user_pool_id
```

**Alternative (managed platform):**
- Render/Fly provide environment variable UI
- Secrets are encrypted at rest
- No need to manage SSM/Secrets Manager

**Critical security gap:**
- All secrets in current `.env` files must be rotated before ANY cloud deploy
- Database passwords, JWT secrets, API keys all need fresh generation

### 5. What's the deployment pipeline look like for staging vs prod?

**Current pipeline (theoretical, not working):**

Staging:
```yaml
# On push to main:
1. Build backend image → Push to ECR
2. Update ECS task definition with new image
3. Force new deployment
4. Wait for service stability
```

Prod:
```yaml
# On manual workflow_dispatch:
1. Use existing ECR image tag from staging
2. Update ECS task definition
3. Force new deployment
4. Wait for service stability
```

**What's missing:**
1. Migration step (where/when do migrations run?)
2. Rollback procedure (how to revert bad deploy?)
3. Health check validation (deploy should fail if /health/ready fails)
4. Smoke tests after deploy
5. Database backup before migration
6. Blue-green or canary deployment strategy

**Recommended deployment flow:**

Staging (automatic on push to main):
```
1. CI runs tests
2. CI builds backend image → push to registry
3. CI builds frontend static assets → push to S3/CDN
4. CI runs migrations as separate job (with backup first)
5. CI updates backend service (rolling deploy)
6. CI waits for health checks
7. CI runs smoke tests
8. CI invalidates CDN cache
9. CI posts deploy notification
```

Prod (manual approval):
```
1. Promote staging image tag to prod
2. Backup database
3. Run migrations
4. Update backend service
5. Wait for health checks
6. Run smoke tests
7. Monitor for 15 minutes
8. Post deploy notification
```

**None of this is built yet.**

### 6. Are there any infrastructure concerns (DNS, SSL, cloud provider setup)?

**DNS:**
- Domain: Not configured
- Route53 hosted zone: Does not exist
- Need to decide: What's the production domain?

**SSL/TLS:**
- ACM certificate: Does not exist
- CloudFront distribution: Does not exist
- Frontend is currently served over HTTP (localhost)

**Cloud provider setup:**

AWS account status: Unknown
- Do you have an AWS account?
- Do you have billing alarms configured?
- Do you have IAM users/roles set up?
- Do you have GitHub OIDC provider configured?

**Cost estimation (AWS):**
- RDS Postgres db.t4g.micro (staging): ~$15/month
- RDS Postgres db.t4g.small (prod): ~$30/month
- ECS Fargate 0.25 vCPU / 0.5 GB (staging): ~$7/month
- ECS Fargate 0.5 vCPU / 1 GB (prod): ~$15/month
- ALB: ~$16/month per environment
- CloudFront: ~$1/month (low traffic)
- Route53: ~$1/month
- **Estimated total: $100-120/month (staging + prod)**

**Alternative (Render):**
- Web service (starter): $7/month
- Postgres (starter): $7/month
- **Total: $28/month (staging + prod = $56/month)**

## What You Can Do RIGHT NOW (Local Dry Run)

Here's your complete local dry-run workflow:

```bash
# 1. Create a new environment (local prod simulation)
make prod-reset  # Stop and remove volumes (fresh start)

# 2. Start new environment
make prod-up
make prod-ops-migrate

# 3. Seed the database
go run ./cmd/tools/import-bundles -in ./exports/bundles -dry-run=false
go run ./cmd/tools/seed-simulations -n-sims=10000

# 4. Get admin access (needs CLI tool - see gap below)
# TODO: Build create-admin CLI tool
# For now, manually insert admin user:
make query SQL="INSERT INTO core.users (id, email, role) VALUES (uuid_generate_v4(), 'admin@test.com', 'admin')"

# 5. Test invite flow
# (Use Mailpit UI at http://localhost:8025)
make api-test ENDPOINT=/api/admin/invites METHOD=POST DATA='{"email":"user@test.com"}'

# 6-8. Test tournament/calcutta/simulation lifecycle
# (Use frontend at http://localhost:3000 or curl)

# Verify everything works
make api-test ENDPOINT=/api/health
make query SQL="SELECT COUNT(*) FROM core.tournaments"
make query SQL="SELECT COUNT(*) FROM core.calcuttas"

# Tear down
make prod-down
```

**This works TODAY.** No infrastructure needed.

## Critical Gaps for Real Staging/Prod

### Infrastructure (BLOCKING)
- [ ] Choose platform: AWS (3 weeks) vs Render/Fly (1 week)
- [ ] Provision staging database
- [ ] Provision staging compute
- [ ] Configure DNS and SSL
- [ ] Set up secrets management
- [ ] Configure monitoring and alerts
- [ ] Write deployment runbooks

### CI/CD (BLOCKING)
- [ ] Set GitHub secrets for deploy
- [ ] Add migration step to deploy pipeline
- [ ] Add health check validation
- [ ] Add smoke tests
- [ ] Add rollback procedure
- [ ] Add deploy notifications

### Application (NICE TO HAVE)
- [ ] Build `create-admin` CLI tool
- [ ] Add deployment verification script
- [ ] Document seeding order
- [ ] Add E2E tests for dry-run flow

### Security (CRITICAL)
- [ ] Rotate all secrets before first cloud deploy
- [ ] Generate new DB passwords
- [ ] Generate new JWT secrets
- [ ] Review CORS allowlist
- [ ] Enable HTTPS everywhere

## Recommendations

Given the 4-week timeline to March Madness:

**Option 1: Managed Platform (RECOMMENDED)**
- Week 1: Set up Render/Fly staging + prod
- Week 2: Test deployment pipeline, run dry-run
- Week 3: Load testing and incident drills
- Week 4: Final rehearsal and go-live prep

**Option 2: AWS (RISKY)**
- Week 1-2: Write Terraform, provision infra
- Week 3: CI/CD pipeline, dry-run testing
- Week 4: Load testing and hope nothing breaks
- **Risk**: No time for iteration if something goes wrong

**Option 3: Local-Only (FALLBACK)**
- Skip staging/prod entirely
- Run the entire competition on your laptop
- Use ngrok or CloudFlare Tunnel for public access
- **Pros**: No infra work needed
- **Cons**: Single point of failure (your laptop)

## Next Steps

1. **DECIDE**: AWS vs managed platform vs local-only
2. **If managed platform**: Set up Render/Fly accounts this week
3. **If AWS**: Start writing Terraform today (you're already behind)
4. **Regardless**: Build the `create-admin` CLI tool
5. **Test local dry-run flow** to validate the 8-step lifecycle

## Key Files Referenced

Infrastructure:
- `/Users/andrewcopp/Developer/Calcutta/docker-compose.yml` - Dev environment
- `/Users/andrewcopp/Developer/Calcutta/docker-compose.local-prod.yml` - Local prod testing
- `/Users/andrewcopp/Developer/Calcutta/Makefile` - Dev productivity commands
- `/Users/andrewcopp/Developer/Calcutta/infra/terraform/` - Terraform stubs (not ready)

CI/CD:
- `/Users/andrewcopp/Developer/Calcutta/.github/workflows/backend_ci.yml` - Tests (works)
- `/Users/andrewcopp/Developer/Calcutta/.github/workflows/backend_deploy.yml` - Deploy (expects AWS)

Seeding:
- `/Users/andrewcopp/Developer/Calcutta/backend/cmd/tools/import-bundles/` - Bundle importer
- `/Users/andrewcopp/Developer/Calcutta/backend/cmd/tools/seed-simulations/` - Simulation seeder
- `/Users/andrewcopp/Developer/Calcutta/backend/exports/bundles/` - Historical data

Configuration:
- `/Users/andrewcopp/Developer/Calcutta/.env.example` - Config template
- `/Users/andrewcopp/Developer/Calcutta/backend/internal/platform/` - Config loading

Documentation:
- `/Users/andrewcopp/Developer/Calcutta/docs/runbooks/cloud_deploy_readiness.md` - AWS checklist
