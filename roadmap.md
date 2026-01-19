# Wunderlist/Todoist Backend — Production Roadmap

> This roadmap is structured in phases (priority order). Each phase contains goals, recommended implementation tasks, and suggested tests.

---

## Phase 0 — Auth & Session Security ✅ (Done)

You already implemented a strong auth foundation:

* JWT access tokens with `jti=sessionID` (Mongo session `_id` hex)
* Redis-backed JWT blacklist (`jwt:blacklist:<jti>`) for **single-session logout**
* Redis-backed user-wide revoke (`jwt:user_revoked_at:<userID>`) for **Logout All Sessions** + **Admin Force Logout**
* Refresh token sessions stored in MongoDB with:
  * `token_hash` (bcrypt)
  * `token_sha` (sha256 hex) → indexed O(1) lookup
  * rotation (single-use refresh)
  * refresh reuse rejection
  * UA/IP binding
  * hijack detection → revoke all sessions
* Rate limiting (Redis in prod, memory limiter for tests)
* Integration tests + perf regression tests for logout/refresh (prevents scan regressions)

✅ **Status:** Auth is production-grade.

---

## Phase 1 — Backend E2E Completion (Identity + UX)

### 1. `GET /api/me` (current user profile)

**Goal:** Frontend can render logged-in profile quickly.

* Add handler: returns
  * `id`
  * `email`
  * `role`
  * `created_at`
* Never return password fields or internal auth details

**Tests**
* returns 200 with valid access token
* returns 401 without token

---

### 2. `DELETE /api/sessions/current` (logout current session)

**Goal:** Clean UX for “Logout” button.

* Delete refresh session using session id from JWT `jti`
* Revoke access token immediately using blacklist

**Tests**
* after logout-current, old token should fail immediately (401)
* refresh token should fail (401)

---

### 3. Password reset flow

**Goal:** Real-world readiness.

* Endpoints:
  * `POST /forgot-password`
  * `POST /reset-password`
* Reset token:
  * short TTL (10–30 min)
  * single-use
  * stored hashed (bcrypt/sha) in DB or Redis

**Tests**
* reset token expires
* reset token single-use
* password updated → old password fails

---

### 4. Email verification (optional but recommended)

**Goal:** Avoid fake accounts in production.

* Add email verification token + resend endpoint
* For Todoist-like apps, can be optional in MVP

**Tests**
* cannot login with unverified email (if enforced)

---

## Phase 2 — Core Product APIs (Todoist-like MVP)

### 5. Projects/Lists (existing) ✅ but finalize

**Goal:** Ensure stable, complete behavior.

* Validate title rules (length, trimming)
* Add consistent response shape
* Optional: soft delete / archive

**Tests**
* already good (ownership/RBAC) ✅

---

### 6. Tasks — scale-ready pagination (Cursor Pagination)

**Goal:** Stable pagination at scale (no skip/limit drift).

* Cursor-based pagination for tasks (recommended)
* Stable sorting: `created_at DESC, _id DESC`
* Response example:
  * `items: []`
  * `next_cursor: "<token>"`

**Tests**
* deterministic pagination
* invalid cursor → 400
* works even with concurrent inserts

---

### 7. Filters / Views (Todoist-like)

**Goal:** frontend views become easy.

Add query filters:
* `status=active|completed`
* `due=today|overdue|upcoming`
* `priority=low|medium|high`
* `list_id=<id>`

**Tests**
* each filter returns correct tasks

---

### 8. Labels/Tags (Todoist must-have)

**Goal:** Full Todoist feel.

* Add `labels` collection
* Task can have multiple labels
* Endpoints:
  * `POST /api/labels`
  * `GET /api/labels`
  * `PATCH /api/labels/:id`
  * `DELETE /api/labels/:id`
  * attach/detach labels to tasks

**Tests**
* ownership enforcement
* attach/detach works

---

### 9. Comments / Activity (nice-to-have, high value)

**Goal:** audit + collaborative feeling.

* Add task comments
* Store activity feed events for major actions

**Tests**
* comment create/list/delete

---

## Phase 3 — Collaboration (Optional / Todoist-like)

### 10. Sharing projects/lists

* Invite users
* Roles: `owner`, `editor`, `viewer`
* Permissions:
  * write tasks
  * comment
  * share/revoke access

**Tests**
* viewer cannot write
* editor can write
* owner can invite/remove

---

## Phase 4 — Production Hardening (Must-have)

### 11. Structured logging + request tracing

**Goal:** Faster debugging in production.

* Add `X-Request-Id` middleware
* Structured logs (Zap recommended)
* Log fields:
  * request_id
  * path
  * method
  * status
  * latency
  * user_id (if available)
  * ip, ua

**Tests**
* response contains `X-Request-Id`

---

### 12. Configuration hardening

**Goal:** Stop insecure defaults from reaching prod.

* If `APP_ENV=production`:
  * require JWT secrets (no fallback defaults)
  * validate Mongo URI is explicitly provided
* Safe logging: never log secrets

**Tests**
* config validation fails when missing secrets in production

---

### 13. Centralized error handling

**Goal:** Consistent API errors.

* Define standard error codes (example):
  * `AUTH_INVALID_TOKEN`
  * `AUTH_TOKEN_REVOKED`
  * `RATE_LIMITED`
  * `DB_ERROR`
* Add error middleware to map internal errors → HTTP responses

**Tests**
* ensure consistent JSON error format

---

### 14. Redis lifecycle handling

**Goal:** Clean shutdown and stronger readiness.

* Close Redis on shutdown
* Optional: reconnect strategies / timeouts

---

## Phase 5 — Observability & Reliability

### 15. Metrics

**Goal:** measure production health.

* Prometheus `/metrics`
* Counters:
  * login attempts
  * refresh attempts
  * revoked token usage
  * hijack events
* Latency histograms per route

---

### 16. Health/readiness upgrades

Improve `/ready`:
* Mongo ping
* Redis ping (only if redis configured)

---

## Phase 6 — Performance & Scaling

### 17. Caching (Redis)

**Goal:** reduce DB load.

Cache targets:
* lists by user
* active tasks
* task list per list_id

Cache keys examples:
* `cache:user:<uid>:lists`
* `cache:user:<uid>:tasks:<listID>`

Invalidation triggers:
* create/update/delete list/task

**Tests**
* cache hit/miss
* invalidation correctness

---

### 18. Async jobs

Use background jobs for:
* reminders
* digest emails
* cleanup tasks

Implementation options:
* Redis queue
* Kafka (since you already plan Kafka stack)

---

## Phase 7 — Testing Maturity

### 19. Test layers

* unit tests for services
* integration tests (Mongo + Redis) ✅ already strong
* perf regression tests ✅ done
* contract tests (OpenAPI schema)

---

### 20. CI pipeline

GitHub Actions:
* `go test ./...`
* `go test -race ./...`
* `golangci-lint run`
* integration tests using Mongo + Redis containers
* optional coverage enforcement

---

## Phase 8 — Deployment

### 21. Docker production build

* multi-stage Dockerfile
* non-root user
* small image (scratch/distroless)

---

### 22. Kubernetes production readiness

* resource requests/limits
* liveness/readiness probes
* HPA autoscaling
* ingress + TLS
* secrets management

