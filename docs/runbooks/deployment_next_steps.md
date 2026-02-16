# Deployment Next Steps - Decision Framework

**Date**: 2026-02-15
**Deadline**: 4 weeks to March Madness (mid-March 2026)
**Status**: DECISION REQUIRED

## Quick Start - Test Locally RIGHT NOW

You can run the complete dry-run flow locally today:

```bash
# One-command dry run
make dry-run

# Or step-by-step:
make prod-reset              # Fresh start
make prod-up                 # Start production containers
make prod-ops-migrate        # Run migrations
make import-bundles          # Seed historical data
make seed-simulations        # Generate simulation data
make create-admin EMAIL=admin@test.com  # Create admin user

# Test the API
make api-test ENDPOINT=/api/health
make api-test ENDPOINT=/api/tournaments
make api-test ENDPOINT=/api/calcuttas

# Frontend
# Open http://localhost:3000

# Mailpit (test emails)
# Open http://localhost:8025

# Teardown
make prod-down
```

## The Critical Decision: Platform Choice

You need to decide on a deployment platform THIS WEEK.

### Option 1: Managed Platform (Render/Fly/Railway) - RECOMMENDED

**Timeline**: 1-2 weeks to production-ready
**Effort**: LOW
**Cost**: $30-60/month (staging + prod)
**Risk**: LOW

**Pros:**
- Infrastructure managed for you (DB, SSL, DNS, monitoring, backups)
- Deploy in minutes, not weeks
- Built-in CI/CD integration
- Automatic SSL certificates
- Built-in logging and metrics
- Database backups included
- You focus on application code, not infrastructure

**Cons:**
- Less control over infrastructure
- Vendor lock-in (but migration is straightforward)
- Fewer customization options

**Week-by-week plan:**

Week 1 (Feb 15-21):
- [ ] Sign up for Render or Fly.io
- [ ] Create staging environment
- [ ] Deploy backend service
- [ ] Deploy Postgres database
- [ ] Run migrations
- [ ] Test health endpoints
- [ ] Deploy frontend to CloudFlare Pages or Vercel

Week 2 (Feb 22-28):
- [ ] Create production environment
- [ ] Set up custom domain + SSL
- [ ] Test deployment pipeline
- [ ] Run full dry-run in staging
- [ ] Load testing
- [ ] Set up monitoring alerts

Week 3 (Mar 1-7):
- [ ] User acceptance testing
- [ ] Practice incident response
- [ ] Document runbooks
- [ ] Final security review

Week 4 (Mar 8-14):
- [ ] Final rehearsal
- [ ] Go-live prep
- [ ] Launch

**Recommended provider: Render.com**

Why Render:
- Simple, modern UI
- Postgres with automatic backups
- Free SSL certificates
- Built-in CI/CD from GitHub
- Good observability
- Reasonable pricing

**How to get started:**

1. Sign up at render.com
2. Create a new Web Service:
   - Name: calcutta-backend-staging
   - Repository: Your GitHub repo
   - Branch: main
   - Build Command: `cd backend && go build -o bin/api ./cmd/api`
   - Start Command: `./backend/bin/api`
   - Environment: Add all vars from .env.example
3. Create a PostgreSQL database:
   - Name: calcutta-db-staging
   - Plan: Starter ($7/month)
   - Connect to backend service
4. Add Deploy Hook for migrations:
   - Pre-deploy command: `cd backend && go run ./cmd/migrate -up`
5. Deploy frontend to CloudFlare Pages or Vercel (free)

### Option 2: AWS (ECS + RDS + CloudFront) - RISKY FOR TIMELINE

**Timeline**: 2-3 weeks to production-ready
**Effort**: HIGH
**Cost**: $100-150/month (staging + prod)
**Risk**: HIGH (might not finish in time)

**Pros:**
- Full control over infrastructure
- Industry-standard patterns
- Scales to enterprise needs
- Fine-grained security controls
- Good for long-term ownership

**Cons:**
- Significant upfront work (Terraform, VPC, IAM, ALB, ECS, RDS, CloudFront, Route53, ACM)
- Operational complexity (you own the entire stack)
- Debugging is harder (more moving parts)
- 2-3 weeks might not be enough time

**What needs to be built:**

Infrastructure (Terraform):
- [ ] VPC with public/private subnets across 2 AZs
- [ ] RDS Postgres with security groups, parameter groups, backups
- [ ] ECS cluster + Fargate task definitions + service
- [ ] ECR repository for container images
- [ ] Application Load Balancer + target groups + health checks
- [ ] IAM roles for ECS tasks, GitHub Actions
- [ ] CloudFront distribution + S3 bucket for frontend
- [ ] Route53 hosted zone + ACM certificates
- [ ] SSM Parameter Store for secrets
- [ ] CloudWatch log groups + alarms
- [ ] Security groups and network ACLs

CI/CD:
- [ ] GitHub OIDC provider in AWS
- [ ] GitHub secrets for deploy
- [ ] Migration runner in CI
- [ ] Rollback procedures
- [ ] Deploy smoke tests

