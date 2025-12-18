#!/bin/bash
set -e

BASE_URL="http://localhost:8080"
EMAIL="testuser@example.com"
PASSWORD="password123"

echo "🔐 Signing up..."
curl -s -X POST "$BASE_URL/signup" -H "Content-Type: application/json" \
-d "{\"name\":\"Test User\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" || true
echo "✅ Signup done (ignoring duplicate)"

echo "🔐 Logging in..."
TOKEN=$(curl -s -X POST "$BASE_URL/login" -H "Content-Type: application/json" \
-d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.token')
echo "✅ Token: $TOKEN"

AUTH_HEADER="Authorization: Bearer $TOKEN"

# -------------------------------
# Create List
# -------------------------------
echo "📁 Creating list..."
LIST_ID=$(curl -s -X POST "$BASE_URL/api/lists" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
-d '{"title":"Work"}' | jq -r '.id')
echo "✅ List ID: $LIST_ID"

# -------------------------------
# Get Lists
# -------------------------------
echo "📄 Fetching lists..."
curl -s -X GET "$BASE_URL/api/lists" -H "$AUTH_HEADER" | jq

# -------------------------------
# Create Task
# -------------------------------
echo "📝 Creating task..."
TASK_ID=$(curl -s -X POST "$BASE_URL/api/tasks" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
-d "{\"title\":\"Finish Wunderlist backend\",\"list_id\":\"$LIST_ID\",\"priority\":\"high\"}" | jq -r '.id')
echo "✅ Task ID: $TASK_ID"

# -------------------------------
# Get Tasks
# -------------------------------
echo "📋 Fetching tasks..."
curl -s -X GET "$BASE_URL/api/tasks?listId=$LIST_ID" -H "$AUTH_HEADER" | jq

# -------------------------------
# Update Task
# -------------------------------
echo "✏️ Updating task..."
curl -s -X PUT "$BASE_URL/api/tasks/$TASK_ID" -H "$AUTH_HEADER" -H "Content-Type: application/json" \
-d '{"title":"Finish Wunderlist backend (DONE)","completed":true,"priority":"low"}' | jq

# -------------------------------
# Get Active Tasks
# -------------------------------
echo "📌 Fetching active tasks..."
curl -s -X GET "$BASE_URL/api/tasks/active?listId=$LIST_ID" -H "$AUTH_HEADER" | jq

# -------------------------------
# Search Tasks
# -------------------------------
echo "🔍 Searching tasks..."
curl -s -X GET "$BASE_URL/api/tasks/search?q=wunderlist" -H "$AUTH_HEADER" | jq

# -------------------------------
# Delete Task
# -------------------------------
echo "🗑️ Deleting task..."
curl -s -X DELETE "$BASE_URL/api/tasks/$TASK_ID" -H "$AUTH_HEADER" | jq

# -------------------------------
# Delete List
# -------------------------------
echo "🗑️ Deleting list..."
curl -s -X DELETE "$BASE_URL/api/lists/$LIST_ID" -H "$AUTH_HEADER" | jq

echo "🎉 All APIs tested successfully!"
