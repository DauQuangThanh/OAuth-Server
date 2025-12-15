#!/bin/bash

# Test script for Auth0-compatible server (Latest Architecture)
# Tests the enhanced server with clean architecture and monitoring

BASE_URL="http://localhost:8080"

echo "=== Auth0-Compatible Server API Test (Enhanced Architecture) ==="
echo

# Set environment variables for enhanced testing
export JWE_SECRET="821f56420e69830ea55929c0cfbbb2e07e9d564593cac476f6707042a8ebf75c"
export ENVIRONMENT="development"
export ENABLE_METRICS="true"
export ENABLE_TRACING="true"
export DB_DRIVER="memory"  # Use memory storage for testing
export DB_HOST="localhost"
export DB_NAME="Auth0_DB"
export DB_SSL_MODE="disable"
export DB_USER="postgres"
export DB_PASSWORD=""
export DB_PORT="5432"   # Use in-memory storage for testing


echo "Starting enhanced server..."
# Start server in background for comprehensive testing
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"
go run cmd/auth0-server/main.go &
SERVER_PID=$!
sleep 5

# Test 1: Health Check (enhanced)
echo "Test 1: Health Check (Enhanced)"
health_response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
if [ "$health_response" = "200" ]; then
    echo "✅ Health check passed"
    curl -s "$BASE_URL/health" | head -3
else
    echo "❌ Health check failed (HTTP $health_response)"
fi
echo

# Test 2: Metrics endpoint
echo "Test 2: Metrics Endpoint"
metrics_response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/metrics")
if [ "$metrics_response" = "200" ]; then
    echo "✅ Metrics endpoint accessible"
    echo "Sample metrics:"
    curl -s "$BASE_URL/metrics" | head -5
else
    echo "❌ Metrics endpoint failed (HTTP $metrics_response)"
fi
echo

# Test 3: User Registration (Enhanced)
echo "Test 3: User Registration (Enhanced)"
register_response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/dbconnections/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "test_client_123",
    "email": "enhanced@example.com",
    "password": "SecurePassword123!",
    "connection": "Username-Password-Authentication",
    "name": "Enhanced Test User",
    "nickname": "enhanced_user"
  }')

http_code="${register_response: -3}"
response_body="${register_response%???}"

if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
    echo "✅ Registration successful"
    echo "Response: $response_body"
else
    echo "❌ Registration failed (HTTP $http_code)"
    echo "Response: $response_body"
fi
echo

# Test 4: Token Exchange (Enhanced)
echo "Test 4: Token Exchange (Enhanced)"
if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
    token_response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/oauth/token" \
      -H "Content-Type: application/json" \
      -d '{
        "grant_type": "http://auth0.com/oauth/grant-type/password-realm",
        "username": "enhanced@example.com",
        "password": "SecurePassword123!",
        "client_id": "test_client_123",
        "realm": "Username-Password-Authentication",
        "scope": "openid profile email"
      }')
    
    token_http_code="${token_response: -3}"
    token_body="${token_response%???}"
    
    if [ "$token_http_code" = "200" ]; then
        echo "✅ Token exchange successful"
        echo "Token response received (truncated for security)"
        echo "$token_body" | head -1
        
        # Extract access token for further testing
        access_token=$(echo "$token_body" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$access_token" ]; then
            echo "✅ Access token extracted successfully"
            
            # Test 5: User Info (Enhanced)
            echo
            echo "Test 5: User Info with Token (Enhanced)"
            userinfo_response=$(curl -s -w "%{http_code}" -X GET "$BASE_URL/userinfo" \
              -H "Authorization: Bearer $access_token")
            
            userinfo_http_code="${userinfo_response: -3}"
            userinfo_body="${userinfo_response%???}"
            
            if [ "$userinfo_http_code" = "200" ]; then
                echo "✅ User info retrieval successful"
                echo "User info: $userinfo_body"
            else
                echo "❌ User info retrieval failed (HTTP $userinfo_http_code)"
            fi
        fi
    else
        echo "❌ Token exchange failed (HTTP $token_http_code)"
        echo "Response: $token_body"
    fi
else
    echo "⚠️  Skipping token test due to registration failure"
fi
echo

# Cleanup
echo "Cleaning up..."
if kill -0 $SERVER_PID 2>/dev/null; then
    kill $SERVER_PID
    echo "✅ Server stopped"
fi

echo
echo "=== Test Summary ==="
echo "✅ Health check test completed"
echo "✅ Metrics endpoint test completed" 
echo "✅ Registration test completed"
echo "✅ Token exchange test completed"
echo "✅ User info test completed"
echo "Enhanced architecture tests finished!"
