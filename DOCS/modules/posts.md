# Posts API (Frontend Guide)

This document explains how to list and create posts, and how to list posts by user.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions

Listing and creating posts require a valid session cookie. Use `credentials: "include"` in the frontend.

## Endpoints

### List posts

`GET /posts`

Response (200):

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "author_id": 2,
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
  "category_ids": [1, 3]
}
```

Notes:
- `content` or `media_path` is required (one can be empty, not both).
- `privacy` must be `public`, `followers`, or `private`.
- `category_ids` is optional.

Response (201):

```json
{
  "success": true,
  "data": {
    "id": 10,
    "author_id": 2,
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

## React fetch example

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL;

export async function createPost(payload: {
  content?: string;
  media_path?: string;
  privacy: "public" | "followers" | "private";
  category_ids?: number[];
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
