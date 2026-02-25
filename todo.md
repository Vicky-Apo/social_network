# Frontend TODO (Based on Current Implementation)

This list summarizes what is implemented in the frontend, which APIs it uses, and what still needs to be built. It follows the backend contract in `DOCS/modules/*` and `project_askings.md`.

## Implemented Pages / Features (Current)

- `frontend/src/app/page.tsx` (Landing): Marketing-only, no API usage.
- `frontend/src/app/login/page.tsx`:
  - Uses `POST /auth/login` with cookies.
- `frontend/src/app/register/page.tsx`:
  - Uses `POST /auth/register` with cookies.
- `frontend/src/app/dashboard/page.tsx`:
  - Uses `GET /auth/me`
  - Uses `GET /posts`
  - Uses `POST /posts`
  - Uses `GET /posts/{id}/comments`, `POST /posts/{id}/comments`
  - Uses `POST /posts/{id}/reactions`, `GET /posts/{id}/reactions`
  - Uses `POST /comments/{id}/reactions`, `GET /comments/{id}/reactions`
  - Uses `GET /notifications`, `GET /notifications/unread-count`
  - Uses `PATCH /notifications/{id}/read`, `PATCH /notifications/read-all`
  - Uses `GET /users` and `GET /users?q=...`
  - Uses WebSocket `/ws` (chat + notification events)
- `frontend/src/app/groups/page.tsx`:
  - Uses `GET /auth/me`
  - Uses `GET /groups`
- `frontend/src/app/groups/[id]/page.tsx`:
  - Uses `GET /auth/me`
  - Uses `GET /groups/{id}`
  - Uses `GET /groups/{id}/posts`
  - Uses `POST /groups/{id}/posts`
  - Uses `GET /posts/{id}/comments`, `POST /posts/{id}/comments`
  - Uses `POST /posts/{id}/reactions`, `GET /posts/{id}/reactions`
  - Uses `POST /comments/{id}/reactions`, `GET /comments/{id}/reactions`
- `frontend/src/app/messages/page.tsx`:
  - Uses `GET /auth/me`
  - Uses `GET /conversations`, `GET /conversations/{id}/messages`
  - Uses `PATCH /conversations/{id}/read`
  - Uses `GET /conversations/unread-counts`
  - Uses `POST /messages/{id}/reactions`, `GET /messages/{id}/reactions`
  - Uses WebSocket `/ws` (chat + typing + unread counts)

## Missing Pages / Features (Must Implement)

### Profile + Followers
- Pages:
  - Profile view page (public/private rules)
  - Profile edit page (nickname/about/avatar)
  - Followers list page
  - Following list page
- APIs to use:
  - `GET /profiles/{id}`
  - `GET /profiles/{id}/full` (profile + posts + activity)
  - `PATCH /profiles/{id}`
  - `PATCH /profiles/{id}/visibility`
  - `GET /profiles/{id}/followers`
  - `GET /profiles/{id}/following`

### Follow Requests + Unfollow
- Pages / UI:
  - Follow button on user cards / profile page
  - Incoming requests page
  - Sent requests page
  - Actions: accept/decline/cancel/unfollow/remove follower
- APIs to use:
  - `POST /follow-requests`
  - `GET /follow-requests`
  - `GET /follow-requests/sent`
  - `PATCH /follow-requests/{id}`
  - `DELETE /users/{id}/followers`
  - `DELETE /followers/{id}`

### Posts Privacy + Media Uploads
- UI:
  - Privacy selector (public/followers/private)
  - Allowed users selection for `private`
  - File upload for image/GIF in posts and comments
- APIs to use:
  - `POST /uploads` (get `media_path`)
  - `POST /posts` (with `privacy`, `allowed_user_ids`, `media_path`)
  - `POST /posts/{id}/comments` (with `media_path`)
  - `GET /uploads/{path...}` (for rendering secured media)

### Groups: Create + Membership + Invitations + Join Requests
- Pages / UI:
  - Create group page/form
  - Group join request flow
  - Group invitation flow
  - Group members list page
  - Leave group action
- APIs to use:
  - `POST /groups`
  - `GET /groups/{id}/members`
  - `POST /groups/{id}/invitations`
  - `GET /group-invitations`
  - `PATCH /group-invitations/{id}`
  - `POST /groups/{id}/join-requests`
  - `GET /groups/{id}/join-requests`
  - `PATCH /group-join-requests/{id}`
  - `DELETE /groups/{id}/members/me`

### Group Events
- Pages / UI:
  - Group events list
  - Create event form
  - Event detail + RSVP
- APIs to use:
  - `POST /groups/{id}/events`
  - `GET /groups/{id}/events`
  - `GET /events/{id}`
  - `PATCH /events/{id}`
  - `DELETE /events/{id}`
  - `POST /events/{id}/responses`
  - `GET /events/{id}/responses`

### Notifications (Global UI)
- Requirement: notifications on **every page**.
- Needed:
  - Shared notification tray / header component in groups/messages pages.
- APIs already used but not everywhere:
  - `GET /notifications`
  - `GET /notifications/unread-count`
  - `PATCH /notifications/{id}/read`
  - `PATCH /notifications/read-all`

### Chat: Group Chat & Media
- Missing UI:
  - Group chat room (if conversations include group chat)
  - Media upload for chat messages (if supported)
- APIs to use:
  - `GET /conversations` (already)
  - `GET /conversations/{id}/messages` (already)
  - WebSocket `/ws` (already)
  - `POST /uploads` + `media_path` for message attachments (if enabled)

## Known Contract Mismatches (Fix in Frontend Later)

- Login response in frontend expects `token` but backend is cookie-session based.
  - Update login flow to rely on cookie + `/auth/me` for session state.

