# Profile API (Frontend Guide)

This document explains profile access, followers/following lists, and visibility updates.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions

All profile endpoints require a valid session cookie. Use `credentials: "include"` in the frontend.

## Endpoints

### Get profile

`GET /profiles/{id}`

Response (200):

```json
{
  "success": true,
  "data": {
    "user": {
      "id": 2,
      "email": "jane@example.com",
      "first_name": "Jane",
      "last_name": "Doe",
      "date_of_birth": "31/12/2000",
      "nickname": "jdoe",
      "about": "Hi there",
      "is_public": true,
      "created_at": "2025-01-24T12:34:56Z",
      "updated_at": "2025-01-24T12:34:56Z"
    },
    "followers_count": 10,
    "following_count": 5,
    "is_following": true,
    "is_followed_by": false
  }
}
```

Notes:
- If the profile is private, only followers or the profile owner can access it.
- `date_of_birth` format is `DD/MM/YYYY`.

### List followers

`GET /profiles/{id}/followers`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "email": "john@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "date_of_birth": "01/01/1999",
      "nickname": null,
      "about": null,
      "is_public": true,
      "created_at": "2025-01-24T12:34:56Z",
      "updated_at": "2025-01-24T12:34:56Z"
    }
  ]
}
```

### List following

`GET /profiles/{id}/following`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 3,
      "email": "alice@example.com",
      "first_name": "Alice",
      "last_name": "Smith",
      "date_of_birth": "02/02/2001",
      "nickname": null,
      "about": null,
      "is_public": true,
      "created_at": "2025-01-24T12:34:56Z",
      "updated_at": "2025-01-24T12:34:56Z"
    }
  ]
}
```

### Update visibility

`PATCH /profiles/{id}/visibility`

Request body (JSON):

```json
{
  "is_public": true
}
```

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "updated",
    "is_public": true
  }
}
```

## React fetch example

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL;

export async function getProfile(id: number) {
  const res = await fetch(`${API_BASE}/profiles/${id}`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("Profile fetch failed");
  return res.json();
}

export async function listFollowers(id: number) {
  const res = await fetch(`${API_BASE}/profiles/${id}/followers`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("Followers fetch failed");
  return res.json();
}

export async function listFollowing(id: number) {
  const res = await fetch(`${API_BASE}/profiles/${id}/following`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("Following fetch failed");
  return res.json();
}

export async function updateVisibility(id: number, isPublic: boolean) {
  const res = await fetch(`${API_BASE}/profiles/${id}/visibility`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ is_public: isPublic }),
  });
  if (!res.ok) throw new Error("Visibility update failed");
}
```

## Notes

- Only the profile owner can update visibility.
- User activity and posts will be added to profile responses later.
