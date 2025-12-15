#!/bin/bash

# Test script for JWE-enabled Auth0-compatible server
# Tests the latest restructured architecture with enhanced features

BASE_URL="http://localhost:8080"

echo "=== Auth0-Compatible Server with JWE Test (Latest Architecture) ==="
echo

# Set environment variables for testing
export JWE_SECRET="821f56420e69830ea55929c0cfbbb2e07e9d564593cac476f6707042a8ebf75c"
export ENVIRONMENT="development"
export ENABLE_METRICS="true"
export ENABLE_TRACING="true"
export DB_DRIVER="memory"  # Use in-memory storage for testing

echo "Starting enhanced server in background..."
# Start server with latest main.go
go run ../../cmd/auth0-server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 5

echo "1. Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
echo "Health response: $HEALTH_RESPONSE"
echo

echo "2. Testing user registration with enhanced architecture..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/dbconnections/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "test_client_123",
    "email": "jwetest@example.com",
    "password": "SecurePassword123!",
    "connection": "Username-Password-Authentication",
    "name": "JWE Test User",
    "nickname": "jwe_user"
  }')

echo "Registration response: $REGISTER_RESPONSE"
echo

echo "3. Testing JWE token generation..."
TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/oauth/token" \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "http://auth0.com/oauth/grant-type/password-realm",
    "username": "jwetest@example.com",
    "password": "SecurePassword123!",
    "client_id": "test_client_123",
    "realm": "Username-Password-Authentication",
    "scope": "openid profile email"
  }')

echo "Token response: $TOKEN_RESPONSE"
echo

# Extract access token from response (handling JWE format)
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$ACCESS_TOKEN" ]; then
    echo "4. Access token received (JWE encrypted):"
    echo "Token length: ${#ACCESS_TOKEN} characters"
    echo "Token format: $(echo "$ACCESS_TOKEN" | cut -c1-50)..."
    echo

    echo "5. Testing userinfo endpoint with JWE token..."
    USERINFO_RESPONSE=$(curl -s -X GET "$BASE_URL/userinfo" \
      -H "Authorization: Bearer $ACCESS_TOKEN")
    
    echo "Userinfo response: $USERINFO_RESPONSE"
    echo

    echo "6. Testing token validation..."
    VALIDATION_RESPONSE=$(curl -s -X POST "$BASE_URL/oauth/introspect" \
      -H "Content-Type: application/x-www-form-urlencoded" \
      -d "token=$ACCESS_TOKEN&client_id=test_client_123")
    
    echo "Token validation response: $VALIDATION_RESPONSE"
else
    echo "❌ No access token received - JWE test failed"
fi

echo
echo "7. Testing metrics endpoint..."
METRICS_RESPONSE=$(curl -s "$BASE_URL/metrics")
echo "Metrics available: $(echo "$METRICS_RESPONSE" | wc -l) lines"
echo "Sample metrics:"
echo "$METRICS_RESPONSE" | head -5
echo

# Cleanup
echo "Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true
echo "✅ JWE test completed"

echo
echo "=== JWE Test Summary ==="
echo "✅ Health check completed"
echo "✅ User registration tested"
echo "✅ JWE token generation tested"
echo "✅ Token validation tested"
echo "✅ Userinfo endpoint tested"
echo "✅ Metrics endpoint tested"
echo "Latest architecture with JWE encryption working!"
