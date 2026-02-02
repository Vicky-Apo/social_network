# Frontend Agent Instructions

These instructions apply to work inside `frontend/`.

## Goals
- Build the client UI for the social network using Next.js.
- Communicate with the expected backend API over HTTP using sessions + cookies.
- Keep the frontend aligned with Postgres-backed API contracts (no SQLite-specific assumptions).

## Backend Contract (Expected)
- Base URL: `NEXT_PUBLIC_API_BASE_URL` (default to `http://localhost:8080` in code if unset).
- Auth endpoints:
  - `POST /auth/register` expects JSON:
    - `email`, `password`, `first_name`, `last_name`, `date_of_birth` (format `DD/MM/YYYY`)
    - optional: `nickname`, `about`
  - `POST /auth/login` expects JSON: `email`, `password`
  - `POST /auth/logout` uses cookie session
  - `GET /auth/me` uses cookie session
- Responses use a wrapper:
  - success: `{ "success": true, "data": ... }`
  - error: `{ "success": false, "error": "..." }`
- Cookies are HttpOnly, so frontend must send `credentials: "include"` for auth-protected calls.

## Frontend Conventions
- Use the App Router (`src/app/...`).
- Keep pages client-side only when required (forms, auth flows).
- Store auth session in context for UI state only; the source of truth is the server session cookie.
- Use snake_case when sending JSON to the backend.
- Validate inputs client-side but rely on backend errors for final messaging.

## Files to Start From
- `src/app/login/page.tsx`
- `src/app/register/page.tsx`
- `src/app/component/AuthContext.tsx`

## For the "Main Page When Logged In"
- First call `GET /auth/me` on load to determine authenticated user.
- If unauthorized, redirect to `/login`.
- Show basic feed shell: header + left nav + main feed + right rail.
- Fetch posts from `/posts` with `credentials: "include"`.
- Prefer composable components under `src/app/component/`.

## Styling
- Use existing Tailwind setup.
- Keep UI clean and responsive (mobile-first, then desktop grids).

## Non-Goals (for now)
- Do not implement websocket chat UI yet.
- Do not add avatar upload unless backend supports multipart.
