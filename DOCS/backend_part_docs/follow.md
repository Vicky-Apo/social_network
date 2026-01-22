## Follow APIs (backend)

These endpoints are currently **unauthenticated**. The frontend must send actor IDs explicitly in the request body.

Base URL (local): `http://localhost:8080`

### POST /follow-requests
Create a follow request or immediately follow if the target profile is public.

Request body:
```json
{
  "requester_id": 1,
  "target_id": 2
}
```

Responses:
- 200 OK with follow result:
```json
{
  "status": "followed"
}
```
or
```json
{
  "status": "requested",
  "request": {
    "id": 10,
    "requester_id": 1,
    "target_id": 2,
    "created_at": "2026-01-22T21:50:00Z",
    "updated_at": "2026-01-22T21:50:00Z"
  }
}
```
- 400 Bad Request (missing IDs or trying to follow self)
- 404 Not Found (requester or target does not exist)
- 409 Conflict (already following or request already exists)

### POST /follow-requests/{id}/accept
Accept a follow request (only the target can accept).

Request body:
```json
{
  "actor_id": 2
}
```

Responses:
- 204 No Content
- 403 Forbidden (actor is not the target)
- 404 Not Found (request not found)

### POST /follow-requests/{id}/decline
Decline a follow request (only the target can decline).

Request body:
```json
{
  "actor_id": 2
}
```

Responses:
- 204 No Content
- 403 Forbidden
- 404 Not Found

### POST /unfollow
Remove an existing follow relationship.

Request body:
```json
{
  "follower_id": 1,
  "following_id": 2
}
```

Responses:
- 204 No Content
- 400 Bad Request (missing IDs)
