#!/bin/bash

# Notification System Test Setup Script
echo "🔧 Setting up notification system test environment..."

# Default values
BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
EMAIL="${EMAIL:-andrew.sameh@dockmaster.com}"
PASSWORD="${PASSWORD:-Password@123}"

echo "📡 Base URL: $BASE_URL"
echo "👤 Email: $EMAIL"

# Function to get auth token
get_auth_token() {
    echo "🔐 Getting authentication token..."

    AUTH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$EMAIL\",
            \"password\": \"$PASSWORD\"
        }")

    if [ $? -eq 0 ]; then
        TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.data.accessToken // empty')
        USER_ID=$(echo "$AUTH_RESPONSE" | jq -r '.data.user.id // empty')
        ORG_ID=$(echo "$AUTH_RESPONSE" | jq -r '.data.user.organizationId // empty')
        MARINA_ID=$(echo "$AUTH_RESPONSE" | jq -r '.data.user.marinaId // empty')

        if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
            echo "✅ Authentication successful!"
            echo "🔑 Token: ${TOKEN:0:20}..."
            echo "👤 User ID: $USER_ID"
            echo "🏢 Organization ID: $ORG_ID"
            echo "⚓ Marina ID: $MARINA_ID"

            # Export environment variables
            export TEST_AUTH_TOKEN="$TOKEN"
            export TEST_USER_ID="$USER_ID"
            export TEST_ORG_ID="$ORG_ID"
            export TEST_MARINA_ID="$MARINA_ID"
            export TEST_BASE_URL="$BASE_URL"

            # Try to get a customer ID for testing
            echo "🔍 Looking for test customer..."
            CUSTOMERS_RESPONSE=$(curl -s -X GET "$BASE_URL/customers/list?pageSize=1" \
                -H "Authorization: Bearer $TOKEN" \
                -H "Content-Type: application/json")

            CUSTOMER_ID=$(echo "$CUSTOMERS_RESPONSE" | jq -r '.data[0].customerID // empty')
            if [ -n "$CUSTOMER_ID" ] && [ "$CUSTOMER_ID" != "null" ]; then
                echo "👥 Found customer ID: $CUSTOMER_ID"
                export TEST_CUSTOMER_ID="$CUSTOMER_ID"
            else
                echo "⚠️  No customers found - message tests will be skipped"
            fi

            return 0
        else
            echo "❌ Authentication failed!"
            echo "Response: $AUTH_RESPONSE"
            return 1
        fi
    else
        echo "❌ Failed to connect to server!"
        return 1
    fi
}

# Function to check server status
check_server() {
    echo "🏥 Checking server health..."

    HEALTH_RESPONSE=$(curl -s "$BASE_URL/health" 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo "✅ Server is running!"
        return 0
    else
        echo "❌ Server is not responding at $BASE_URL"
        echo "💡 Make sure the server is running with: go run cmd/api/main.go"
        return 1
    fi
}

# Function to check dependencies
check_dependencies() {
    echo "📦 Checking dependencies..."

    if ! command -v curl &>/dev/null; then
        echo "❌ curl is not installed"
        return 1
    fi

    if ! command -v jq &>/dev/null; then
        echo "❌ jq is not installed (required for JSON parsing)"
        echo "💡 Install with: brew install jq (macOS) or apt-get install jq (Linux)"
        return 1
    fi

    echo "✅ Dependencies are available"
    return 0
}

# Function to display environment variables
show_env() {
    echo ""
    echo "🌍 Environment variables for testing:"
    echo "export TEST_AUTH_TOKEN=\"$TEST_AUTH_TOKEN\""
    echo "export TEST_USER_ID=\"$TEST_USER_ID\""
    echo "export TEST_ORG_ID=\"$TEST_ORG_ID\""
    echo "export TEST_MARINA_ID=\"$TEST_MARINA_ID\""
    echo "export TEST_CUSTOMER_ID=\"$TEST_CUSTOMER_ID\""
    echo "export TEST_BASE_URL=\"$TEST_BASE_URL\""
    echo ""
    echo "💡 You can copy these exports to your shell or run:"
    echo "   source <(./setup.sh)"
}

# Function to run the test
run_test() {
    echo "🚀 Running notification system tests..."
    cd "$(dirname "$0")"
    go run main.go
}

# Main execution
main() {
    if ! check_dependencies; then
        exit 1
    fi

    if ! check_server; then
        exit 1
    fi

    if ! get_auth_token; then
        exit 1
    fi

    show_env

    echo ""
    read -p "🤔 Do you want to run the tests now? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        run_test
    else
        echo "💡 To run tests later, use: go run main.go"
        echo "💡 Or run with custom parameters:"
        echo "   TEST_AUTH_TOKEN=\"$TEST_AUTH_TOKEN\" go run main.go"
    fi
}

# Handle sourcing vs execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # Script is being executed
    main "$@"
else
    # Script is being sourced
    if check_dependencies && check_server && get_auth_token; then
        echo "✅ Environment variables exported!"
    fi
fi
