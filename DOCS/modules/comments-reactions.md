# Comments & Reactions API (Frontend Guide)

This document explains how the comments and reactions endpoints work and how to use them from a React frontend.

## Base URL

Set a base URL for the backend in the frontend environment, for example:

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Authentication

All endpoints require a valid session cookie (same as auth endpoints). Make sure to include `credentials: "include"` in your requests.

## Comments Endpoints

### Create Comment

`POST /posts/{id}/comments`

Creates a new comment on a post.

**URL Parameters:**
- `id` - The post ID (integer)

**Request body (JSON):**

```json
{
  "author_id": 1,
  "content": "This is a great post!",
  "media_path": "/uploads/reply.gif"
}
```

**Notes:**
- `author_id` is the user ID of the comment author
- `content` is required and cannot be empty
- `media_path` is optional (string path to image/GIF)
- The `post_id` is automatically extracted from the URL path

**Response (201):**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "post_id": 1,
    "author_id": 1,
    "content": "This is a great post!",
    "media_path": null,
    "like_count": 0,
    "dislike_count": 0,
    "created_at": "2025-01-24T12:34:56Z",
    "updated_at": "2025-01-24T12:34:56Z"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid post ID or request body
- `500 Internal Server Error` - Failed to create comment (e.g., post doesn't exist)

### Get Post Comments

`GET /posts/{id}/comments?limit=20&offset=0`

Retrieves all comments for a specific post.

**URL Parameters:**
- `id` - The post ID (integer)

**Query Parameters:**
- `limit` (optional, default 20, max 100)
- `offset` (optional, default 0)

**Response (200):**

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "post_id": 1,
      "author_id": 1,
      "content": "This is a great post!",
      "media_path": null,
      "like_count": 1,
      "dislike_count": 0,
      "created_at": "2025-01-24T12:34:56Z",
      "updated_at": "2025-01-24T12:34:56Z"
    },
    {
      "id": 2,
      "post_id": 1,
      "author_id": 2,
      "content": "I agree!",
      "media_path": null,
      "like_count": 0,
      "dislike_count": 1,
      "created_at": "2025-01-24T12:35:10Z",
      "updated_at": "2025-01-24T12:35:10Z"
    }
  ]
}
```

**Notes:**
- Returns an empty array `[]` if the post has no comments
- Returns `404 Not Found` if the post doesn't exist

## Reactions Endpoints

### Toggle Post Reaction

`POST /posts/{id}/reactions`

Adds, updates, or removes a reaction to a post.

**URL Parameters:**
- `id` - The post ID (integer)

**Request body (JSON):**

```json
{
  "user_id": 1,
  "reaction": "like"
}
```

**Valid reaction types:**
- `"like"` - User likes the post
- `"dislike"` - User dislikes the post

**Response (200):**

```json
{
  "success": true,
  "data": {
    "status": "added"
  }
}
```

**Notes:**
- If the same reaction is sent again, it is removed (`status: "removed"`).
- If the opposite reaction is sent, it is updated (`status: "updated"`).

**Error Responses:**
- `400 Bad Request` - Invalid post ID, invalid reaction type, or invalid request body

### Get Post Reactions

`GET /posts/{id}/reactions`

Retrieves all reactions for a specific post.

**URL Parameters:**
- `id` - The post ID (integer)

**Response (200):**

```json
{
  "success": true,
  "data": [
    {
      "user_id": 1,
      "reaction": "like",
      "created_at": "2025-01-24T12:34:56Z"
    },
    {
      "user_id": 2,
      "reaction": "dislike",
      "created_at": "2025-01-24T12:35:10Z"
    }
  ]
}
```

**Notes:**
- Returns an empty array `[]` if the post has no reactions
- Each user can only have one reaction per post (upsert behavior)

### Toggle Comment Reaction

`POST /comments/{id}/reactions`

Adds or updates a reaction to a comment. If the user already has a reaction, it will be updated.

**URL Parameters:**
- `id` - The comment ID (integer)

**Request body (JSON):**

```json
{
  "user_id": 1,
  "reaction": "like"
}
```

**Valid reaction types:**
- `"like"` - User likes the comment
- `"dislike"` - User dislikes the comment

**Response (200):**

```json
{
  "success": true,
  "data": {
    "status": "added"
  }
}
```

**Notes:**
- If the same reaction is sent again, it is removed (`status: "removed"`).
- If the opposite reaction is sent, it is updated (`status: "updated"`).

**Error Responses:**
- `400 Bad Request` - Invalid comment ID, invalid reaction type, or invalid request body

### Get Comment Reactions

`GET /comments/{id}/reactions`

Retrieves all reactions for a specific comment.

**URL Parameters:**
- `id` - The comment ID (integer)

**Response (200):**

```json
{
  "success": true,
  "data": [
    {
      "user_id": 1,
      "reaction": "like",
      "created_at": "2025-01-24T12:34:56Z"
    },
    {
      "user_id": 2,
      "reaction": "like",
      "created_at": "2025-01-24T12:35:10Z"
    }
  ]
}
```

**Notes:**
- Returns an empty array `[]` if the comment has no reactions
- Each user can only have one reaction per comment (upsert behavior)

## React Fetch Examples

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL;

// Create a comment on a post
export async function createComment(postId: number, authorId: number, content: string, mediaPath?: string) {
  const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({
      author_id: authorId,
      content: content,
      media_path: mediaPath,
    }),
  });

  if (!res.ok) {
    throw new Error("Failed to create comment");
  }

  return res.json();
}

// Get all comments for a post
export async function getPostComments(postId: number) {
  const res = await fetch(`${API_BASE}/posts/${postId}/comments`, {
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("Failed to get comments");
  }

  return res.json();
}

// Add a reaction to a post
export async function addPostReaction(postId: number, userId: number, reaction: "like" | "dislike") {
  const res = await fetch(`${API_BASE}/posts/${postId}/reactions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({
      user_id: userId,
      reaction: reaction,
    }),
  });

  if (!res.ok) {
    throw new Error("Failed to add reaction");
  }

  return res.json();
}

