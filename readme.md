# Run all unit + integration tests
go test ./tests/...

# Run only integration tests
go test ./tests/integration/...




PORT=8080
MONGO_URI=mongodb://localhost:27017
MONGO_DB=wunderlist
MONGO_TIMEOUT_SEC=10
JWT_SECRET=supersecret
JWT_REFRESH_SECRET=superrefreshsecret
JWT_ACCESS_TTL_MIN=15
JWT_REFRESH_TTL_HOURS=168
APP_ENV=development




🔥 High-Priority (Must Fix Before Production)
1. Security & Authentication

 Implement refresh tokens to renew JWTs safely.

 Add iss, aud, nbf claims in JWT for stronger validation.

 Ensure strong password hashing (bcrypt/argon2).

 Add user ownership checks for every endpoint to prevent cross-user access.

 Sanitize all $text queries to prevent injection.

2. Database & Performance

 Ensure all MongoDB queries use proper indexes:

Tasks: user_id + list_id + completed

Active tasks: partial index on completed: false

Search: text index on title

 Add maximum pagination limits (e.g., limit <= 100).

 Context timeouts on all DB operations (already partially implemented).

3. Input Validation

 Validate all request payloads using struct validation tags (validator package).

 Check required fields (title, list_id, email) and enforce length limits.

⚡ Medium-Priority (Should Fix Before Production for Reliability)
1. Logging & Observability

 Replace log.Printf with structured logging (Zap/Zerolog/Logrus).

 Include request IDs in logs for traceability.

 Log internal error stacks, but don’t expose to clients.

 Add metrics (Prometheus/Grafana) for request rates, response times, errors.

2. Middleware & HTTP Hardening

 Add CORS middleware if frontend is separate.

 Add security headers: X-Content-Type-Options, X-Frame-Options, Strict-Transport-Security.

 Add global error-handling middleware to standardize responses.

3. Swagger / API Documentation

 Document all public endpoints (/signup, /login, /google-login).

 Include example request & response bodies for clarity.

 Ensure query params and path variables are accurately documented.

🛠 Low-Priority (Nice to Have / Polish Features)
1. Graceful Shutdown

 Use http.Server with context for OS signal handling.

 Close MongoDB connections gracefully on shutdown.

 Ensure background goroutines respect context cancellation.

2. Testing & CI/CD

 Unit tests for all handlers and middleware.

 Integration tests with a MongoDB test instance.

 Add CI/CD pipeline to run tests, linting, and code coverage checks.

3. Configuration & Environment

 Validate critical env variables on startup: JWTSecret, MongoDBName, Port.

 Use secrets management for sensitive data.

 Support multiple environments (dev/staging/prod) via separate configs.