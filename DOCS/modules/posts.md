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

Notes:
- Returns only non-group posts (global feed).

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

### Get post by ID

`GET /posts/{id}`

Response (200):

```json
{
  "success": true,
  "data": {
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
}
```

### Create post

`POST /posts`

Request body (JSON):

```json
{
  "content": "My post",
  "media_path": "/uploads/cat.gif",
  "privacy": "public",
  "group_id": null,
  "category_ids": [1, 3],
  "allowed_user_ids": [5, 8]
}
```

Notes:
- `content` or `media_path` is required (one can be empty, not both).
- `privacy` must be `public`, `followers`, or `private`.
- `category_ids` is optional.
- `followers` is the "almost private" option (only followers can see the post).
- `allowed_user_ids` is required only when `privacy` is `private` (must be followers of the author).
- `allowed_user_ids` is ignored for `public` and `followers`.
- `group_id` is optional here. For group posts, prefer `POST /groups/{id}/posts` and do not send `category_ids` or `allowed_user_ids`.
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

### List posts by user

`GET /posts?author_id={id}&limit=20&offset=0`

Notes:
- Results respect the author's profile privacy and post visibility.

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

### Filter posts by category

`GET /posts?category_id=1&limit=20&offset=0`

Notes:
- Category filtering applies only to non-group posts.

### List posts by group

`GET /groups/{id}/posts?limit=20&offset=0`

Notes:
- Only group members can access group posts.

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
- `category_ids` and `allowed_user_ids` are not allowed for group posts.
- `privacy` is still required, but `private` is rejected for group posts.

## React fetch example

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL;

export async function createPost(payload: {
  content?: string;
  media_path?: string;
  privacy: "public" | "followers" | "private";
  category_ids?: number[];
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
