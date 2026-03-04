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

Error responses:
- `400 Bad Request` - Invalid request body or missing/empty title
- `401 Unauthorized` - Not logged in or invalid session


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

Error responses:
- `400 Bad Request` - Invalid pagination parameters
- `401 Unauthorized` - Not logged in or invalid session


### Get group by ID

`GET /groups/{id}`

Notes:
- Accessible to any authenticated user for discovery. Member-only data is exposed via other endpoints (members, posts, events, chat).

Error responses:
- `400 Bad Request` - Invalid group id
- `401 Unauthorized` - Not logged in or invalid session
- `404 Not Found` - Group not found

### List group members

`GET /groups/{id}/members`

Notes:
- Only group members can access the member list.
- Returns `404` if the group does not exist.

Error responses:
- `400 Bad Request` - Invalid group id
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not a member of the group
- `404 Not Found` - Group not found

### Invite user to group

`POST /groups/{id}/invitations`

Request body (JSON):

```json
{
  "invitee_id": 22
}
```

Error responses:
- `400 Bad Request` - Invalid group id or invitee id, or inviting yourself
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to invite to this group
- `404 Not Found` - Group not found
- `409 Conflict` - User is already a member or invitation already exists

Notes:
- The frontend may cache invited user details locally to avoid extra profile lookups for sent invites.
- The create group page should call `/auth/me` once on mount (use `useEffect`) to avoid repeated checks.

### List invitations for me

`GET /group-invitations`

Response includes:
- `group_title` (string)

Notes:
- Use `group_title` to display invitation group names without extra `/groups/{id}` calls.

Error responses:
- `401 Unauthorized` - Not logged in or invalid session

### Accept or decline invitation

`PATCH /group-invitations/{id}`

Request body (JSON):

```json
{
  "status": "accepted"
}
```

Allowed values: `accepted`, `declined`.

Error responses:
- `400 Bad Request` - Invalid status or invalid invitation id
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to update this invitation
- `404 Not Found` - Invitation not found

### Request to join group

`POST /groups/{id}/join-requests`

Notes:
- Returns `404` if the group does not exist.

Error responses:
- `400 Bad Request` - Invalid group id
- `401 Unauthorized` - Not logged in or invalid session
- `404 Not Found` - Group not found
- `409 Conflict` - Already a member, already invited, or request already exists

### List join requests (creator only)

`GET /groups/{id}/join-requests`

Notes:
- Returns `404` if the group does not exist.
- Each join request includes a `user` object with name/avatar to avoid extra profile lookups.

Error responses:
- `400 Bad Request` - Invalid group id
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - Only the group creator can view join requests
- `404 Not Found` - Group not found

### Accept or decline join request

`PATCH /group-join-requests/{id}`

Request body (JSON):

```json
{
  "status": "accepted"
}
```

Allowed values: `accepted`, `declined`.

Error responses:
- `400 Bad Request` - Invalid status or invalid join request id
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - Only the group creator can update join requests
- `404 Not Found` - Join request not found

### Leave group

`DELETE /groups/{id}/members/me`

Notes:
- Group creator cannot leave their own group.

Error responses:
- `400 Bad Request` - Invalid group id
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not a member of the group
- `404 Not Found` - Group not found
- `409 Conflict` - Group creator cannot leave

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
