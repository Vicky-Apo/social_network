# ID Strategy Summary (Social Media App)

## Core Principle

Use **two IDs per entity**:

- **Internal ID** (`BIGSERIAL` / `BIGINT`) → database only
- **Public ID** (`UUID`) → API, frontend, URLs

Internal IDs exist for performance.  
Public IDs exist for security and safe exposure.

---

## Why This Approach

- `BIGSERIAL` is fast, index-efficient, and ideal for joins
- UUIDs prevent:
  - ID guessing
  - User enumeration
  - Data scraping
- Avoids performance penalties of UUID primary keys
- Matches proven industry architecture patterns

---

## Database Rules

- Keep `id BIGSERIAL PRIMARY KEY` on all main tables
- Add `public_id UUID NOT NULL DEFAULT gen_random_uuid()`
- Enforce `UNIQUE` constraint on `public_id`
- Do **not** add `public_id` to join tables
- Enable `pgcrypto` once for UUID generation

---

## API Rules

- Never expose numeric IDs
- All endpoints accept and return `public_id`
- URLs use UUIDs instead of integers
- Backend translates `public_id → id` internally

---

## Backend Rules (Go)

- Resolve `public_id` to internal `id` at request entry
- Use internal IDs for all joins and queries
- Authorization is based on relationships, not ID secrecy

---

## Frontend Rules (React)

- Store only `public_id`
- Never store or rely on numeric IDs
- Routing and links use UUIDs exclusively

---

## Security Reality

- IDs are not secrets
- Predictable IDs reduce attack cost
- Authorization must always be enforced independently

---

## Performance Reality

- BIGINT primary keys provide:
  - Smaller indexes
  - Faster joins
  - Better cache locality
- UUIDs only where public exposure is required

---

## Migration Strategy

1. Add `public_id` columns
2. Backfill via default UUID generation
3. Update backend to accept UUIDs
4. Update frontend routes
5. Deprecate numeric ID usage

---

## Final Verdict

This design is:

- Simple
- Fast
- Secure
- Scalable without premature complexity

It balances performance today with flexibility tomorrow.
