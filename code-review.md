# Code Review: ENG-09-group-events

**Branch:** `ENG-09-group-events`
**Commits reviewed:** `4240025c` (add all other apis left to create), `a919d226` (add missing features and create testing)
**Scope:** ~13,200 lines added/modified across 99 files
**Date:** 2026-02-12

---

## Junk File

- **`backend/127.0.0.1:33299:`** — Empty file accidentally committed. Should be deleted and added to `.gitignore`.

---

## CRITICAL

### 1. Access check silently skipped in comment service

**File:** `backend/internal/usecase/comment/service.go:37-45, 78-86`

When `s.access` is `nil`, the authorization check is completely skipped instead of failing. Every other service in the codebase hard-fails with `"access service not configured"`. This means a wiring bug could leave comment creation/reading completely unprotected.

```go
// Line 37
if s.access != nil {
    ok, err := s.access.CanViewPost(ctx, req.AuthorID, req.PostID)
    ...
}
// If s.access == nil, falls through with NO access check.
```

**Recommendation:** Change both occurrences to fail hard when `s.access == nil`:

```go
if s.access == nil {
    return CommentDTO{}, errors.New("access service not configured")
}
```

### 2. Path traversal — missing post-resolution containment check

**File:** `backend/internal/transport/http/handler/media.go:39, 66`

The `strings.Contains(path, "..")` check is good but incomplete. After `filepath.Join(h.uploadDir, path)`, the result should be verified to still be within the upload directory. `filepath.Join` resolves `..` components, so the final path must be prefix-checked:

```go
fullPath := filepath.Join(h.uploadDir, filepath.FromSlash(path))
if !strings.HasPrefix(fullPath, filepath.Clean(h.uploadDir) + string(os.PathSeparator)) {
    // reject
}
```

### 3. CORS default: `AllowedOrigins: "*"` with `AllowCredentials: true`

**File:** `backend/internal/config/config.go:103-106`

```go
AllowedOrigins:   utils.GetString("CORS_ALLOWED_ORIGINS", "*"),
AllowCredentials: utils.GetBool("CORS_ALLOW_CREDENTIALS", true),
```

This combination is a security misconfiguration. If the CORS middleware reflects the `Origin` header (common behavior), this becomes a credential-stealing vulnerability. The default should either set `AllowCredentials: false` or restrict origins to a specific value (e.g., `http://localhost:3000`).

---

## HIGH

### 4. Raw `err.Error()` leaked to API clients

**File:** `backend/internal/transport/http/handler/post.go:149, 175`

`CreateInGroup` and `Create` send the raw error message to the client via `RespondWithError(w, http.StatusBadRequest, err.Error())`. This can leak internal details (DB errors, stack traces). Other handlers (event, group) properly map errors through dedicated `mapXxxError` functions.

### 5. Missing `<= 0` validation on parsed post ID

**File:** `backend/internal/transport/http/handler/comment.go:41, 78`

Unlike every other handler, the comment handler only checks `err != nil` but does not check if the parsed ID is `<= 0`. Zero or negative IDs pass validation.

### 6. `ParsePathID` accepts zero/negative IDs

**File:** `backend/internal/transport/http/utils/request.go:22-28`

All profile and some post handlers use `ParsePathID`, which does not validate `id > 0`. Routes like `GET /profiles/0` or `GET /profiles/-1` are accepted as valid.

### 7. Upload handler has no user identity tracking

**File:** `backend/internal/transport/http/handler/upload.go:35`

The handler never calls `GetUserID(r.Context())`. While the route is auth-gated by middleware, there is no user attribution on uploaded files — no per-user quotas, no audit trail.

---

## MEDIUM

### 8. Race conditions in group invite/join

**File:** `backend/internal/usecase/group/service.go:125-173, 219-258`

`InviteToGroup` and `RequestJoin` follow a check-then-act pattern (check membership, check existing invitation, then create) without a transaction. Concurrent requests can produce duplicate invitations or join requests.

**Recommendation:** Either wrap these operations in a database transaction, or ensure the repository layer has unique constraints that cause `CreateInvitation`/`CreateJoinRequest` to fail on duplicates.

### 9. Race condition in reaction toggle

**File:** `backend/internal/usecase/message_reaction/service.go:52-66`

