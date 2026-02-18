# Social Network Project Askings

This file summarizes the project requirements and constraints for the social network.

## Core Features
- Followers (follow requests, accept/decline, unfollow).
- Profiles with public/private visibility and user activity.
- Posts with privacy levels (public, followers, private).
- Comments on posts.
- Reactions on posts and comments (like/dislike).
- Groups (create, invite, join requests, events, group posts, group chat).
- Notifications (global visibility across pages, separate from messages).
- Chats (direct and group, real-time over WebSocket).

## Auth
- Registration fields: `email`, `password`, `first_name`, `last_name`, `date_of_birth`.
- Optional registration fields: `nickname`, `about`, `avatar`.
- Login via sessions and cookies (no bearer tokens).
- Users remain logged in until logout.

## Backend
- Database is Postgres (not SQLite).
- Migrations are required and must run on startup.
- WebSocket support for real-time chat and notifications.
- Image handling must support JPEG, PNG, GIF (store file path in DB).

## Frontend
- Use a JS framework (Next.js already selected).
- Client must send `credentials: "include"` for authenticated requests.
- Use `NEXT_PUBLIC_API_BASE_URL` for backend base URL.
- Match API usage described in `DOCS/modules/*.md`.

## Docker
- Two images: one for backend, one for frontend.
- Expose the necessary ports for frontend and backend communication.

## Status Notes
- Groups HTTP endpoints are not implemented yet (see `DOCS/modules/groups.md`).
- WebSocket chat is available at `/ws` (see `DOCS/modules/chat.md`).
