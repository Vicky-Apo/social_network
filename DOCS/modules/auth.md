# Auth API (Frontend Guide)

This document explains how the authentication endpoints work and how to use them from a React frontend.

## Base URL

Set a base URL for the backend in the frontend environment, for example:

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions (how it works)

- On successful login, the backend sets an HttpOnly session cookie.
- The frontend should send requests with `credentials: "include"` so the cookie is sent.
- Protected routes require a valid session cookie (no bearer tokens).
- Logout clears the cookie and deletes the session server-side if it exists.

## Endpoints

### Register

`POST /auth/register`

Request body (JSON):

```json
{
  "email": "jane@example.com",
  "password": "supersecret",
  "first_name": "Jane",
  "last_name": "Doe",
  "date_of_birth": "31/12/2000",
  "avatar_path": "/uploads/avatar/jane.png",
  "nickname": "jdoe",
  "about": "Hi there"
}
```

Notes:
- `date_of_birth` format is `DD/MM/YYYY`.
- `avatar_path`, `nickname`, and `about` are optional and can be omitted or set to `null`.

Response (201):

```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "jane@example.com",
    "first_name": "Jane",
    "last_name": "Doe",
    "date_of_birth": "31/12/2000",
    "avatar_path": "/uploads/avatar/jane.png",
    "nickname": "jdoe",
    "about": "Hi there",
    "is_public": false,
    "created_at": "2025-01-24T12:34:56Z"
  }
}
```

### Login

`POST /auth/login`

Request body (JSON):

```json
{
  "email": "jane@example.com",
  "password": "supersecret"
}
```

Response (200):

```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "email": "jane@example.com",
      "first_name": "Jane",
      "last_name": "Doe",
      "date_of_birth": "31/12/2000",
      "avatar_path": "/uploads/avatar/jane.png",
      "nickname": "jdoe",
      "about": "Hi there",
      "is_public": false,
      "created_at": "2025-01-24T12:34:56Z"
    }
  }
}
```

Important:
- The cookie is the source of truth for auth.
- Protected endpoints do not accept Authorization headers.

### Logout

`POST /auth/logout`

No body required. This clears the session cookie.

Response (200):

```json
{
  "success": true,
  "data": null
}
```

### Current user

`GET /auth/me`

Requires a valid session cookie.

Response (200):

```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "jane@example.com",
    "first_name": "Jane",
    "last_name": "Doe",
    "date_of_birth": "31/12/2000",
    "avatar_path": "/uploads/avatar/jane.png",
    "nickname": "jdoe",
    "about": "Hi there",
    "is_public": false,
    "created_at": "2025-01-24T12:34:56Z"
  }
}
```

## React fetch example

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL;

export async function login(email: string, password: string) {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email, password }),
  });

  if (!res.ok) {
    throw new Error("Login failed");
  }

  return res.json();
}

export async function me() {
  const res = await fetch(`${API_BASE}/auth/me`, {
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("Unauthorized");
  }

  return res.json();
}

export async function logout() {
  const res = await fetch(`${API_BASE}/auth/logout`, {
    method: "POST",
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("Logout failed");
  }
}
```

## Axios example

```ts
import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  withCredentials: true,
});

export async function login(email: string, password: string) {
  const { data } = await api.post("/auth/login", { email, password });
  return data;
}

export async function me() {
  const { data } = await api.get("/auth/me");
  return data;
}

export async function logout() {
  await api.post("/auth/logout");
}
```

## CORS notes

For cookies to work in the browser:
- `Access-Control-Allow-Credentials: true` must be enabled (it is in the backend config).
- `Access-Control-Allow-Origin` must match the frontend origin (avoid `*` when credentials are used).
