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

Error responses:
- `400 Bad Request` - Missing/invalid `target_id`, or trying to follow yourself
- `401 Unauthorized` - Not logged in or invalid session
- `404 Not Found` - Target user not found
- `409 Conflict` - Already following or a pending request already exists

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
      "requester": {
        "id": 1,
        "first_name": "John",
        "last_name": "Doe",
        "nickname": "jdoe",
        "avatar_path": "/uploads/avatars/john.png"
      },
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
      "target": {
        "id": 3,
        "first_name": "Anna",
        "last_name": "Smith",
        "nickname": "asmith",
        "avatar_path": "/uploads/avatars/anna.png"
      },
      "status": "pending",
      "created_at": "2025-01-24T12:34:56Z"
    }
  ]
}
```

Notes:
- Only pending requests are returned.
- Follow request responses include requester/target user details to avoid extra `/users` fetches.

### Update follow request status

`PATCH /follow-requests/{id}`

Request body (JSON):

```json
{
  "status": "accepted"
}
```

Allowed values: `accepted`, `declined`, `canceled`.

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "accepted"
  }
}
```

Error responses:
- `400 Bad Request` - Invalid status
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to update this request
- `404 Not Found` - Follow request not found
- `409 Conflict` - Request is not pending

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

Error responses:
- `400 Bad Request` - Invalid user id
- `401 Unauthorized` - Not logged in or invalid session
- `404 Not Found` - User not found
- `409 Conflict` - You are not currently following this user

### Remove follower

`DELETE /followers/{id}`

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "removed"
  }
}
```

Error responses:
- `400 Bad Request` - Invalid user id
- `401 Unauthorized` - Not logged in or invalid session
- `404 Not Found` - User not found
- `409 Conflict` - The user is not your follower

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

export async function updateFollowRequest(id: number, status: "accepted" | "declined" | "canceled") {
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

export async function removeFollower(userId: number) {
  const res = await fetch(`${API_BASE}/followers/${userId}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!res.ok) throw new Error("Remove follower failed");
}
```

## Notes

- You cannot follow yourself.
- You cannot unfollow unless you are already following the user.
