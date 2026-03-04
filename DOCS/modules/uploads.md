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
- `kind` (optional) — one of: `media`, `avatar`, `post`, `comment`, `message`

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
- The `kind` value is informational; access control depends on where the path is stored (post/comment/message/avatar).
- `group` uploads are accepted but not currently served unless you add group media support.

Error responses:
- `400 Bad Request` - Invalid upload, invalid content type, or invalid kind
- `401 Unauthorized` - Not logged in or invalid session
- `500 Internal Server Error` - Failed to save file

### Serve uploaded file (secured)

`GET /uploads/{path...}`

This endpoint enforces per-entity privacy:
- Post media: requires `CanViewPost`
- Comment media: requires access to the parent post
- Message media: requires conversation membership
- Avatar: requires `CanViewProfile`

If access is denied, the API returns `404` to avoid leaking file existence.

Error responses:
- `400 Bad Request` - Invalid path
- `401 Unauthorized` - Not logged in or invalid session
- `404 Not Found` - File not found or access denied
- `500 Internal Server Error` - Failed to authorize or serve file
