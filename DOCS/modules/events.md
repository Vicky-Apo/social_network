# Events API (Frontend Guide)

This document explains how to create and view group events, and respond to them.

## Base URL

- `http://localhost:8080` (backend)
- `http://localhost:3000` (frontend)

## Cookie-based sessions

All event endpoints require a valid session cookie. Use `credentials: "include"` in the frontend.

## Endpoints

### Create event

`POST /groups/{id}/events`

Request body (JSON):

```json
{
  "title": "Weekly Standup",
  "description": "Team sync",
  "event_time": "2026-02-14T18:00:00Z"
}
```

Notes:
- Only group members can create events.
- Returns `404` if the group does not exist.

### List group events

`GET /groups/{id}/events?limit=20&offset=0`

Notes:
- Only group members can view group events.
- Returns `404` if the group does not exist.
- Each event includes `group_title` so the frontend can display the group name without a separate `/groups/{id}` call.

### Get event by ID

`GET /events/{id}`

Response includes:
- `group_title` (string)
- `responses_count` (number)

Notes:
- Use `group_title` for display to avoid a separate group lookup.
- Use `responses_count` to show a count and lazily fetch `/events/{id}/responses` only when needed.

### Update event

`PATCH /events/{id}`

Request body (JSON):

```json
{
  "title": "Updated Standup",
  "description": "Updated description",
  "event_time": "2026-02-15T18:00:00Z"
}
```

Notes:
- Only the event creator can update the event.
- `event_time` must be in the future.
- This endpoint expects a full update (provide `title`, `description`, and `event_time`).

### Delete event

`DELETE /events/{id}`

Notes:
- Only the event creator can delete the event.

### Respond to event

`POST /events/{id}/responses`

Request body (JSON):

```json
{
  "response": "going"
}
```

Allowed values: `going`, `not_going`.

### List event responses

`GET /events/{id}/responses`

Notes:
- Only group members can view responses.
