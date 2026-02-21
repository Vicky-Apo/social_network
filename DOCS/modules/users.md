# Users API (Frontend Guide)

This document explains how to list and search users for discovery.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions

All user endpoints require a valid session cookie. Use `credentials: "include"` in the frontend.

## Endpoints

### List users (optional search + pagination)

`GET /users`

Search:

`GET /users?q=jo`

Pagination:

`GET /users?limit=20&offset=0`

`GET /users?q=jo&limit=20&offset=0`

Notes:
- All users are returned, including private users.
- Results are always limited to lightweight fields (id/name/avatar/nickname).
- The current user is not included in the results.

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

Error responses:
- `400 Bad Request` - Invalid pagination parameters
- `401 Unauthorized` - Not logged in or invalid session
- `500 Internal Server Error` - Failed to list users

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

export async function listUsers(limit = 20, offset = 0) {
  const res = await fetch(`${API_BASE}/users?limit=${limit}&offset=${offset}`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("List users failed");
  return res.json();
}

export async function searchUsers(query: string, limit = 20, offset = 0) {
  const res = await fetch(
    `${API_BASE}/users?q=${encodeURIComponent(query)}&limit=${limit}&offset=${offset}`,
    {
      credentials: "include",
    }
  );
  if (!res.ok) throw new Error("Search users failed");
  return res.json();
}
```
