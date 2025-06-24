#!/bin/bash

# Quick Notification System Test
echo "🚀 Quick notification system test..."

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
EMAIL="${EMAIL:-andrew.sameh@dockmaster.com}"
PASSWORD="${PASSWORD:-Password@123}"

echo "📡 Base URL: $BASE_URL"
echo "👤 Email: $EMAIL"

# Step 1: Get authentication token
echo ""
echo "🔐 Getting authentication token..."
AUTH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$EMAIL\",
        \"password\": \"$PASSWORD\"
    }")

if [ $? -ne 0 ]; then
    echo "❌ Failed to connect to server"
    exit 1
fi

TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.data.accessToken // empty')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "❌ Authentication failed"
    echo "Response: $AUTH_RESPONSE"
    exit 1
fi

echo "✅ Authentication successful"
echo "🔑 Token: ${TOKEN:0:20}..."

# Step 2: Test unread count
echo ""
echo "📊 Testing unread count endpoint..."
UNREAD_RESPONSE=$(curl -s -X GET "$BASE_URL/notifications/unread-count" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

UNREAD_STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/notifications/unread-count" \
    -H "Authorization: Bearer $TOKEN")

echo "Status Code: $UNREAD_STATUS_CODE"
echo "Response: $UNREAD_RESPONSE"

if [ "$UNREAD_STATUS_CODE" = "200" ]; then
    UNREAD_COUNT=$(echo "$UNREAD_RESPONSE" | jq -r '.data.count // 0')
    echo "✅ Unread count: $UNREAD_COUNT"
else
    echo "❌ Unread count endpoint failed"
fi

# Step 3: Test list notifications
echo ""
echo "📋 Testing list notifications endpoint..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/notifications/list?pageSize=10" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

LIST_STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/notifications/list?pageSize=10" \
    -H "Authorization: Bearer $TOKEN")

echo "Status Code: $LIST_STATUS_CODE"
echo "Response: $LIST_RESPONSE"

if [ "$LIST_STATUS_CODE" = "200" ]; then
    NOTIFICATION_COUNT=$(echo "$LIST_RESPONSE" | jq -r '.data | length')
    echo "✅ Total notifications: $NOTIFICATION_COUNT"
else
    echo "❌ List notifications endpoint failed"
fi

# Step 4: Test notifications by type
echo ""
echo "🔍 Testing notifications by type endpoint..."
TYPE_RESPONSE=$(curl -s -X GET "$BASE_URL/notifications/type/message?pageSize=10" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

TYPE_STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/notifications/type/message?pageSize=10" \
    -H "Authorization: Bearer $TOKEN")

echo "Status Code: $TYPE_STATUS_CODE"
echo "Response: $TYPE_RESPONSE"

if [ "$TYPE_STATUS_CODE" = "200" ]; then
    MESSAGE_COUNT=$(echo "$TYPE_RESPONSE" | jq -r '.data | length')
    echo "✅ Message notifications: $MESSAGE_COUNT"
else
    echo "❌ Notifications by type endpoint failed"
fi

# Step 5: Test SSE endpoint (just check if it connects)
echo ""
echo "🌊 Testing SSE endpoint connection..."
SSE_STATUS_CODE=$(timeout 5s curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/notifications/stream" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Accept: text/event-stream" \
    -H "Cache-Control: no-cache")

echo "SSE Status Code: $SSE_STATUS_CODE"

if [ "$SSE_STATUS_CODE" = "200" ]; then
    echo "✅ SSE endpoint accessible"
else
    echo "❌ SSE endpoint failed or timeout"
fi

echo ""
echo "🎉 Quick test completed!"
echo ""
echo "💡 To run the full test suite:"
echo "   export TEST_AUTH_TOKEN=\"$TOKEN\""
echo "   go run main.go"
