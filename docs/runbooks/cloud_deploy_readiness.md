# Cloud Deploy Readiness (AWS)

This doc is the checklist/runbook to take Calcutta from local-only development to a production AWS deployment suitable for competition operations.

Assumptions:
- Frontend is static hosting.
- Production is mostly read-only for most users.
- Admins will update game state and Calcutta owners can edit entries occasionally.

---

## Decisions (current)

- [x] Backend compute: ECS Fargate behind an ALB
- [x] Frontend hosting: S3 + CloudFront + Route53 + ACM
- [x] Auth: Cognito User Pool (staging/prod) + local dev auth mode (offline)
- [x] Infrastructure as Code: Terraform

---

## Why AWS (and when Fly/Render/Supabase are compelling)

AWS is a good default when you want:
- Fine-grained control over networking, security boundaries, and scaling.
- Long-term ownership of the platform.
- Standard building blocks (RDS, ALB, CloudFront, S3, IAM, CloudWatch) that map well to "boring" production architecture.

Fly/Render/Supabase can still be compelling if you optimize for speed-to-prod and small-team ops:
- Fly/Render:
  - Great developer experience for simple web backends.
  - Lower operational overhead than ECS.
  - Good choice if you want to avoid learning/maintaining AWS networking/IAM.
- Supabase:
  - Strong option if you want Postgres + auth + storage quickly.
  - Works well if your backend is mostly thin CRUD on Postgres.

Given you’re comfortable with AWS and you want reliability for a time-bound competition, AWS is appropriate. The main risk with AWS is over-building (too many moving parts). This plan intentionally keeps the architecture minimal.

---

## Target AWS architecture (minimal, production-appropriate)

Frontend (static):
- S3 bucket (private)
- CloudFront distribution
- ACM certificate (TLS)
- Route53 DNS

Backend API:
- ECS Fargate service (single service) behind an ALB
- ECR for container images
- CloudWatch logs
- Optional: AWS WAF on ALB/CloudFront

Database:
- RDS Postgres (managed)
- Automated backups + retention policy
- Parameter group / connection limits tuned for your API

Secrets/config:
- SSM Parameter Store or Secrets Manager

Environments:
- `local` (required)
- `staging` (required)
- `prod` (required)

---

## Recommended auth approach (pragmatic + testable)

Recommendation: use Cognito for `staging`/`prod`, but do not build/own a full auth system (password storage, email verification, account recovery).

Instead, implement auth in the backend behind a small interface so local dev can run fully offline:
- `staging`/`prod`: Cognito ID token verification (JWKS/RS256)
- `local`: `AUTH_MODE=dev` accepting `X-Dev-User: <user_id>` for offline dev

Backend switch:
- `AUTH_MODE=legacy`: existing Calcutta JWT/session auth
- `AUTH_MODE=dev`: dev header auth (`X-Dev-User`) + existing API key auth
- `AUTH_MODE=cognito`: Cognito ID token auth + existing API key auth

Required Cognito env vars when `AUTH_MODE=cognito`:
- `COGNITO_REGION`
- `COGNITO_USER_POOL_ID`
- `COGNITO_APP_CLIENT_ID`

Optional Cognito provisioning behavior:
- `COGNITO_AUTO_PROVISION=true` will create a `users` row on first successful Cognito login (keyed by email) so DB-backed permissions work
- `COGNITO_ALLOW_UNPROVISIONED=true` will allow authentication to succeed even if a `users` row is missing (use with caution)

This gives you:
- Offline local dev and local tests.
- Real production-grade auth with fewer footguns.
- A small delta between local and cloud (same middleware, different verifier).

---

## Infrastructure as Code (IaC) recommendation (Terraform vs CDK vs CloudFormation)

Recommendation: Terraform for AWS infra.

Rationale:
- Terraform keeps infra definitions in a widely-used, vendor-neutral toolchain.
- In practice, "take our ball and go home" is only partially true (your resources are still AWS-specific), but Terraform makes it easier to:
  - reorganize accounts/environments
  - change CI/CD provider
  - migrate to alternate clouds with less rewrite than CloudFormation/CDK

Notes on alternatives:
- CDK:
  - Strong developer experience.
  - Still produces CloudFormation, and is AWS-first.
  - Good if you want to express infra in code and accept AWS lock-in.
- CloudFormation:
  - Very stable, AWS-native.
  - Verbose; tends to slow iteration for small teams.

Operational guidance:
- Use Terraform for long-lived infra (VPC/networking, RDS, ALB/ECS service shell, S3/CloudFront, IAM, SSM/Secrets Manager).
- Keep application deploys (new image tags, task definition revision) owned by CI/CD to avoid Terraform being the bottleneck for every deploy.

---

# Non-negotiables (must be done before first real cloud deploy)

## Config + secrets contract
- [x] Backend: all configuration is driven by environment variables (no implicit localhost assumptions)
- [x] Backend: required env vars are validated on startup (fail fast with clear error)
- [ ] No secrets committed to git (repo + CI)
- [ ] Use AWS SSM Parameter Store or Secrets Manager for prod/staging secrets
- [ ] Rotate any secrets that have ever lived in local `.env` files once cloud deploy starts

## Database migrations are the only schema change mechanism
- [ ] All schema changes are versioned migrations committed to the repo
- [ ] Deploy pipeline runs migrations as a gated step (or as a one-off “release job”)
- [ ] Staging migrations must run automatically (to catch issues before prod)
- [ ] Document the migration procedure and rollback expectations