`ToggleReaction` checks `HasMessageReaction` then either removes or adds. Two concurrent toggles can both see `exists=true` and both attempt removal, or both see `exists=false` and both try to add.

**Recommendation:** Use an upsert/atomic toggle at the repository level, or wrap in a transaction.

### 10. N+1 query in chat ListConversations

**File:** `backend/internal/usecase/chat/service.go:203-251`

For each conversation, up to 3 additional DB calls are made (members, group ID, last message). 50 conversations = 50-150 extra queries. Unread counts are already fetched in bulk, but the rest are not.

**Recommendation:** Add bulk/batch repository methods like `GetLastMessages(ctx, conversationIDs)` and `GetConversationMembersMap(ctx, conversationIDs)`.

### 11. Missing content validation in comment service

**File:** `backend/internal/usecase/comment/service.go:36-53`

No validation that content or media is provided. Empty comments can be created. The post service properly validates this.

**Recommendation:**

```go
if strings.TrimSpace(req.Content) == "" && strings.TrimSpace(req.MediaPath) == "" {
    return CommentDTO{}, errors.New("content or media is required")
}
```

### 12. No past-date validation for events

**File:** `backend/internal/usecase/event/service.go:53-55`

Only checks `EventTime.IsZero()`, not whether it's in the past. Users can create events scheduled for past dates.

**Recommendation:**

```go
if req.EventTime.Before(time.Now()) {
    return EventDTO{}, ErrInvalidEventTime
}
```

### 13. `activity_limit` has no upper bound

**File:** `backend/internal/transport/http/handler/profile.go:92`

Unlike `limit` which is capped at `MaxLimit (100)`, `activity_limit` can be set to any value, causing unbounded DB queries.

### 14. Arbitrary status strings accepted in group handler

**File:** `backend/internal/transport/http/handler/group.go:239, 347`

`UpdateInvitation` and `UpdateJoinRequest` only check `status != ""`. No whitelist of valid values (e.g., "accepted", "rejected"). The raw user input is echoed back in the JSON response.

### 15. Upload directory permissions too open

**File:** `backend/pkg/utils/fs.go:10`

```go
return os.MkdirAll(path, 0o755)
```

Permission `0755` means world-readable for directories holding private user content (private messages, private post media). Should be `0o750` or `0o700`.

### 16. Swallowed error in chat unique-constraint fallback

**File:** `backend/pkg/db/postgres/repositories/chat/repository.go:64-73`

On a 23505 unique constraint violation, the fallback SELECT error is silently discarded if it also fails, making the original INSERT error the only thing returned. The SELECT failure should be returned or logged.

### 17. `RowsAffected` errors silently ignored

**File:** `backend/pkg/db/postgres/repositories/group/repository.go:370, 500, 606`

If `RowsAffected()` returns an error (e.g., the driver does not support it), the error is silently ignored and execution continues as if the operation succeeded.

**Recommendation:**

```go
rows, err := res.RowsAffected()
if err != nil {
    return fmt.Errorf("rows affected: %w", err)
}
if rows == 0 {
    return domaingroup.ErrNotMember
}
```

---

## LOW

### 18. Inconsistent `ParsePathID` vs `r.PathValue` usage

`post.go:184`, `profile.go:34,69,159,193,229,275` use the deprecated manual `ParsePathID` approach while other handlers use Go 1.22's `r.PathValue("id")`. Should be unified.

### 19. Silent error swallowing in notifications

`backend/internal/usecase/event/service.go:202-227`, `backend/internal/usecase/group/service.go:240-257` — notification failures are completely silent with no logging. The event service has no logger dependency at all.

### 20. ILIKE wildcard injection

`backend/pkg/db/postgres/repositories/group/repository.go:156`, `backend/pkg/db/postgres/repositories/user/repository.go:194` — `%` and `_` in search input are not escaped, allowing pattern manipulation. Not a security vulnerability but causes unexpected search behavior.

**Recommendation:**

```go
escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(strings.TrimSpace(queryText))
pattern := "%" + escaped + "%"
```

### 21. `ProfileDTO.User` uses `any` type

**File:** `backend/internal/usecase/profile/dto.go:32`

Using `any` (`interface{}`) for the `User` field loses compile-time type safety. Consider using two separate DTO types (`FullProfileDTO` and `LimitedProfileDTO`) or a well-documented interface.

### 22. 403/404 logged as "bad request"

