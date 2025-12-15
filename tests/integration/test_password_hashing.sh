#!/bin/bash

echo "=== Password Hashing Verification Test ==="
echo

# Clean up any existing test users
echo "1. Cleaning up existing test users..."
PGPASSWORD="" psql -h localhost -U postgres -d Auth0_DB -c "DELETE FROM users WHERE email LIKE '%test%';"
echo

# Start server with PostgreSQL
echo "2. Starting server with PostgreSQL..."
DB_DRIVER=postgres go run ../../cmd/auth0-server/main.go &
SERVER_PID=$!
sleep 3

echo "3. Testing user signup with password hashing..."
SIGNUP_RESPONSE=$(curl -s -X POST http://localhost:8080/dbconnections/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"hash_test@example.com","password":"plaintextpassword123"}')

echo "Signup response: $SIGNUP_RESPONSE"
echo

echo "4. Checking password in database (should be hashed)..."
PGPASSWORD="" psql -h localhost -U postgres -d Auth0_DB -c "
SELECT 
  id, 
  email, 
  substring(password, 1, 20) as password_start,
  length(password) as password_length,
  (password LIKE '\$2a\$%' OR password LIKE '\$2b\$%' OR password LIKE '\$2y\$%') as is_bcrypt
FROM users 
WHERE email = 'hash_test@example.com';"
echo

echo "5. Testing login with the same password..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "http://auth0.com/oauth/grant-type/password-realm",
    "username": "hash_test@example.com",
    "password": "plaintextpassword123",
    "client_id": "test_client",
    "realm": "Username-Password-Authentication"
  }')

echo "Login response: $LOGIN_RESPONSE"
echo

echo "6. Testing login with wrong password (should fail)..."
WRONG_LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "http://auth0.com/oauth/grant-type/password-realm",
    "username": "hash_test@example.com",
    "password": "wrongpassword",
    "client_id": "test_client",
    "realm": "Username-Password-Authentication"
  }')

echo "Wrong password response: $WRONG_LOGIN_RESPONSE"
echo

# Cleanup
echo "7. Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true
echo

echo "=== Password Hashing Test Completed ==="
echo "✅ If bcrypt hashes are shown in database, password security is working correctly"
