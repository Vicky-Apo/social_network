# Posts API (Frontend Guide)

This document explains how to list and create posts, and how to list posts by user or group.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions

Listing and creating posts require a valid session cookie. Use `credentials: "include"` in the frontend.

## Endpoints

### List posts

`GET /posts`

Query parameters:
- `limit` (optional, default 10, max 100)
- `offset` (optional, default 0)
- `groups_only` (optional, `true`/`false`): return only group posts from groups the user belongs to
- `author_id` (optional): list posts by a specific user (see "List posts by user")
- Response header: `X-Total-Count` (total number of posts for pagination)

Notes:
- Returns non-group posts the user can see (public/followers/private rules)
  plus group posts from groups the user is a member of.
- If `groups_only=true`, only group posts are returned.
- Author visibility is checked once per request using ownership/public/follower status to avoid redundant DB lookups.
- Feed visibility is computed with a shared SQL CTE to keep logic consistent and reduce repeated joins.

Error responses:
- `400 Bad Request` - Invalid pagination parameters
- `401 Unauthorized` - Not logged in or invalid session

Response (200):

```json
{
  "success": true,
  "data": [
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
```

Headers:
- `X-Total-Count`: total posts available for the current query.

### Create post

`POST /posts`

Request body (JSON):

```json
{
  "content": "My post",
  "media_path": "/uploads/cat.gif",
  "privacy": "public",
  "group_id": null,
  "allowed_user_ids": [5, 8]
}
```

Notes:
- `content` or `media_path` is required (one can be empty, not both).
- `privacy` must be `public`, `followers`, or `private`.
- `followers` is the "almost private" option (only followers can see the post).
- `allowed_user_ids` is required only when `privacy` is `private` (must be followers of the author).
- `allowed_user_ids` is ignored for `public` and `followers`.
- `group_id` is optional here. For group posts, prefer `POST /groups/{id}/posts` and do not send `allowed_user_ids`.
- If `group_id` is provided, the post is stored with `privacy = public` and access is enforced by group membership.
- Use `POST /uploads` to get a `media_path` if you need to attach an image/GIF.

Response (201):

```json
{
  "success": true,
  "data": {
    "id": 10,
    "author_id": 2,
    "group_id": null,
    "author_first_name": "Jane",
    "author_last_name": "Doe",
    "author_nickname": "jdoe",
    "author_avatar_path": "/uploads/avatars/jane.png",
    "content": "My post",
    "media_path": "/uploads/cat.gif",
    "privacy": "public",
    "comment_count": 0,
    "like_count": 0,
    "dislike_count": 0,
    "created_at": "2025-01-24T12:34:56Z",
    "updated_at": "2025-01-24T12:34:56Z"
  }
}
```

Error responses:
- `400 Bad Request` - Invalid request body (bad privacy, missing content/media, invalid allowed_user_ids, invalid group_id)
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - Not allowed to post in the group
- `404 Not Found` - Group not found (for group posts)

### Update post

`PATCH /posts/{id}`

Request body (JSON):

```json
{
  "content": "Updated content",
  "media_path": "/uploads/new.png",
  "privacy": "followers",
  "allowed_user_ids": [5, 8]
}
```

Notes:
- Only the post author can update.
- `content` or `media_path` must be present after update (one can be empty, not both).
- `privacy` can be `public`, `followers`, or `private`.
- `allowed_user_ids` is only used for `private` posts.
- Group posts cannot change privacy or allowed users.

Response (200):

```json
{
  "success": true,
  "data": {
    "id": 10,
    "author_id": 2,
    "group_id": null,
    "author_first_name": "Jane",
    "author_last_name": "Doe",
    "author_nickname": "jdoe",
    "author_avatar_path": "/uploads/avatars/jane.png",
    "content": "Updated content",
    "media_path": "/uploads/new.png",
    "privacy": "followers",
    "comment_count": 0,
    "like_count": 0,
    "dislike_count": 0,
    "created_at": "2025-01-24T12:34:56Z",
    "updated_at": "2025-01-24T13:00:00Z"
  }
}
```

Error responses:
- `400 Bad Request` - Invalid request body (bad privacy, missing content/media, invalid allowed_user_ids)
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to update this post
- `404 Not Found` - Post not found

### Delete post

`DELETE /posts/{id}`

Notes:
- Only the post author can delete.

Response (200):

```json
{
  "success": true,
  "data": {
    "status": "deleted"
  }
}
```

Error responses:
- `400 Bad Request` - Invalid post id
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to delete this post
- `404 Not Found` - Post not found

### List posts by user

`GET /posts?author_id={id}&limit=10&offset=0`

Notes:
- Results respect the author's profile privacy and post visibility.

Error responses:
- `400 Bad Request` - Invalid author id or pagination
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not allowed to view this user's posts

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 2,
      "author_id": 2,
      "group_id": null,
      "author_first_name": "Jane",
      "author_last_name": "Doe",
      "author_nickname": "jdoe",
      "author_avatar_path": "/uploads/avatars/jane.png",
      "content": "Second post",
      "media_path": null,
      "privacy": "public",
      "comment_count": 1,
      "like_count": 2,
      "dislike_count": 0,
      "created_at": "2025-01-24T12:34:56Z",
      "updated_at": "2025-01-24T12:34:56Z"
    }
  ]
}
```

### List posts by group

`GET /groups/{id}/posts?limit=10&offset=0`

Notes:
- Only group members can access group posts.
- Returns `404` if the group does not exist.

Error responses:
- `400 Bad Request` - Invalid group id or pagination
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - You are not a member of the group
- `404 Not Found` - Group not found

### Create post in group

`POST /groups/{id}/posts`

Request body (JSON):

```json
{
  "content": "Hello group",
  "media_path": "/uploads/group.png",
  "privacy": "public"
}
```

Notes:
- `group_id` is taken from the URL.
- `allowed_user_ids` is not allowed for group posts.
- `privacy` is still required, but group posts are stored as `public` (group access enforced separately).
- Returns `404` if the group does not exist.

Error responses:
- `400 Bad Request` - Invalid request body (missing content/media, invalid privacy)
- `401 Unauthorized` - Not logged in or invalid session
- `403 Forbidden` - Not allowed to post in the group
- `404 Not Found` - Group not found

## React fetch example

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL;

export async function createPost(payload: {
  content?: string;
  media_path?: string;
  privacy: "public" | "followers" | "private";
  allowed_user_ids?: number[];
}) {
  const res = await fetch(`${API_BASE}/posts`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(payload),
  });
  if (!res.ok) throw new Error("Create post failed");
  return res.json();
}

export async function listUserPosts(userId: number, limit = 20, offset = 0) {
  const res = await fetch(`${API_BASE}/posts?author_id=${userId}&limit=${limit}&offset=${offset}`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("List user posts failed");
  return res.json();
}
```
