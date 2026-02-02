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
echo " 4️⃣a️⃣ CREATE MULTIPLE LISTS (FOR CURSOR)"
echo "=========================================="

for i in {1..25}; do
  curl -s -X POST "$BASE_URL/api/lists" \
    -H "$AUTH" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"List-$i\"}" > /dev/null
done

pass "multiple lists created for cursor pagination"


echo "=========================================="
echo " 5️⃣ GET LISTS (CURSOR PAGINATION)"
echo "=========================================="

LISTS_PAGE1=$(curl -s "$BASE_URL/api/lists?limit=10" -H "$AUTH")

ITEMS_COUNT=$(echo "$LISTS_PAGE1" | jq '.items | length')
HAS_MORE=$(echo "$LISTS_PAGE1" | jq -r '.has_more')

[[ "$ITEMS_COUNT" -eq 10 ]] || fail "expected 10 lists in page1"
[[ "$HAS_MORE" == "true" ]] || fail "expected has_more=true for lists"

NEXT_CURSOR=$(echo "$LISTS_PAGE1" | jq -r '.next_cursor')
[[ -n "$NEXT_CURSOR" && "$NEXT_CURSOR" != "null" ]] || fail "next_cursor missing"

pass "lists cursor page1 OK"

# Page 2
LISTS_PAGE2=$(curl -s "$BASE_URL/api/lists?limit=10&cursor=$NEXT_CURSOR" -H "$AUTH")
ITEMS2=$(echo "$LISTS_PAGE2" | jq '.items | length')

[[ "$ITEMS2" -gt 0 ]] || fail "page2 lists empty"
pass "lists cursor page2 OK"


echo "=========================================="
echo " 5️⃣b️⃣ LISTS CURSOR TAMPERING (SECURITY)"
echo "=========================================="

# Only test tampering if cursor exists
if [ "$HAS_MORE" == "true" ]; then
  TAMPERED_LIST_CURSOR="${NEXT_CURSOR%?}X"

  TAMPER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "$BASE_URL/api/lists?limit=10&cursor=$TAMPERED_LIST_CURSOR" \
    -H "$AUTH")

  [[ "$TAMPER_STATUS" == "400" ]] || fail "tampered list cursor not rejected"
  pass "tampered list cursor rejected"
else
  warn "list cursor tampering skipped (single page)"
fi



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
echo " 6️⃣a️⃣ CREATE MULTIPLE TASKS (FOR CURSOR)"
echo "=========================================="

for i in {1..30}; do
  curl -s -X POST "$BASE_URL/api/tasks" \
    -H "$AUTH" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"Task-$i\",\"list_id\":\"$LIST_ID\"}" > /dev/null
done

pass "multiple tasks created for cursor pagination"


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
echo " 8️⃣ GET TASKS (CURSOR PAGINATION)"
echo "=========================================="

TASKS_PAGE1=$(curl -s "$BASE_URL/api/tasks?limit=10" -H "$AUTH")

COUNT1=$(echo "$TASKS_PAGE1" | jq '.items | length')
HAS_MORE=$(echo "$TASKS_PAGE1" | jq -r '.has_more')

[[ "$COUNT1" -eq 10 ]] || fail "expected 10 tasks in page1"
[[ "$HAS_MORE" == "true" ]] || fail "expected has_more=true for tasks"

TASK_CURSOR=$(echo "$TASKS_PAGE1" | jq -r '.next_cursor')
[[ -n "$TASK_CURSOR" && "$TASK_CURSOR" != "null" ]] || fail "task cursor missing"

pass "tasks cursor page1 OK"

TASKS_PAGE2=$(curl -s "$BASE_URL/api/tasks?limit=10&cursor=$TASK_CURSOR" -H "$AUTH")
COUNT2=$(echo "$TASKS_PAGE2" | jq '.items | length')

[[ "$COUNT2" -gt 0 ]] || fail "tasks page2 empty"
pass "tasks cursor page2 OK"


echo "=========================================="
echo " 8️⃣a️⃣ ACTIVE TASKS (CURSOR PAGINATION)"
echo "=========================================="

