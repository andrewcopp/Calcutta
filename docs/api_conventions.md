# API Conventions

This document describes conventions for the HTTP API served by `backend/cmd/api`.

## Base URL

- Local dev default: `http://localhost:8080`
- API prefix: `/api`

## Content type

- Requests and responses use JSON unless otherwise specified.
- Successful responses set `Content-Type: application/json`.

## Error envelope

All error responses follow this shape:

```json
{
  "error": {
    "code": "validation_error",
    "message": "players must bid on a minimum of 3 teams",
    "field": "teams",
    "requestId": "..."
  }
}
```

Where:
- `code` is a stable, machine-readable error code.
- `message` is safe to show to a user.
- `field` is optional and indicates which input field caused the error.
- `requestId` is included for debugging/log correlation.

## Status codes

- `200 OK`
  - Successful GET
  - Successful PATCH when returning a response body
- `201 Created`
  - Successful resource creation
- `204 No Content`
  - Successful mutation when no body is returned
- `400 Bad Request`
  - Invalid request body
  - Validation failures
- `401 Unauthorized`
  - Authentication required or invalid credentials
- `403 Forbidden`
  - Authenticated but not allowed
- `404 Not Found`
  - Resource not found
- `409 Conflict`
  - Uniqueness / already exists / state conflicts
- `423 Locked`
  - Business rule lockouts (example: tournament lock prevents editing bids)
- `500 Internal Server Error`
  - Unhandled server error

## PATCH semantics

- PATCH endpoints update an existing resource.
- Prefer partial update DTOs where fields are pointers/optional.
- If a PATCH mutates a nested collection (e.g. entry bid teams), prefer a transactional replace on the server side.

## Authentication

- Browser clients use access tokens (Authorization: Bearer) and refresh tokens (HTTP-only cookie).
- Some admin/s2s endpoints may use API keys.

## Authorization policy

- Prefer centralized authorization decisions in a policy layer (e.g. `backend/internal/policy/*`).
- Handlers should avoid duplicating permission checks.