## Health, timeouts, and safe shutdown
- [x] Add/confirm `GET /health/live` endpoint (process is alive)
- [x] Add/confirm `GET /health/ready` endpoint (DB reachable, migrations compatible)
- [ ] HTTP server timeouts are configured (read/write/idle)
- [ ] DB connection pool has explicit limits
- [ ] Service exits cleanly on SIGTERM (ECS rolling deploys rely on this)

## Authentication + authorization for admin/write paths
- [ ] Define user roles: `admin`, `owner`, `viewer`
- [ ] Protect all write endpoints (admin/game state updates; owner entry edits)
- [x] Choose auth: Cognito User Pool for `staging`/`prod`
- [x] Implement backend auth behind an interface (Cognito verifier + local dev verifier)
- [ ] Decide how roles are represented in JWT claims (Cognito Groups vs custom claims)
- [ ] Explicit CORS allowlist for the CloudFront frontend domain
- [ ] Enforce TLS everywhere (CloudFront + ALB)
- [ ] Add basic rate limiting / request size limits to reduce abuse surface

## Backups + restore drill
- [ ] Enable automated RDS backups (with defined retention)
- [ ] Confirm PITR (point-in-time recovery) is available/configured (if supported in chosen setup)
- [ ] Perform a restore drill into staging (practice the process before March)
- [ ] Define RPO/RTO targets (even if modest)

## Observability baseline
- [ ] Structured logs (JSON) including request IDs
- [ ] Error logging at boundaries (handlers/CLIs), not deep in services
- [ ] Metrics for:
  - [ ] request count / latency (p50/p95)
  - [ ] 4xx/5xx rate
  - [ ] DB query latency + pool saturation
- [ ] Alerts for:
  - [ ] high 5xx rate
  - [ ] high p95 latency
  - [ ] container restart loop / unhealthy tasks
  - [ ] DB connection exhaustion / CPU saturation

---

# Keep local dev simple while keeping cloud delta small

## Make Docker the shared contract
- [ ] Backend Dockerfile builds a production image deterministically (multi-stage build)
- [ ] Backend container runs as non-root
- [ ] Local `docker-compose.yml` is the reference for dependencies (Postgres, etc.)
- [ ] Same environment variables are used locally and in AWS (only values differ)

## One-command workflows
- [ ] `make dev` starts dependencies and runs the app locally
- [ ] `make test` runs unit tests (no DB/integration tests per repo standards)
- [ ] `make lint` / `make fmt` if applicable
- [ ] `make build` produces the same artifacts CI produces

## Staging is “prod with smaller scale”
- [ ] Staging uses the same deploy mechanism as prod (same ECS task definition shape)
- [ ] CI deploys to staging on merge to main (or on release branch)
- [ ] Prod deploy is a promotion of the same artifact (image tag)

## Reduce production surface area (competition mode)
- [ ] Keep production API read-heavy by design
- [ ] Gate admin-only operations behind auth and an allowlist if necessary
- [ ] If data science pipelines exist, run them outside of prod request path (scheduled jobs)

---

# CI/CD (AWS)

This plan assumes GitHub Actions.

## Build and publish
- [ ] CI builds backend container image and pushes to ECR
- [ ] CI builds frontend static assets and publishes to S3 (with cache-busting)
- [ ] CI invalidates CloudFront cache on frontend deploy

## Deploy
- [ ] CI updates ECS service (rolling deploy)
- [ ] CI runs DB migrations as a discrete step (before ECS deploy, or as a release job)
- [ ] CI has separate targets for staging and prod

## Access control
- [ ] CI uses a dedicated AWS IAM role with least privilege
- [ ] No long-lived AWS keys stored in CI (use OIDC / role assumption)

GitHub Actions:
- [ ] Configure AWS OIDC provider + IAM roles per environment (`staging`, `prod`)
- [ ] Staging deploy: automatic on merge to `main`
- [ ] Prod deploy: manual approval / protected environment

## Infrastructure as Code
- [ ] Add Terraform configuration for `staging` and `prod`
- [ ] Use separate state per environment (separate backend/state key)
- [ ] Ensure Terraform can create infra from scratch in a new AWS account

---

# AWS Implementation choices (recommended defaults)

Backend compute (pick one):
- [x] ECS Fargate + ALB (recommended default)
- [ ] App Runner (simpler than ECS, less control)

Frontend:
- [ ] S3 + CloudFront + Route53 + ACM

Database:
- [ ] RDS Postgres (single instance to start; Multi-AZ optional)

Secrets:
- [ ] SSM Parameter Store (simple) or Secrets Manager (rotation features)

---

# Pre-March “game day” rehearsal

- [ ] Do a full staging cutover rehearsal:
  - [ ] deploy backend
  - [ ] deploy frontend
  - [ ] run migrations
  - [ ] restore a DB snapshot into staging
- [ ] Load test key read endpoints (basic concurrency + burst)
- [ ] Practice an incident response:
  - [ ] simulate 5xx spike and verify alerts
  - [ ] roll back to previous image
  - [ ] confirm logs and dashboards make diagnosis possible

---

# Open questions (fill in before implementation)

- [ ] How will users be provisioned? (invite-only vs self-serve signup)
- [ ] Do we require verified email in prod? (recommended: yes)
- [ ] What are the admin workflows exactly (game state updates)?
- [ ] What are the owner edit workflows exactly (entry edits) and how do we prevent abuse?
- [ ] What are the expected traffic levels during competition?
- [ ] What is the source of truth for game state (manual admin entry vs ingestion)?
