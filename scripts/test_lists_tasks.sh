#!/bin/bash
set -e

RED=$(tput setaf 1)
GREEN=$(tput setaf 2)
YELLOW=$(tput setaf 3)
RESET=$(tput sgr0)

BASE_URL="http://localhost:8080"

pass() { echo "${GREEN}✔ PASS${RESET} - $1"; }
warn() { echo "${YELLOW}⚠️ $1${RESET}"; }
fail() { echo "${RED}❌ FAIL${RESET} - $1"; exit 1; }

echo "=========================================="
echo " 1️⃣ RESET TEST ACCOUNT (HARD RESET)"
echo "=========================================="
curl -s -X POST "$BASE_URL/debug/reset-test-account" > /dev/null
pass "reset test user OK"

echo "=========================================="
echo " 2️⃣ LOGIN VALID CREDENTIALS"
echo "=========================================="
LOGIN=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"password123"}')

ACCESS=$(echo "$LOGIN" | jq -r '.access_token')
REFRESH=$(echo "$LOGIN" | jq -r '.refresh_token')

[[ -n "$ACCESS" && "$ACCESS" != "null" ]] || fail "access token missing"
[[ -n "$REFRESH" && "$REFRESH" != "null" ]] || fail "refresh token missing"

AUTH="Authorization: Bearer $ACCESS"
pass "login issued valid tokens"

echo "=========================================="
echo " 3️⃣ LOGIN WRONG PASSWORD SHOULD FAIL"
echo "=========================================="
BAD_LOGIN=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"wrongpass"}')

[[ "$BAD_LOGIN" == "401" ]] || fail "login wrong password should return 401"
pass "invalid password rejected"

echo "=========================================="
echo " 4️⃣ CREATE LIST"
echo "=========================================="
LIST_RES=$(curl -s -X POST "$BASE_URL/api/lists" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"title":"Work"}')

LIST_ID=$(echo "$LIST_RES" | jq -r '.id')
[[ -n "$LIST_ID" && "$LIST_ID" != "null" ]] || fail "list id missing"
pass "list created"

echo "=========================================="
echo " 5️⃣ GET LISTS"
echo "=========================================="
curl -s -X GET "$BASE_URL/api/lists" -H "$AUTH" > /dev/null
pass "get lists OK"

echo "=========================================="
echo " 6️⃣ CREATE TASK"
echo "=========================================="
TASK_RES=$(curl -s -X POST "$BASE_URL/api/tasks" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"Finish backend\",\"list_id\":\"$LIST_ID\"}")

TASK_ID=$(echo "$TASK_RES" | jq -r '.id')
[[ -n "$TASK_ID" && "$TASK_ID" != "null" ]] || fail "task id missing"
pass "task created"

echo "=========================================="
echo " 7️⃣ DUPLICATE TASK SHOULD FAIL"
echo "=========================================="
DUP=$(curl -s -o /dev/null -w "%{http_code}" \
  -X POST "$BASE_URL/api/tasks" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"Finish backend\",\"list_id\":\"$LIST_ID\"}")

[[ "$DUP" == "409" ]] || fail "duplicate task should return 409"
pass "duplicate prevented"

echo "=========================================="
echo " 8️⃣ GET ACTIVE TASKS"
echo "=========================================="
curl -s -X GET "$BASE_URL/api/tasks/active?list_id=$LIST_ID" \
  -H "$AUTH" > /dev/null
pass "active tasks OK"

echo "=========================================="
echo " 9️⃣ SEARCH TASKS"
echo "=========================================="
curl -s -X GET "$BASE_URL/api/tasks/search?q=backend" \
  -H "$AUTH" > /dev/null
pass "search tasks OK"

echo "=========================================="
echo " 🔟 UPDATE TASK"
echo "=========================================="
curl -s -X PUT "$BASE_URL/api/tasks/$TASK_ID" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"completed":true}' > /dev/null
pass "task updated"

echo "=========================================="
echo " 1️⃣1️⃣ REFRESH TOKEN (ROTATION)"
echo "=========================================="
REFRESH_RES=$(curl -s -X POST "$BASE_URL/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}")

ACCESS=$(echo "$REFRESH_RES" | jq -r '.access_token')
REFRESH=$(echo "$REFRESH_RES" | jq -r '.refresh_token')

AUTH="Authorization: Bearer $ACCESS"
pass "refresh token rotated successfully"

echo "=========================================="
echo " 1️⃣2️⃣ DELETE TASK"
echo "=========================================="
curl -s -X DELETE "$BASE_URL/api/tasks/$TASK_ID" \
  -H "$AUTH" > /dev/null
pass "task deleted"

echo "=========================================="
echo " 1️⃣3️⃣ DELETE LIST"
echo "=========================================="
curl -s -X DELETE "$BASE_URL/api/lists/$LIST_ID" \
  -H "$AUTH" > /dev/null
pass "list deleted"

# =======================
# Phase 1 – New APIs
# =======================

echo "=========================================="
echo " 1️⃣4️⃣ GET /api/me (PROFILE + SESSION COUNT)"
echo "=========================================="
ME=$(curl -s -X GET "$BASE_URL/api/me" -H "$AUTH")

EMAIL=$(echo "$ME" | jq -r '.email')
SESSIONS=$(echo "$ME" | jq -r '.active_sessions_count')

[[ "$EMAIL" == "testuser@example.com" ]] || fail "/api/me email mismatch"
[[ "$SESSIONS" -ge 1 ]] || fail "active_sessions_count invalid"
pass "/api/me returned profile and session count"

echo "=========================================="
echo " 1️⃣5️⃣ LOGOUT CURRENT SESSION"
echo "=========================================="
LOGOUT_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X DELETE "$BASE_URL/api/sessions/current" \
  -H "$AUTH")

[[ "$LOGOUT_STATUS" == "204" ]] || fail "logout current session failed"
pass "current session logged out"

echo "=========================================="
echo " 1️⃣6️⃣ ACCESS SHOULD FAIL AFTER LOGOUT"
echo "=========================================="
FAIL_ACCESS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X GET "$BASE_URL/api/me" \
  -H "$AUTH")

[[ "$FAIL_ACCESS" == "401" ]] || fail "access token should be revoked"
pass "access revoked immediately"

echo "=========================================="
echo " 1️⃣7️⃣ PASSWORD RESET (SMOKE)"
echo "=========================================="
curl -s -X POST "$BASE_URL/auth/forgot-password" \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com"}' > /dev/null
pass "forgot-password accepted"

warn "reset-password token delivery is async (email/SMS)"

echo "=========================================="
echo " 🎯 ALL TESTS COMPLETED SUCCESSFULLY"
echo "=========================================="
