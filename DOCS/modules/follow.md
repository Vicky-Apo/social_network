# Follow API (Frontend Guide)

This document explains follow requests, follow acceptance/decline, and unfollow flows.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions

All follow endpoints require a valid session cookie. Use `credentials: "include"` in the frontend.

## Endpoints

### Create follow request

`POST /follow-requests`

Request body (JSON):

```json
{
  "target_id": 2
}
```

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "requested",
    "request": {
      "id": 10,
      "requester_id": 1,
      "target_id": 2,
      "status": "pending",
      "created_at": "2025-01-24T12:34:56Z"
    }
  }
}
```

If the target profile is public, the response is:

```json
{
  "success": true,
  "data": {
    "status": "followed"
  }
}
```

### List pending requests

`GET /follow-requests`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 10,
      "requester_id": 1,
      "target_id": 2,
      "status": "pending",
      "created_at": "2025-01-24T12:34:56Z"
    }
  ]
}
```

Notes:
- Only pending requests are returned.

### List sent requests

`GET /follow-requests/sent`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 11,
      "requester_id": 1,
      "target_id": 3,
      "status": "pending",
      "created_at": "2025-01-24T12:34:56Z"
    }
  ]
}
```

Notes:
- Only pending requests are returned.

### Update follow request status

`PATCH /follow-requests/{id}`

Request body (JSON):

```json
{
  "status": "accepted"
}
```

Allowed values: `accepted`, `declined`.

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "accepted"
  }
}
```

### Unfollow

`DELETE /users/{id}/followers`

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "unfollowed"
  }
}
```

## React fetch example

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL;

export async function requestFollow(targetId: number) {
  const res = await fetch(`${API_BASE}/follow-requests`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ target_id: targetId }),
  });
  if (!res.ok) throw new Error("Follow request failed");
  return res.json();
}

export async function listFollowRequests() {
  const res = await fetch(`${API_BASE}/follow-requests`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("List requests failed");
  return res.json();
}

export async function updateFollowRequest(id: number, status: "accepted" | "declined") {
  const res = await fetch(`${API_BASE}/follow-requests/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ status }),
  });
  if (!res.ok) throw new Error("Update request failed");
}

export async function unfollow(userId: number) {
  const res = await fetch(`${API_BASE}/users/${userId}/followers`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!res.ok) throw new Error("Unfollow failed");
}
```

## Notes

- You cannot follow yourself.
- You cannot unfollow unless you are already following the user.