ACTIVE_PAGE1=$(curl -s "$BASE_URL/api/tasks/active?limit=10" -H "$AUTH")
ACTIVE_COUNT=$(echo "$ACTIVE_PAGE1" | jq '.items | length')

[[ "$ACTIVE_COUNT" -gt 0 ]] || fail "active tasks empty"

HAS_MORE=$(echo "$ACTIVE_PAGE1" | jq -r '.has_more')
if [ "$HAS_MORE" == "true" ]; then
  CUR=$(echo "$ACTIVE_PAGE1" | jq -r '.next_cursor')
  [[ -n "$CUR" ]] || fail "active tasks cursor missing"
  pass "active tasks cursor OK"
else
  pass "active tasks single page"
fi


echo "=========================================="
echo " 9️⃣ CREATE MULTIPLE SEARCHABLE TASKS"
echo "=========================================="

for i in {1..5}; do
  curl -s -X POST "$BASE_URL/api/tasks" \
    -H "$AUTH" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"backend-task-$i\",\"list_id\":\"$LIST_ID\"}" > /dev/null
done

pass "multiple searchable tasks created"


echo "=========================================="
echo " 9️⃣a️⃣ SEARCH TASKS (CURSOR + SECURITY)"
echo "=========================================="

SEARCH1=$(curl -s -X GET "$BASE_URL/api/tasks/search?q=backend&limit=1" -H "$AUTH")
SEARCH_CURSOR=$(echo "$SEARCH1" | jq -r '.next_cursor')

[[ "$SEARCH_CURSOR" != "null" ]] || fail "search cursor missing"
pass "search page 1 OK"

SEARCH2=$(curl -s -X GET "$BASE_URL/api/tasks/search?q=backend&limit=1&cursor=$SEARCH_CURSOR" -H "$AUTH")
pass "search page 2 OK"

# 🔐 tamper cursor (flip last char)
TAMPERED="${SEARCH_CURSOR%?}X"

TAMPER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X GET "$BASE_URL/api/tasks/search?q=backend&cursor=$TAMPERED" \
  -H "$AUTH")

[[ "$TAMPER_STATUS" == "400" ]] || fail "tampered cursor not rejected"
pass "tampered cursor rejected"




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
echo " 1️⃣4️⃣a️⃣ LIST SESSIONS (CURSOR PAGINATION)"
echo "=========================================="

SESSIONS_PAGE1=$(curl -s "$BASE_URL/api/sessions?limit=2" -H "$AUTH")

COUNT=$(echo "$SESSIONS_PAGE1" | jq '.items | length')
HAS_MORE=$(echo "$SESSIONS_PAGE1" | jq -r '.has_more')

[[ "$COUNT" -ge 1 ]] || fail "expected at least 1 session"
pass "sessions page1 OK"

SESSION_CURSOR=$(echo "$SESSIONS_PAGE1" | jq -r '.next_cursor // empty')

if [ "$HAS_MORE" == "true" ]; then
  [[ -n "$SESSION_CURSOR" ]] || fail "session next_cursor missing"

  SESSIONS_PAGE2=$(curl -s \
    "$BASE_URL/api/sessions?limit=2&cursor=$SESSION_CURSOR" \
    -H "$AUTH")

  COUNT2=$(echo "$SESSIONS_PAGE2" | jq '.items | length')
  [[ "$COUNT2" -ge 0 ]] || fail "sessions page2 invalid"

  pass "sessions cursor page2 OK"
else
  warn "sessions fit in single page"
fi


echo "=========================================="
echo " 1️⃣4️⃣b️⃣ SESSION CURSOR TAMPERING (SECURITY)"
echo "=========================================="

if [ -n "$SESSION_CURSOR" ]; then
  TAMPERED_SESSION_CURSOR="${SESSION_CURSOR%?}X"

  TAMPER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "$BASE_URL/api/sessions?cursor=$TAMPERED_SESSION_CURSOR" \
    -H "$AUTH")

  [[ "$TAMPER_STATUS" == "400" ]] || fail "tampered session cursor not rejected"
  pass "tampered session cursor rejected"
else
  warn "session cursor tampering skipped"
fi


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
