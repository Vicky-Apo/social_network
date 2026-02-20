# Groups API (Frontend Guide)

This document explains how to create groups, browse groups, manage invitations, and handle join requests.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions

All group endpoints require a valid session cookie. Use `credentials: "include"` in the frontend.

## Endpoints

### Create group

`POST /groups`

Request body (JSON):

```json
{
  "title": "Go Builders",
  "description": "A group for Go enthusiasts"
}
```

Response (201):

```json
{
  "success": true,
  "data": {
    "id": 1,
    "creator_id": 10,
    "title": "Go Builders",
    "description": "A group for Go enthusiasts",
    "member_count": 1,
    "is_member": true,
    "created_at": "2025-01-24T12:34:56Z",
    "updated_at": "2025-01-24T12:34:56Z"
  }
}
```

### List groups (browse/search)

`GET /groups?limit=20&offset=0`

Optional search:

`GET /groups?q=go&limit=20&offset=0`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "creator_id": 10,
      "title": "Go Builders",
      "description": "A group for Go enthusiasts",
      "member_count": 3,
      "is_member": false,
      "created_at": "2025-01-24T12:34:56Z",
      "updated_at": "2025-01-24T12:34:56Z"
    }
  ]
}
```

### Get group by ID

`GET /groups/{id}`

Notes:
- Accessible to any authenticated user for discovery. Member-only data is exposed via other endpoints (members, posts, events, chat).

### List group members

`GET /groups/{id}/members`

Notes:
- Only group members can access the member list.
- Returns `404` if the group does not exist.

### Invite user to group

`POST /groups/{id}/invitations`

Request body (JSON):

```json
{
  "invitee_id": 22
}
```

### List invitations for me

`GET /group-invitations`

### Accept or decline invitation

`PATCH /group-invitations/{id}`

Request body (JSON):

```json
{
  "status": "accepted"
}
```

Allowed values: `accepted`, `declined`.

### Request to join group

`POST /groups/{id}/join-requests`

Notes:
- Returns `404` if the group does not exist.

### List join requests (creator only)

`GET /groups/{id}/join-requests`

Notes:
- Returns `404` if the group does not exist.

### Accept or decline join request

`PATCH /group-join-requests/{id}`

Request body (JSON):

```json
{
  "status": "accepted"
}
```

Allowed values: `accepted`, `declined`.

### Leave group

`DELETE /groups/{id}/members/me`

Notes:
- Group creator cannot leave their own group.

## React fetch example

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL;

export async function listGroups(q = "", limit = 20, offset = 0) {
  const params = new URLSearchParams({ limit: String(limit), offset: String(offset) });
  if (q) params.set("q", q);
  const res = await fetch(`${API_BASE}/groups?${params.toString()}`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("List groups failed");
  return res.json();
}
```