// Get all reactions for a post
export async function getPostReactions(postId: number) {
  const res = await fetch(`${API_BASE}/posts/${postId}/reactions`, {
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("Failed to get reactions");
  }

  return res.json();
}

// Add a reaction to a comment
export async function addCommentReaction(commentId: number, userId: number, reaction: "like" | "dislike") {
  const res = await fetch(`${API_BASE}/comments/${commentId}/reactions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({
      user_id: userId,
      reaction: reaction,
    }),
  });

  if (!res.ok) {
    throw new Error("Failed to add reaction");
  }

  return res.json();
}

// Get all reactions for a comment
export async function getCommentReactions(commentId: number) {
  const res = await fetch(`${API_BASE}/comments/${commentId}/reactions`, {
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("Failed to get reactions");
  }

  return res.json();
}
```

## Axios Examples

```ts
import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  withCredentials: true,
});

// Create a comment on a post
export async function createComment(postId: number, authorId: number, content: string, mediaPath?: string) {
  const { data } = await api.post(`/posts/${postId}/comments`, {
    author_id: authorId,
    content: content,
    media_path: mediaPath,
  });
  return data;
}

// Get all comments for a post
export async function getPostComments(postId: number) {
  const { data } = await api.get(`/posts/${postId}/comments`);
  return data;
}

// Add a reaction to a post
export async function addPostReaction(postId: number, userId: number, reaction: "like" | "dislike") {
  const { data } = await api.post(`/posts/${postId}/reactions`, {
    user_id: userId,
    reaction: reaction,
  });
  return data;
}

// Get all reactions for a post
export async function getPostReactions(postId: number) {
  const { data } = await api.get(`/posts/${postId}/reactions`);
  return data;
}

// Add a reaction to a comment
export async function addCommentReaction(commentId: number, userId: number, reaction: "like" | "dislike") {
  const { data } = await api.post(`/comments/${commentId}/reactions`, {
    user_id: userId,
    reaction: reaction,
  });
  return data;
}

// Get all reactions for a comment
export async function getCommentReactions(commentId: number) {
  const { data } = await api.get(`/comments/${commentId}/reactions`);
  return data;
}
```

## Important Notes

1. **Upsert Behavior**: When adding a reaction, if the user already has a reaction on that post/comment, it will be updated (not duplicated). This means each user can only have one reaction per post/comment.

2. **Foreign Key Constraints**: Comments and reactions require valid post/user IDs. If you try to create a comment on a non-existent post, you'll get a 500 error.

3. **Error Handling**: Always check `res.ok` or handle axios errors. The API returns standard HTTP status codes:
   - `200` - Success (GET requests)
   - `201` - Created (POST requests)
   - `400` - Bad Request (invalid input)
   - `404` - Not Found (resource doesn't exist)
   - `500` - Internal Server Error (database constraint violations, etc.)

4. **Session Cookies**: All endpoints require authentication via session cookies. Make sure your frontend includes `credentials: "include"` (fetch) or `withCredentials: true` (axios).
