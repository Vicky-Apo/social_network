# Groups API (Frontend Guide)

This document describes the REST API shape for groups, along with requests and responses.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Authentication

All endpoints require a valid session cookie. Use `credentials: "include"` in your requests.

## Models

### Group

```json
{
  "id": 7,
  "title": "Gophers",
  "description": "Go developers club",
  "creator_id": 2,
  "created_at": "2025-02-01T10:00:00Z",
  "updated_at": "2025-02-01T10:00:00Z"
}
```

### Group Member

```json
{
  "group_id": 7,
  "user_id": 3,
  "joined_at": "2025-02-01T10:05:00Z"
}
```

### Group Invitation

```json
{
  "id": 15,
  "group_id": 7,
  "inviter_id": 2,
  "invitee_id": 5,
  "created_at": "2025-02-01T10:10:00Z",
  "updated_at": "2025-02-01T10:10:00Z"
}
```

### Group Join Request

```json
{
  "id": 9,
  "group_id": 7,
  "user_id": 5,
  "created_at": "2025-02-01T10:12:00Z",
  "updated_at": "2025-02-01T10:12:00Z"
}
```

### Group Event

```json
{
  "id": 3,
  "group_id": 7,
  "creator_id": 2,
  "title": "Monthly meetup",
  "description": "Talks and Q&A",
  "event_time": "2025-03-01T18:00:00Z",
  "created_at": "2025-02-01T10:30:00Z",
  "updated_at": "2025-02-01T10:30:00Z"
}
```

## Endpoints

### Create group

`POST /groups`

Request body (JSON):

```json
{
  "title": "Gophers",
  "description": "Go developers club"
}
```

Response (201):

```json
{
  "success": true,
  "data": {
    "group": { ... }
  }
}
```

### List/search groups

`GET /groups?query=go&limit=20&offset=0`

Response (200):

```json
{
  "success": true,
  "data": [
    { ... }
  ]
}
```

Notes:
- Private groups are only returned if the requester is a member.

### Get group by id

`GET /groups/{id}`

Response (200):

```json
{
  "success": true,
  "data": {
    "group": { ... },
    "members_count": 12,
    "is_member": true,
    "role": "member"
  }
}
```

### Update group

`PATCH /groups/{id}`

Request body (JSON):

```json
{
  "title": "New title",
  "description": "Updated",
}
```

Response (200):

```json
{
  "success": true,
  "data": {
    "group": { ... }
  }
}
```

Notes:
- Only the creator or admins can update the group.

### Delete group

`DELETE /groups/{id}`

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "deleted"
  }
}
```

### List group members

`GET /groups/{id}/members?limit=20&offset=0`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 3,
      "first_name": "Jane",
      "last_name": "Doe",
      "nickname": "jdoe",
      "avatar_path": "/uploads/avatars/jane.png"
    }
  ]
}
```

### Leave group

`DELETE /groups/{id}/members/me`

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "left"
  }
}
```

### Remove member (admin)

`DELETE /groups/{id}/members/{user_id}`

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "removed"
  }
}
```

### Invite user

`POST /groups/{id}/invitations`

Request body (JSON):

```json
{
  "invitee_id": 5
}
```

Response (200):

```json
{
  "success": true,
  "data": {
    "invitation": { ... }
  }
}
```

### List received invitations (for me)

`GET /group-invitations?limit=20&offset=0`

Response (200):

```json
{
  "success": true,
  "data": [
    { ... }
  ]
}
```

### Accept/decline invitation

`PATCH /group-invitations/{id}`

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

### Request to join

`POST /groups/{id}/join-requests`

Response (200):

```json
{
  "success": true,
  "data": {
    "request": { ... }
  }
}
```

### List pending join requests (admin)

`GET /groups/{id}/join-requests?limit=20&offset=0`

Response (200):

```json
{
  "success": true,
  "data": [
    { ... }
  ]
}
```

### Approve/decline join request (admin)

`PATCH /groups/{id}/join-requests/{request_id}`

Request body (JSON):

```json
{
  "status": "accepted"
}
```

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "accepted"
  }
}
```

### Create group post

`POST /groups/{id}/posts`

Request body (JSON):

```json
{
  "content": "Hello group",
  "media_path": "/uploads/group-post.png"
}
```

Response (201):

```json
{
  "success": true,
  "data": { ... }
}
```

### List group posts

`GET /groups/{id}/posts?limit=20&offset=0`

Response (200):

```json
{
  "success": true,
  "data": [
    { ... }
  ]
}
```

### Group post comments

`POST /groups/{id}/posts/{post_id}/comments`

`GET /groups/{id}/posts/{post_id}/comments`

`PATCH /group-comments/{id}`

`DELETE /group-comments/{id}`

Notes:
- Same request/response shape as standard comments.
- Distinct routes avoid mixing group vs public posts.

### Create event

`POST /groups/{id}/events`

Request body (JSON):

```json
{
  "title": "Monthly meetup",
  "description": "Talks and Q&A",
  "event_time": "2025-03-01T18:00:00Z"
}
```

Response (201):

```json
{
  "success": true,
  "data": { ... }
}
```

### List events

`GET /groups/{id}/events?limit=20&offset=0`

Response (200):

```json
{
  "success": true,
  "data": [
    { ... }
  ]
}
```

### RSVP to event

`POST /groups/{id}/events/{event_id}/rsvp`

Request body (JSON):

```json
{
  "response": "going"
}
```

Allowed values: `going`, `not_going`.

Response (200):

```json
{
  "success": true,
  "data": {
    "response": "going"
  }
}
```

## Error responses

- `400 Bad Request` for invalid ids or request body.
- `401 Unauthorized` when session is missing or invalid.
- `403 Forbidden` when the requester lacks permission.
- `404 Not Found` for missing entities.
- `409 Conflict` for duplicate requests/invites/memberships.
