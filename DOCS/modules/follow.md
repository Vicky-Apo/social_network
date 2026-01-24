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
  "status": "requested",
  "request": {
    "id": 10,
    "requester_id": 1,
    "target_id": 2,
    "created_at": "2025-01-24T12:34:56Z",
    "updated_at": "2025-01-24T12:34:56Z"
  }
}
```

If the target profile is public, the response is:

```json
{
  "status": "followed"
}
```

### List pending requests

`GET /follow-requests`

Response (200):

```json
[
  {
    "id": 10,
    "requester_id": 1,
    "target_id": 2,
    "created_at": "2025-01-24T12:34:56Z",
    "updated_at": "2025-01-24T12:34:56Z"
  }
]
```

### Accept follow request

`POST /follow-requests/{id}/accept`

No body required. The authenticated user must be the target of the request.

Response (200):

```json
{
  "status": "accepted"
}
```

### Decline follow request

`POST /follow-requests/{id}/decline`

No body required. The authenticated user must be the target of the request.

Response (200):

```json
{
  "status": "declined"
}
```

### Unfollow

`POST /unfollow`

Request body (JSON):

```json
{
  "following_id": 2
}
```

Response (200):

```json
{
  "status": "unfollowed"
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

export async function acceptRequest(id: number) {
  const res = await fetch(`${API_BASE}/follow-requests/${id}/accept`, {
    method: "POST",
    credentials: "include",
  });
  if (!res.ok) throw new Error("Accept failed");
}

export async function declineRequest(id: number) {
  const res = await fetch(`${API_BASE}/follow-requests/${id}/decline`, {
    method: "POST",
    credentials: "include",
  });
  if (!res.ok) throw new Error("Decline failed");
}

export async function unfollow(followingId: number) {
  const res = await fetch(`${API_BASE}/unfollow`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ following_id: followingId }),
  });
  if (!res.ok) throw new Error("Unfollow failed");
}
```

## Notes

- You cannot follow yourself.
- You cannot unfollow unless you are already following the user.