`backend/internal/transport/http/handler/event.go`, `group.go`, `message_reaction.go` — all non-5xx errors are funneled to `logBadRequest`, including 403 Forbidden and 404 Not Found. The `chat.go` handler correctly uses `logForbidden` and `logNotFound`.

### 23. Duplicate `nullableString` helper

Defined identically in both `backend/pkg/db/postgres/repositories/post/repository.go` and `backend/pkg/db/postgres/repositories/user/repository.go`. Should be extracted to a shared utility.

### 24. `response.go` uses `log.Printf` instead of structured logger

**File:** `backend/internal/transport/http/utils/response.go:24`

The rest of the application uses `logger.Logger`, but this file falls back to `log.Printf`, breaking log consistency.

### 25. No `Content-Disposition` header on media serving

**File:** `backend/internal/transport/http/handler/media.go:67`

`http.ServeFile` does not set `Content-Disposition`. For user-uploaded content, serving with `Content-Disposition: inline` or `attachment` and `X-Content-Type-Options: nosniff` is recommended to prevent MIME-type sniffing attacks.

### 26. N+1 in media `FindByPath`

**File:** `backend/pkg/db/postgres/repositories/media/repository.go:23-67`

Queries four separate tables sequentially (posts, comments, messages, users). Could be consolidated into a single `UNION ALL` query.

### 27. Extra round-trip in `UpdateProfile`

**File:** `backend/pkg/db/postgres/repositories/user/repository.go:58-101`

Does an `UPDATE` then calls `GetByID` separately. Could use `RETURNING` clause instead.

### 28. Fragile deferred rollback pattern

**File:** `backend/pkg/db/postgres/repositories/post/repository.go:128-132`

Uses a named-return-value-dependent rollback instead of the idiomatic `defer tx.Rollback()` (which is safe because rolling back a committed transaction is a no-op).

### 29. `ParseMultipartForm` maxMemory set too high

**File:** `backend/internal/transport/http/handler/upload.go:40`

`r.ParseMultipartForm(h.maxBytes)` uses the full file size limit as the in-memory buffer. Should be a smaller value like `10 << 20` (10 MB) regardless of max file size.

### 30. Group post privacy not forced

**File:** `backend/internal/usecase/post/service.go:91-111`

A user can create a group post with `privacy: "followers"`, which has ambiguous meaning in a group context. While currently safe (access control bypasses privacy for group posts), the stored value is misleading.

### 31. `RequestJoin` doesn't check pending invitations

**File:** `backend/internal/usecase/group/service.go:219-258`

If User A has been invited to a group but hasn't responded, they can also submit a join request, leading to inconsistent state.

### 32. `healthz` endpoint matches all HTTP methods

**File:** `backend/internal/transport/http/router.go:45`

Unlike all other routes which specify the HTTP method, the health check accepts all methods (POST, DELETE, etc.).

---

## INFORMATIONAL

### 33. No event deletion or update capability

The event service and repository have no `Delete` or `Update` methods. The event creator has no way to cancel or modify an event. May be intentional for current scope.

### 34. `GetGroup` has no access check

**File:** `backend/internal/usecase/group/service.go:85-91`

Any authenticated user can view any group's details. May be intentionally public for group discovery, but should be explicitly documented.

### 35. Redundant access checks in `post/service.go` ListByAuthor

**File:** `backend/internal/usecase/post/service.go:196-225`

The `!isOwner && !isFollower` conditions are redundant since `CanViewProfile` already returns `true` for followers/owners. Currently benign but fragile if access logic changes.

### 36. Repeated scan boilerplate in post repository

The post repository has essentially identical scan logic repeated in `List`, `ListByAuthor`, `ListByCategory`, and `ListByGroup`. Should be extracted into a `scanPost` helper.

---

## Summary

| Severity | Count | Key Areas |
|----------|-------|-----------|
| Critical | 3 | Auth bypass in comments, path traversal, CORS misconfiguration |
| High | 4 | Error leaking, missing validation, no upload attribution |
| Medium | 10 | Race conditions, N+1 queries, missing validation, error handling |
| Low | 15 | Code consistency, minor security hardening, DRY violations |
| Informational | 4 | Missing features, documentation gaps |

**Recommendation:** The critical and high issues should be fixed before merging. The medium issues are worth addressing but could be tracked as follow-up tasks. The low issues are improvements for code quality.
