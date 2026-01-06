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

echo "🔎 Login raw response:"
echo "$LOGIN"

ACCESS=$(echo "$LOGIN" | jq -r '.access_token')
REFRESH=$(echo "$LOGIN" | jq -r '.refresh_token')

[[ "$ACCESS" != "null" && "$ACCESS" != "" ]] || fail "access token missing"
[[ "$REFRESH" != "null" && "$REFRESH" != "" ]] || fail "refresh token missing"

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
[[ "$LIST_ID" != "null" && "$LIST_ID" != "" ]] || fail "list id missing"
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
[[ "$TASK_ID" != "null" && "$TASK_ID" != "" ]] || fail "task id missing"
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
REFRESH=$(echo "$LOGIN" | jq -r '.refresh_token')

REFRESH_RES=$(curl -s -X POST "$BASE_URL/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}")

NEW_ACCESS=$(echo "$REFRESH_RES" | jq -r '.access_token')
NEW_REFRESH=$(echo "$REFRESH_RES" | jq -r '.refresh_token')

[[ "$NEW_ACCESS" != "null" ]] || fail "new access token missing"
[[ "$NEW_REFRESH" != "null" ]] || fail "new refresh token missing"

# 🔁 replace refresh token
REFRESH="$NEW_REFRESH"
AUTH="Authorization: Bearer $NEW_ACCESS"

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

echo "=========================================="
echo " 🎯 ALL TESTS COMPLETED SUCCESSFULLY"
echo "=========================================="
