# Frontend

React + TypeScript frontend (Vite).

## Requirements

- Node.js 18+
- npm 9+

## Environment

The frontend discovers the backend via:
- `VITE_API_URL` (preferred)
- `VITE_API_BASE_URL` (fallback)

Defaults to `http://localhost:8080`.

## Running locally

```bash
npm install
npm run dev
```

Vite runs on `http://localhost:3000`.

## Notes

- API requests are made via `src/api/apiClient.ts`.
- Access token is stored in `localStorage` and refresh token is stored as an HTTP-only cookie.
