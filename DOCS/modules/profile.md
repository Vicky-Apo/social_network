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
- Access checks are based on `is_public`, ownership, and follower status and are computed once per request to reduce redundant DB lookups.

Limited response:
- If the profile is private and the viewer is not a follower, the API returns a limited profile with only basic fields (id, first_name, last_name, nickname, avatar_path, is_public) and `limited: true`.

Example limited response (200):

```json
{
  "success": true,
  "data": {
    "user": {
      "id": 2,
      "first_name": "User",
      "last_name": "Beta",
      "nickname": "beta",
      "avatar_path": "/uploads/avatars/beta.png",
      "is_public": false
    },
    "is_following": false,
    "is_followed_by": false,
    "limited": true
  }
}
```

Error responses:
- `400 Bad Request` - Invalid profile id
- `401 Unauthorized` - Not logged in or invalid session
- `404 Not Found` - Profile not found

### Get full profile (profile + posts + activity)

`GET /profiles/{id}/full?limit=10&offset=0&activity_limit=5`

Notes:
- Returns the same `profile` payload as `GET /profiles/{id}` plus:
  - `posts`: paginated list of posts by the user (same shape as `GET /posts?author_id=...`)
  - `activity.recent_posts`: most recent posts (default 5)
- Respects profile privacy rules. If the profile is private and the viewer is not allowed, it will return the limited profile and **empty** `posts`/`activity`.
- `limit`/`offset` control the `posts` list (default limit 10, max 100).
- `activity_limit` controls the size of `activity.recent_posts` (default 5, max 100). Use `activity_limit=0` to omit recent posts.

Response (200):

```json
{
  "success": true,
  "data": {
    "profile": {
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
    },
    "posts": [
      {
        "id": 1,
        "author_id": 2,
        "group_id": null,
        "author_first_name": "Jane",
        "author_last_name": "Doe",
        "author_nickname": "jdoe",
        "author_avatar_path": "/uploads/avatars/jane.png",
        "content": "Hello world",
        "media_path": "/uploads/post-1.png",
        "privacy": "public",
        "comment_count": 3,
        "like_count": 10,
        "dislike_count": 1,
        "created_at": "2025-01-24T12:34:56Z",
        "updated_at": "2025-01-24T12:34:56Z"
      }
    ],
    "activity": {
      "recent_posts": [
        {
          "id": 1,
          "author_id": 2,
          "group_id": null,
          "author_first_name": "Jane",
          "author_last_name": "Doe",
          "author_nickname": "jdoe",
          "author_avatar_path": "/uploads/avatars/jane.png",
          "content": "Hello world",
          "media_path": "/uploads/post-1.png",
          "privacy": "public",
          "comment_count": 3,
          "like_count": 10,
          "dislike_count": 1,
          "created_at": "2025-01-24T12:34:56Z",
          "updated_at": "2025-01-24T12:34:56Z"
        }
      ]
    }
  }
}
```

Error responses:
- `400 Bad Request` - Invalid profile id, pagination, or `activity_limit`
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to view this profile
- `404 Not Found` - Profile not found

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

Notes:
- Access follows the same privacy rules as the profile itself.

Error responses:
- `400 Bad Request` - Invalid profile id
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to view this profile
- `404 Not Found` - Profile not found

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

Notes:
- Access follows the same privacy rules as the profile itself.

Error responses:
- `400 Bad Request` - Invalid profile id
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to view this profile
- `404 Not Found` - Profile not found

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

Error responses:
- `400 Bad Request` - Invalid profile id or request body
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - Only the profile owner can update visibility
- `404 Not Found` - Profile not found

### Update profile

`PATCH /profiles/{id}`

Request body (JSON):

```json
{
  "nickname": "jdoe",
  "about": "Building cool things",
  "avatar_path": "/uploads/avatar/20260212T120000_abcd1234ef567890.png"
}
```

Notes:
- Only the profile owner can update their profile.
- Use `POST /uploads` to get an `avatar_path`.
- Sending empty strings clears a field (sets it to null).

Error responses:
- `400 Bad Request` - Invalid profile id or request body
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - Only the profile owner can update profile
- `404 Not Found` - Profile not found

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

export async function updateProfile(id: number, payload: {
  nickname?: string | null;
  about?: string | null;
  avatar_path?: string | null;
}) {
  const res = await fetch(`${API_BASE}/profiles/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(payload),
  });
  if (!res.ok) throw new Error("Profile update failed");
}
```

## Notes

- Only the profile owner can update visibility.
- User activity and posts are available via the full profile endpoint (below).

### Get profile with posts + activity

`GET /profiles/{id}/full?limit=&offset=&activity_limit=`

Query params:
- `limit` / `offset` for the main `posts` list
- `activity_limit` (optional) for `activity.recent_posts` (default: `5`)

Response (200):

```json
{
  "success": true,
  "data": {
    "profile": {
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
    },
    "posts": [],
    "activity": {
      "recent_posts": []
    }
  }
}
```

Notes:
- If the profile is private and the viewer is not allowed, `posts` and `activity`
  will be empty and `profile.limited` will be `true`.

Error responses:
- `400 Bad Request` - Invalid profile id or invalid pagination/activity_limit
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to view this profile
- `404 Not Found` - Profile not found
