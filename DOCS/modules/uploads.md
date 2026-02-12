# Uploads API (Frontend Guide)

This document explains how to upload media files and use them in posts, comments, profiles, and messages.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions

Uploads require a valid session cookie. Use `credentials: "include"` in the frontend.

## Endpoint

### Upload file

`POST /uploads`

Content type: `multipart/form-data`

Form fields:
- `file` (required) — the file to upload
- `kind` (optional) — one of: `media`, `avatar`, `post`, `comment`, `message`, `group`

Response (201):

```json
{
  "success": true,
  "data": {
    "path": "/uploads/post/20260212T120000_abcd1234ef567890.png",
    "content_type": "image/png"
  }
}
```

Notes:
- Only `image/jpeg`, `image/png`, and `image/gif` are accepted.
- Max upload size is controlled by `MAX_UPLOAD_MB`.
- Use the returned `path` as `media_path` or `avatar_path` in other endpoints.
