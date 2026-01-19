# Wunderlist/Todoist Backend — Production Roadmap

> This roadmap is structured in phases (priority order). Each phase contains goals, recommended implementation tasks, and suggested tests.

---

## Phase 0 — Auth & Session Security ✅ (Done)

You already implemented a strong auth foundation:

* JWT access tokens with `jti`
* Redis-backed JWT blacklist (single-session logout)
* Redis-backed user-wide revoke (`jwt:user_revoked_at:<userID>`) for **Logout All Sessions** + **Admin Force Logout**
* Refresh token sessions stored in MongoDB with:

  * rotation (single-use refresh)
  * refresh reuse rejection
  * UA/IP binding
  * hijack detection → revoke all sessions
* Rate limiting (Redis in prod, memory limiter for tests)
* Integration tests + perf regression tests

---

## Phase 1 — Production Hardening (Must-have)

### 1. Structured logging + request tracing

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

* Response contains `X-Request-Id`

---

### 2. Configuration hardening

**Goal:** Stop insecure defaults from reaching prod.

* If `APP_ENV=production`:

  * require JWT secrets (no fallback defaults)
  * validate Mongo URI is explicitly provided
* Safe logging: never log secrets

**Tests**

* config validation fails when missing secrets in production

---

### 3. Centralized error handling

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

### 4. Redis lifecycle handling

**Goal:** Clean shutdown and stronger readiness.

* Close Redis on shutdown
* Optional: reconnect strategies / timeouts

---

## Phase 2 — Product Features (Todoist-like)

### 5. Better pagination

**Goal:** Stable pagination at scale.

* Cursor-based pagination for tasks (recommended)
* Stable sort: `created_at + _id`

**Tests**

* deterministic pagination order

---

### 6. Task reminders (core Todoist feature)

**Goal:** Notifications/workflow.

* Add fields to `Task`:

  * `remind_at` (UTC)
  * `timezone`
  * `reminder_sent`
* Add reminder worker:

  * Simple cron loop for MVP
  * Redis queue / Kafka for production

**Tests**

* reminder job marks reminder_sent

---

### 7. Search improvements

**Goal:** Better UX.
Options:

* Mongo text search (current) ✅
* Mongo Atlas Search ✅ recommended
* Meilisearch / Elasticsearch for advanced ranking

---

## Phase 3 — Security & Compliance

### 8. Permission model (RBAC/ABAC)

**Goal:** Strong authorization.

* Move from role-only to permissions:

  * `task:read`, `task:write`
  * `admin:user:list`, etc.

**Tests**

* permission checks for protected routes

---

### 9. Audit logging

**Goal:** Track critical events.
Store events in `audit_logs`:

* login success/failure
* refresh reuse attempt
* session hijack detected
* logout all
* admin force logout

**Tests**

* audit record created for force logout

---

## Phase 4 — Observability & Reliability

### 10. Metrics

**Goal:** measure production health.

* Prometheus `/metrics`
* Counters:

  * login attempts
  * refresh attempts
  * revoked token usage
  * hijack events
* Latency histograms per route

---

### 11. Health/readiness upgrades

Improve `/ready`:

* Mongo ping
* Redis ping (only if redis configured)

---

## Phase 5 — Performance & Scaling

### 12. Caching (Redis)

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

### 13. Async jobs

Use background jobs for:

* reminders
* digest emails
* cleanup tasks

Implementation options:

* Redis queue
* Kafka (since you already plan Kafka stack)

---

## Phase 6 — Testing Maturity

### 14. Test layers

* unit tests for services
* integration tests (Mongo + Redis) ✅ already strong
* perf regression tests ✅ done
* contract tests (OpenAPI schema)

---

### 15. CI pipeline

GitHub Actions:

* `go test ./...`
* `go test -race ./...`
* `golangci-lint run`
* integration tests using Mongo + Redis containers
* optional coverage enforcement

---

## Phase 7 — Deployment

### 16. Docker production build

* multi-stage Dockerfile
* non-root user
* small image (scratch/distroless)

---

### 17. Kubernetes production readiness

* resource requests/limits
* liveness/readiness probes
* HPA autoscaling
* ingress + TLS
* secrets management

---

## Suggested Next Milestone (Recommended)

✅ **Milestone: Observability Pack**

1. request_id middleware
2. structured logging (Zap)
3. prometheus metrics
4. readiness checks (Mongo + Redis)

This gives maximum production confidence with low risk.
