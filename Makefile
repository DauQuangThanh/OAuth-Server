# Makefile for Auth0-Server

.PHONY: help build run test clean setup deps lint format

# Default target
help:
	@echo "Auth0-Server Development Commands"
	@echo "================================="
	@echo ""
	@echo "Setup & Build:"
	@echo "  setup     - Set up the project (install deps, setup database)"
	@echo "  deps      - Download Go dependencies"
	@echo "  build     - Build the server binary"
	@echo "  clean     - Clean build artifacts and logs"
	@echo ""
	@echo "Development:"
	@echo "  run       - Run the server (with live reload if air is available)"
	@echo "  run-mem   - Run the server with in-memory storage"
	@echo "  run-pg    - Run the server with PostgreSQL"
	@echo ""
	@echo "Testing:"
	@echo "  test      - Run all tests"
	@echo "  test-api  - Run API integration tests"
	@echo "  test-pass - Test password security"
	@echo "  test-unit - Run unit tests"
	@echo ""
	@echo "Database:"
	@echo "  db-setup  - Set up the database"
	@echo "  db-debug  - Debug database connection"
	@echo "  db-reset  - Reset the database (drops and recreates)"
	@echo ""
	@echo "Security:"
	@echo "  jwe-key   - Generate a new JWE secret key"
	@echo "  verify    - Verify password security implementation"
	@echo ""
	@echo "Maintenance:"
	@echo "  lint      - Run linters"
	@echo "  format    - Format code"
	@echo "  tidy      - Tidy go modules"

# Variables
BINARY_NAME=auth0-server
BUILD_DIR=build
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

# Setup and dependencies
setup: deps db-setup
	@echo "✅ Project setup complete"

deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod verify

# Build
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/auth0-server/main.go
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean
clean:
	@echo "🧹 Cleaning up..."
	rm -rf $(BUILD_DIR)
	rm -rf logs/*
	rm -f *.log
	rm -f *.pid
	@echo "✅ Cleanup complete"

# Run
run:
	@if command -v air >/dev/null 2>&1; then \
		echo "🚀 Running with live reload (air)..."; \
		air; \
	else \
		echo "🚀 Running server..."; \
		go run cmd/auth0-server/main.go; \
	fi

run-mem:
	@echo "🚀 Running server with in-memory storage..."
	DB_DRIVER=memory go run cmd/auth0-server/main.go

run-pg:
	@echo "🚀 Running server with PostgreSQL..."
	DB_DRIVER=postgres go run cmd/auth0-server/main.go

# Testing
test: test-unit test-api

test-unit:
	@echo "🧪 Running unit tests..."
	go test -v ./internal/... ./pkg/...

test-api:
	@echo "🧪 Running API tests..."
	@chmod +x tests/api/test_api.sh
	./tests/api/test_api.sh

test-pass:
	@echo "🔒 Testing password security..."
	@chmod +x scripts/security/verify_password_security.sh
	./scripts/security/verify_password_security.sh

test-jwe:
	@echo "🔐 Testing JWE encryption..."
	@chmod +x tests/api/test_jwe.sh
	./tests/api/test_jwe.sh

# Database
db-setup:
	@echo "🗄️ Setting up database..."
	@chmod +x scripts/database/setup_database.sh
	./scripts/database/setup_database.sh

db-debug:
	@echo "🔍 Debugging database connection..."
	@chmod +x scripts/database/debug_postgres.sh
	./scripts/database/debug_postgres.sh

db-reset:
	@echo "⚠️ Resetting database..."
	@chmod +x scripts/database/fix_schema.sh
	./scripts/database/fix_schema.sh

# Security
jwe-key:
	@echo "🔐 Generating JWE secret key..."
	@chmod +x scripts/security/generate_jwe_secret.sh
	./scripts/security/generate_jwe_secret.sh

verify:
	@echo "🔒 Verifying security implementation..."
	@chmod +x scripts/security/verify_password_security.sh
	./scripts/security/verify_password_security.sh

# Code quality
lint:
	@echo "🔍 Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

format:
	@echo "📝 Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w $(GO_FILES); \
	else \
		echo "💡 Install goimports for better formatting: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

tidy:
	@echo "📦 Tidying modules..."
	go mod tidy

# Development helpers
install-tools:
	@echo "🛠️ Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Development tools installed"

# Quick development workflow
dev: deps lint test run

# CI/Production workflow  
ci: deps lint test build
