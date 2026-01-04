# Security Policy

## Reporting a vulnerability

Please do not open a public GitHub issue for security-sensitive reports.

Instead, use one of the following:

- GitHub Security Advisories (preferred): open a private report from the repo’s Security tab.
- If Security Advisories are not available, contact the maintainer privately (e.g. via GitHub profile contact info).

## Scope

This repo includes:

- Go backend (`backend/`)
- Frontend (`frontend/`)
- Data-science tooling (`data-science/`)

## Coordinated disclosure

- We’ll acknowledge receipt as soon as possible.
- We’ll work with you on a reasonable disclosure timeline.

## Secrets handling

- Do not include secrets, tokens, credentials, or real `.env` contents in reports, PRs, commits, screenshots, or logs.
- Use `.env.example` for placeholders and keep real env files local (`.env`, `.env.local`, `.envrc` are gitignored).
- If you suspect a secret was exposed, rotate it and document the rotation without recording the secret value.
