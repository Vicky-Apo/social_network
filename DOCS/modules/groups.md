# Groups API (Frontend Guide)

This document describes the current backend surface area for groups based on the codebase scan.

## Status

There are **no HTTP REST endpoints for groups** implemented in the backend at this time. The HTTP router does not register any `/groups` routes.

## What exists

### WebSocket (group chat)

Group messages are supported **via WebSocket** as described in `DOCS/modules/chat.md`.

Send a group message by providing `group_id` in the WebSocket payload:

```json
{
  "type": "chat_message",
  "payload": {
    "group_id": 7,
    "content": "Hello group!"
  }
}
```

### Internal access checks (server-side only)

The backend contains access checks and data access for groups:

- `group` repository supports:
  - `GetByID`
  - `IsMember`
  - `GetMemberIDs`
- Access service includes:
  - `CanViewGroup`
  - `CanPostInGroup`
  - `CanChatInGroup`
  - `CanInviteToGroup`
  - `CanApproveGroupJoin`

These are not exposed via REST yet.

## Next steps (backend work required)

To support full group flows from the frontend, the backend needs HTTP endpoints for:

- Create group
- List groups / search groups
- Invite / accept / decline invitations
- Request to join / approve / decline
- Group posts and comments
- Group events (create + RSVP)

Once these endpoints exist, this document can be updated with concrete routes and examples.
