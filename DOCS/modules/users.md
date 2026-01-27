# Users API (Frontend Guide)

This document explains how to list and search users for discovery.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions

All user endpoints require a valid session cookie. Use `credentials: "include"` in the frontend.

## Endpoints

### List users (optional search)

`GET /users`

Search:

`GET /users?q=jo`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "first_name": "Jane",
      "last_name": "Doe",
      "nickname": "jdoe",
      "avatar_path": "/uploads/avatars/jane.png"
    }
  ]
}
```

Searches `first_name`, `last_name`, and `nickname` (case-insensitive).

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 3,
      "first_name": "John",
      "last_name": "Smith",
      "nickname": null,
      "avatar_path": null
    }
  ]
}
```

## React fetch example

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL;

export async function listUsers() {
  const res = await fetch(`${API_BASE}/users`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("List users failed");
  return res.json();
}

export async function searchUsers(query: string) {
  const res = await fetch(`${API_BASE}/users?q=${encodeURIComponent(query)}`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("Search users failed");
  return res.json();
}
```
