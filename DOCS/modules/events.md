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

### List group events

`GET /groups/{id}/events?limit=20&offset=0`

Notes:
- Only group members can view group events.

### Get event by ID

`GET /events/{id}`

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