Operational:
- [ ] Backup/restore runbooks
- [ ] Incident response playbooks
- [ ] Cost monitoring and alarms
- [ ] Security hardening checklist

**Week-by-week plan:**

Week 1 (Feb 15-21):
- [ ] Write Terraform modules (VPC, RDS, ECS, ALB)
- [ ] Provision staging environment
- [ ] Deploy first version to staging
- [ ] Debug networking/security group issues

Week 2 (Feb 22-28):
- [ ] Finish Terraform (CloudFront, Route53, ACM, SSM)
- [ ] Set up CI/CD pipeline
- [ ] Run migrations in CI
- [ ] Test deployment flow

Week 3 (Mar 1-7):
- [ ] Provision production environment
- [ ] Load testing
- [ ] Practice incident response
- [ ] Document runbooks

Week 4 (Mar 8-14):
- [ ] Final rehearsal
- [ ] Go-live prep
- [ ] **Hope nothing breaks**

**Risk assessment:**

This is a LOT of work for 4 weeks. If you hit any roadblocks (networking issues, IAM permission problems, ECS task not starting, ALB health checks failing), you could easily burn a week debugging.

**Only choose AWS if:**
- You have prior AWS/Terraform experience
- You have time to invest in infrastructure
- You're willing to risk missing the deadline
- You need AWS-specific features

### Option 3: Local + ngrok/CloudFlare Tunnel - FALLBACK

**Timeline**: 1 day
**Effort**: MINIMAL
**Cost**: $0-20/month
**Risk**: MEDIUM (single point of failure)

**Pros:**
- Zero infrastructure work
- Deploy TODAY
- Full control (it's your laptop)

**Cons:**
- Your laptop is the server (must stay on and connected)
- No redundancy or failover
- Harder to scale
- Manual SSL certificate management
- Not suitable for high-traffic events

**How it works:**

1. Run production containers locally
2. Use ngrok or CloudFlare Tunnel to expose backend to internet
3. Deploy frontend to CloudFlare Pages or Vercel
4. Point frontend to ngrok URL

**Quick setup:**

```bash
# Start local prod
make prod-up
make prod-ops-migrate

# Expose via ngrok
ngrok http 8080

# Get ngrok URL (e.g., https://abc123.ngrok.io)
# Update frontend VITE_API_URL to ngrok URL
# Deploy frontend to CloudFlare Pages
```

**When to use this:**
- Emergency fallback if infrastructure isn't ready
- Very small competition (< 100 users)
- You're willing to accept the risk of your laptop failing

## Recommendation

**For a 4-week timeline to March Madness: Use Render.com (Option 1)**

Reasoning:
- You can deploy to staging THIS WEEK
- Infrastructure is managed (DB backups, SSL, monitoring)
- Low risk of missing the deadline
- Focus your time on application features and testing
- Can always migrate to AWS later if needed

**If you had 2-3 months: AWS would be the right choice**

AWS is the correct long-term choice for a production system, but you don't have time to build it properly right now.

## What You Can Do Today

1. **Run the local dry-run:**
   ```bash
   make dry-run
   ```

2. **Sign up for Render.com** and create a staging environment

3. **Test the deployment flow** by deploying to Render staging

4. **Document any issues** you encounter

5. **Make the decision** on platform by end of week

## Key Files Created

New tools:
- `/Users/andrewcopp/Developer/Calcutta/scripts/dry-run-local.sh` - Local dry-run script
- `/Users/andrewcopp/Developer/Calcutta/backend/cmd/tools/create-admin/main.go` - Admin user CLI

New Makefile targets:
```bash
make create-admin EMAIL=admin@example.com  # Create admin user
make import-bundles                         # Import historical data
make seed-simulations                       # Generate simulation data
make dry-run                                # Run complete dry-run flow
```

Documentation:
- `/Users/andrewcopp/Developer/Calcutta/docs/runbooks/dry_run_deployment_assessment.md` - Full analysis
- `/Users/andrewcopp/Developer/Calcutta/docs/runbooks/deployment_next_steps.md` - This file

## Questions to Answer This Week

1. **What's the production domain?** (e.g., marchmarkets.com)
2. **Who has access to domain DNS?** (for SSL certificate validation)
3. **What's the expected user count?** (sizing the database)
4. **What's the budget for hosting?** ($30/month vs $150/month)
5. **Do you have AWS experience?** (affects timeline)
6. **What's your risk tolerance?** (stable but less control vs full control but higher risk)

## Get Help

If you're stuck or need to talk through the decision:
- Review: `/Users/andrewcopp/Developer/Calcutta/docs/runbooks/cloud_deploy_readiness.md`
- Run: `make dry-run` to see what works locally
- Test: Deploy to Render staging this weekend
- Decide: Platform choice by Feb 21

---

**Bottom line**: With 4 weeks to launch, you need to choose Render/Fly (Option 1) or accept significant risk. The local dry-run works today, so test that first to validate your deployment flow. Then deploy to Render staging by end of week to de-risk the timeline.
