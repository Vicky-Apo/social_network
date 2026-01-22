## Profile APIs (backend)

These endpoints are currently **unauthenticated**. Access control uses an optional `viewer_id` query param.

Base URL (local): `http://localhost:8080`

### GET /profiles/{id}
Fetch a profile with follower stats.

Query params:
- `viewer_id` (optional): the requesting user ID

Examples:
- Public profile: `GET /profiles/2`
- Private profile (must be follower or self): `GET /profiles/2?viewer_id=1`

Responses:
- 200 OK:
```json
{
  "user": {
    "id": 2,
    "email": "user@example.com",
    "first_name": "Jane",
    "last_name": "Doe",
    "date_of_birth": "2000-01-02T00:00:00Z",
    "avatar_path": "/uploads/avatar.png",
    "nickname": "jdoe",
    "about": "Hello",
    "is_public": true,
    "created_at": "2026-01-22T21:50:00Z",
    "updated_at": "2026-01-22T21:50:00Z"
  },
  "followers_count": 3,
  "following_count": 5
}
```
- 403 Forbidden (private profile and viewer has no access)
- 404 Not Found

### GET /profiles/{id}/followers
List followers for a profile (private profiles require access).

Query params:
- `viewer_id` (optional)

Responses:
- 200 OK: array of users
- 403 Forbidden
- 404 Not Found

### GET /profiles/{id}/following
List users that the profile is following (private profiles require access).

Query params:
- `viewer_id` (optional)

Responses:
- 200 OK: array of users
- 403 Forbidden
- 404 Not Found

### PATCH /profiles/{id}/visibility
Update the public/private flag for a profile.

Request body:
```json
{
  "is_public": true
}
```

Responses:
- 204 No Content
- 404 Not Found
