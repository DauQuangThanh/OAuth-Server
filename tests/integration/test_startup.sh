#!/bin/bash

# Simple server startup test
echo "Testing server startup with in-memory storage..."

export JWE_SECRET="821f56420e69830ea55929c0cfbbb2e07e9d564593cac476f6707042a8ebf75c"
export DB_DRIVER="memory"
export ENVIRONMENT="development"

echo "Starting server..."
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"
go run cmd/auth0-server/main.go &
SERVER_PID=$!

echo "Server PID: $SERVER_PID"
sleep 3

echo "Testing health endpoint..."
curl -s http://localhost:8080/health || echo "Health check failed"

echo "Stopping server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "Test completed"
