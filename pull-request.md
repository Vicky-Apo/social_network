# Pull Request Summary

## Overview
This PR focuses on Docker reliability, pagination across feeds/posts/comments, and fixes a frontend bug that limited the dashboard feed to only the logged-in user’s posts.

## Key Changes

### 1. Docker & Build Workflow
- Added `DOCKER.md` with clean, repeatable setup steps for team onboarding.
- Refined root `Makefile` with clear sections for local vs Docker workflows and new service-specific Docker targets.
- Backend Docker image updated to Go 1.25.1 to match `go.mod` requirement.
- Frontend build stabilized in Docker by removing Google font downloads and using safe CSS fallbacks.
- `db-reset` now resets **only** Postgres to avoid unnecessary rebuilds.

### 2. Pagination (Posts + Comments)
**Backend**
- Added count methods to posts and comments repositories.
- API now returns `X-Total-Count` for:
  - `GET /posts`
  - `GET /posts?author_id=...`
  - `GET /groups/{id}/posts`
  - `GET /posts/{id}/comments`
- CORS exposes `X-Total-Count` for frontend access.

**Frontend**
- Added a reusable `Pagination` component.
- Dashboard, profile, and group posts now use page-number pagination (10 per page).
- Comments are now paginated (10 per page) in dashboard, profile, and group post UIs.
- Comment pagination state tracked per post (current page + total).

### 3. Feed Visibility Bug (Frontend)
- Fixed dashboard feed request that was incorrectly using `author_id=me`.
- Now uses global feed (`/posts`) by default, so public posts from other users appear as expected.

## Files Touched (Highlights)

### Docker / Build
- `DOCKER.md`
- `Makefile`
- `backend/Dockerfile`
- `frontend/src/app/layout.tsx`
- `frontend/src/app/globals.css`

### Pagination / API Counts
- `backend/internal/domain/post/repository.go`
- `backend/pkg/db/postgres/repositories/post/repository.go`
- `backend/internal/usecase/post/service.go`
- `backend/internal/transport/http/handler/post.go`
- `backend/internal/domain/comment/repository.go`
- `backend/pkg/db/postgres/repositories/comment/repository.go`
- `backend/internal/usecase/comment/service.go`
- `backend/internal/transport/http/handler/comment.go`
- `backend/internal/transport/http/middleware/cors.go`
- `frontend/src/components/Pagination.tsx`
- `frontend/src/app/dashboard/DashboardPage.tsx`
- `frontend/src/app/profile/ProfilePage.tsx`
- `frontend/src/app/groups/[id]/page.tsx`

### Tests Updated
- `backend/internal/transport/http/handler/post_test.go`
- `backend/internal/transport/http/handler/comment_test.go`
- `backend/internal/usecase/post/service_test.go`
- `backend/internal/usecase/comment/service_test.go`
- `backend/internal/usecase/reaction/service_test.go`
- `backend/internal/usecase/access/service_test.go`
- `backend/internal/transport/http/handler/reaction_test.go`

## Notes / Impact
- Pagination is now consistent (10 items/page) across posts and comments.
- Frontend feed visibility now matches backend rules.
- Docker builds no longer break on Google font fetches.

## Follow-Ups (Optional)
- Add pagination to events and members lists for parity.
- Consider adding tests for `X-Total-Count` headers in handlers.
