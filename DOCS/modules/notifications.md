# Notifications API (Frontend Guide)

This document explains how to list and manage notifications.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Authentication

All notification endpoints require a valid session cookie. Use `credentials: "include"` in your requests.

## Endpoints

### List notifications

`GET /notifications?limit=20&offset=0&unread=false`

**Query Parameters:**
- `limit` (optional, default 20, max 100)
- `offset` (optional, default 0)
- `unread` (optional, `true` or `false` — also accepts `1/0/yes/no`)

**Response (200):**

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "user_id": 2,
      "actor_id": 1,
      "type": "follow_request",
      "entity_type": "follow_request",
      "entity_id": 10,
      "metadata": {
        "requester_id": 1
      },
      "is_read": false,
      "read_at": null,
      "created_at": "2025-02-07T15:10:00Z"
    }
  ]
}
```

**Notes:**
- Results are ordered by `created_at DESC`.
- `metadata` is optional and varies by type.
- Notifications are also pushed in real time over WebSocket with type `notification`.

**Error Responses:**
- `400 Bad Request` - Invalid pagination parameters or invalid `unread` value
- `401 Unauthorized` - Not logged in or invalid session

### Unread count

`GET /notifications/unread-count`

**Response (200):**

```json
{
  "success": true,
  "data": {
    "count": 3
  }
}
```

**Error Responses:**
- `401 Unauthorized` - Not logged in or invalid session

### Mark notification as read

`PATCH /notifications/{id}/read`

**Response (200):**

```json
{
  "success": true,
  "data": {
    "status": "read"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid notification id
- `401 Unauthorized` - Not logged in or invalid session
- `404 Not Found` - Notification does not exist or does not belong to the user

### Mark all notifications as read

`PATCH /notifications/read-all`

**Response (200):**

```json
{
  "success": true,
  "data": {
    "updated": 5
  }
}
```

**Error Responses:**
- `401 Unauthorized` - Not logged in or invalid session

## Notification types

The `type` field matches the backend enum values:

- `follow_request`
- `group_invitation`
- `group_join_request`
- `event_created`
- `post_reaction`
- `comment_reaction`
- `comment_on_post`

## Metadata conventions

Example metadata payloads used by the backend:

- `follow_request`: `{ "requester_id": <id> }`
- `group_invitation`: `{ "group_id": <id>, "inviter_id": <id> }`
- `group_join_request`: `{ "group_id": <id>, "user_id": <id> }`
- `event_created`: `{ "group_id": <id>, "title": "Event title", "event_time": "2026-02-14T18:00:00Z" }`
- `post_reaction`: `{ "reaction": "like", "action": "added" }`
- `comment_reaction`: `{ "reaction": "like", "action": "added", "post_id": <id> }`
- `comment_on_post`: `{ "comment_id": <id> }`
