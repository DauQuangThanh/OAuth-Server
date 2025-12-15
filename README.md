# OAuth 2.1 Compliant Auth0-Compatible Authentication Server

A production-ready, OAuth 2.1 compliant authentication server that implements Auth0-compatible APIs, built with Go and designed for high performance, scalability, and security best practices.

**Specification Compliance:**
- [OAuth 2.1 (draft-ietf-oauth-v2-1-14)](https://datatracker.ietf.org/doc/draft-ietf-oauth-v2-1/) - October 2025
- [RFC 9700](https://datatracker.ietf.org/doc/rfc9700/) - OAuth 2.0 Security Best Current Practice (January 2025)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- [OpenID Connect Discovery 1.0](https://openid.net/specs/openid-connect-discovery-1_0.html)

## Overview

This project provides a complete OAuth 2.1 compliant authentication solution compatible with Auth0's API endpoints. It implements only secure OAuth 2.1 flows (authorization code with PKCE) while removing deprecated flows (password/implicit grants), allowing you to run modern, secure authentication services locally or in your own infrastructure while maintaining compatibility with Auth0 client libraries and integrations.

## Features

### 🔐 Authentication & Authorization
- **OAuth 2.1 Compliant** - Authorization code flow with mandatory PKCE (S256)
- **JWT Token Management** - Access tokens and refresh tokens with JWE encryption
- **Account Management** - User registration, profile management
- **Auth0 API Compatibility** - Drop-in replacement for Auth0 endpoints

### 🏗️ Architecture
- **Clean Architecture** - Separation of concerns with dependency injection
- **Domain-Driven Design** - Clear domain boundaries and business logic
- **High Concurrency** - Built for scalability with Go's goroutines
- **Database Agnostic** - Support for PostgreSQL and in-memory storage

### 🚀 Performance & Scalability
- **Worker Pool Pattern** - Efficient background task processing
- **Connection Pooling** - Optimized database connections
- **Caching Layer** - In-memory and Redis caching support
- **Metrics & Monitoring** - Built-in health checks and performance metrics

### 🔧 Development Features
- **Hot Reload** - Development mode with automatic restarts
- **Comprehensive Logging** - Structured logging with context
- **Error Handling** - Detailed error responses and recovery
- **Testing Suite** - Integration and unit test support

## Quick Start

### Prerequisites
- Go 1.24+ 
- PostgreSQL (optional - memory database available for development)
- Make (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd Auth0-Server
   ```

2. **Set up environment variables**
   ```bash
   export JWE_SECRET="your-32-character-secret-key-here"
   export DB_DRIVER="memory"  # or "postgres"
   export ENVIRONMENT="development"
   ```

3. **Build and run**
   ```bash
   make build
   make run
   ```

   Or manually:
   ```bash
   go build -o auth0-server cmd/auth0-server/main.go
   ./auth0-server
   ```

### Database Setup (PostgreSQL)

If using PostgreSQL:

1. **Create database**
   ```bash
   createdb Auth0_DB
   ```

2. **Run setup script**
   ```bash
   ./scripts/database/comprehensive_setup.sh
   ```

3. **Configure environment**
   ```bash
   export DB_DRIVER="postgres"
   export DB_HOST="localhost"
   export DB_PORT="5432"
   export DB_USER="postgres"
   export DB_PASSWORD=""
   export DB_NAME="Auth0_DB"
   ```

## API Endpoints

### Authentication Endpoints

#### User Registration
```bash
POST /dbconnections/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe"
}
```

#### OAuth 2.1 Authorization Flow
```bash
# 1. Start authorization (redirect user to this URL)
GET /authorize?response_type=code&client_id=your-client-id&redirect_uri=http://localhost:3000/callback&code_challenge=CHALLENGE&code_challenge_method=S256&scope=openid+email+profile&state=random-state

# 2. Exchange authorization code for tokens
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=AUTH_CODE&client_id=your-client-id&code_verifier=VERIFIER&redirect_uri=http://localhost:3000/callback
```

#### User Information
```bash
GET /userinfo
Authorization: Bearer <access_token>
```

### Configuration Endpoints

#### OpenID Configuration
```bash
GET /.well-known/openid_configuration
```

#### Health Check
```bash
GET /health
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `JWE_SECRET` | 32-character secret for token encryption | - | ✅ |
| `DB_DRIVER` | Database driver ("memory" or "postgres") | "memory" | ❌ |
| `DB_HOST` | PostgreSQL host | "localhost" | ❌ |
| `DB_PORT` | PostgreSQL port | "5432" | ❌ |
| `DB_USER` | PostgreSQL username | "postgres" | ❌ |
| `DB_PASSWORD` | PostgreSQL password | "" | ❌ |
| `DB_NAME` | Database name | "auth0_db" | ❌ |
| `SERVER_ADDRESS` | Server bind address | ":8080" | ❌ |
| `ENVIRONMENT` | Environment mode | "development" | ❌ |

### Advanced Configuration

The server supports extensive configuration through environment variables:

#### Database Configuration
- Connection pooling settings
- SSL mode configuration
- Connection timeouts

#### Security Configuration
- HTTPS settings
- Token expiration times
- Rate limiting

#### Monitoring Configuration
- Metrics collection
- Health check intervals
- Tracing support

See `internal/config/enhanced.go` for all available options.

## Database Setup

### PostgreSQL Setup

To use PostgreSQL as the database backend:

1. **Install PostgreSQL** (if not already installed):
   ```bash
   # macOS with Homebrew
   brew install postgresql@16
   brew services start postgresql@16
   
   # Ubuntu/Debian
   sudo apt-get install postgresql postgresql-contrib
   
   # CentOS/RHEL
   sudo yum install postgresql-server postgresql-contrib
   ```

2. **Create Database and Schema**:
   ```bash
   # Run the comprehensive setup script
   ./scripts/database/comprehensive_setup.sh
   
   # Or manually:
   createdb auth0_db
   psql auth0_db < database/schema.sql
   ```

3. **Configure Environment Variables**:
   ```bash
   export DB_DRIVER=postgres
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=your_password
   export DB_NAME=auth0_db
   ```

### In-Memory Database

For development and testing, you can use the in-memory database:

```bash
export DB_DRIVER=memory
```

The in-memory database is automatically selected if PostgreSQL connection fails.

## Project Structure

```
Auth0-Server/
├── cmd/auth0-server/           # Application entry point
├── internal/
│   ├── application/            # Application layer
│   │   ├── ports/              # Interface definitions
│   │   └── usecases/           # Business logic
│   ├── domain/                 # Domain layer
│   │   ├── account/            # Account domain
│   │   └── auth/               # Authentication domain
│   ├── infrastructure/         # Infrastructure layer
│   │   ├── cache/              # Caching implementations
│   │   ├── crypto/             # Cryptography services
│   │   ├── monitoring/         # Metrics and health checks
│   │   ├── storage/            # Database implementations
│   │   └── workers/            # Background processing
│   ├── interfaces/             # Interface layer
│   │   └── http/               # HTTP handlers and middleware
│   ├── container/              # Dependency injection
│   └── config/                 # Configuration management
├── pkg/                        # Shared packages
│   ├── errors/                 # Error handling
│   ├── logger/                 # Logging utilities
│   └── server/                 # HTTP server
├── database/                   # Database schemas
├── scripts/                    # Utility scripts
└── tests/                      # Test suites
```

## Development

### Running Tests
```bash
make test
# or
go test ./...
```

### Verification Script
```bash
./verify_project.sh
```

This script performs comprehensive testing including:
- Build verification
- Unit tests execution
- Server startup testing
- API endpoint testing
- Project structure validation

### Code Generation
```bash
make generate  # Generate JWE secret
make clean     # Clean build artifacts
```

## Security

### OAuth 2.1 Security (draft-ietf-oauth-v2-1-14)
- PKCE mandatory for all authorization code flows (S256 only)
- No password grant (Resource Owner Password Credentials removed)
- No implicit grant (response_type=token removed)
- Exact redirect URI matching
- One-time use authorization codes

### Password Security
- Bcrypt hashing with configurable cost factor
- Minimum 8-character password requirement
- Constant-time password comparison

### Token Security
- JWE (JSON Web Encryption) for token encryption
- JWT signing with HMAC-SHA256
- Short-lived access tokens (24 hours)
- Secure refresh token rotation

### API Security (RFC 9700)
- Per-IP rate limiting
- CORS protection
- Security headers
- No bearer tokens in query strings

## Migration from Auth0

This server provides Auth0-compatible endpoints, making migration straightforward:

1. **Update endpoints** - Point your client to the new server
2. **Migrate data** - Export users from Auth0 and import to accounts table
3. **Update configuration** - Set environment variables to match your Auth0 settings

## Monitoring & Observability

### Health Checks
- Database connectivity
- Service health status
- Performance metrics

### Metrics
- Request rates and latency
- Authentication success/failure rates
- Database connection statistics
- Account statistics

### Logging
- Structured JSON logging
- Request tracing
- Error tracking with context

## Production Deployment

### Environment Setup
```bash
export ENVIRONMENT="production"
export JWE_SECRET="$(./scripts/security/generate_jwe_secret.sh)"
export DB_DRIVER="postgres"
# Set other production configurations
```

### Docker Support
```bash
# Build Docker image
docker build -t auth0-server .

# Run with environment file
docker run --env-file .env -p 8080:8080 auth0-server
```

### Performance Tuning
- Configure worker pool size
- Adjust database connection limits
- Enable caching for high-traffic scenarios
- Set appropriate timeout values

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the verification script
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
1. Check the documentation
2. Run the verification script to diagnose issues
3. Review logs for error details
4. Open an issue with detailed information

## Changelog

### v2.0.0 - Account-Based Architecture
- Migrated from user-based to account-based model
- Removed account_code functionality for simplified design
- Enhanced PostgreSQL support
- Improved error handling and logging
- Added comprehensive verification suite

### v1.0.0 - Initial Release
- Auth0-compatible API implementation
- PostgreSQL and memory database support
- JWT token management
- Clean architecture implementation

## ✨ Key Features

### 🔐 OAuth 2.1 & Auth0 Compatibility
- **OAuth 2.1 Compliant**: Full compliance with OAuth 2.1 security best practices
- **Authorization Code + PKCE**: Only secure flow supported (S256 challenge method)
- **No Deprecated Flows**: Password and implicit grants completely removed
- **Auth0 API Compatibility**: Seamless migration from Auth0 with compatible endpoints
- **OpenID Connect**: Full OIDC discovery support with OAuth 2.1 configuration
- **JWT/JWE Support**: Both JWT signing and JWE encryption for maximum security

### 🚀 High Performance & Concurrency
- **Worker Pool System**: Configurable worker pools for background task processing
- **Object Pooling**: Minimized garbage collection through crypto object pooling
- **Concurrent Operations**: Read-write locks and atomic operations for optimal performance
- **Non-blocking Architecture**: Async processing where appropriate

### 🏗️ Clean Architecture
- **Hexagonal Architecture**: Domain-driven design with ports & adapters
- **Dependency Injection**: Clean separation of concerns and testable components
- **Layer Separation**: Domain, Application, Infrastructure, and Interface layers

### � Observability & Monitoring
- **Metrics Collection**: Request metrics, error rates, performance statistics
- **Distributed Tracing**: Request tracing with span correlation
- **Health Checks**: Comprehensive system health monitoring
- **Structured Logging**: JSON logging for production environments

### 🛡️ Enterprise Security
- **Password Security**: bcrypt hashing with configurable cost factor and salt
- **JWE Encryption**: Token encryption in addition to JWT signing
- **Rate Limiting**: Configurable per-IP rate limiting
- **Security Headers**: Comprehensive security header middleware
- **Input Validation**: Structured validation throughout the system

### 🔧 Production Features
- **Graceful Shutdown**: Proper resource cleanup and connection draining
- **Configuration Management**: Environment-based configuration with validation
- **Caching System**: Multi-level caching with TTL support
- **Resource Management**: Configurable limits and timeouts

## 🏗️ Project Structure

The project follows a clean, modular structure for maintainability and scalability:

```
Auth0-Server/
├── cmd/
│   └── auth0-server/     # Application entry point
├── internal/             # Core application code
│   ├── application/      # Use cases and business logic
│   │   ├── ports/       # Interfaces for dependency inversion
│   │   └── usecases/    # Application business logic
│   ├── domain/          # Core domain entities
│   └── infrastructure/ # External dependencies
│       ├── cache/       # Caching implementations
│       ├── crypto/      # Cryptographic services
│       ├── monitoring/  # Metrics and health checks
│       ├── storage/     # Data persistence
│       ├── tracing/     # Distributed tracing
│       └── workers/     # Background task processing
├── pkg/                 # Reusable packages
│   ├── errors/          # Error handling
│   ├── logger/          # Structured logging
│   └── server/          # HTTP server
├── docs/                # Documentation
├── scripts/             # Utility scripts
│   ├── database/        # Database setup and tools
│   ├── security/        # Security tools and verification
│   └── tools/           # Development utilities
├── tests/               # Test files
│   ├── api/            # API integration tests
│   ├── integration/    # Integration tests
│   └── data/           # Test data
├── database/           # Database schemas and migrations
├── config/             # Configuration files
└── logs/               # Application logs (git ignored)
```

## 🏗️ Architecture

For detailed architecture documentation, see [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md).

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 17+ (running locally)
- Optional: curl for testing

### Setup & Run

1. **Setup PostgreSQL Database**:
   ```bash
   # Make sure PostgreSQL is running
   # Create database and tables (REQUIRED for PostgreSQL mode)
   ./scripts/database/setup_database.sh
   ```

   **Note**: You MUST run `scripts/database/setup_database.sh` before starting the server in PostgreSQL mode. The script creates the database and all required tables.

2. **Clone and configure**:
   ```bash
   cd /path/to/auth0-server
   cp config/.env.example .env
   
   # Generate a secure JWE secret (recommended)
   ./scripts/security/generate_jwe_secret.sh
   
   # Or generate manually with Go
   go run scripts/tools/generate_jwe_secret.go
   
   # Or edit .env and set JWE_SECRET manually
   ```

3. **Run the server**:
   ```bash
   # The server will automatically:
   # - Create the database if it doesn't exist
   # - Initialize the required tables on first run
   go run cmd/auth0-server/main.go
   ```

4. **Run with enhanced features**:
   ```bash
   # Configure enhanced features in .env
   export ENVIRONMENT=development
   export ENABLE_METRICS=true
   export ENABLE_TRACING=true
   export WORKER_POOL_SIZE=10
   
   go run cmd/auth0-server/main.go
   ```

5. **Verify the installation**:
   ```bash
   # Run comprehensive verification script
   chmod +x verify_project.sh && ./verify_project.sh
   ```

### Testing

Run the comprehensive test suites:

```bash
# Test OAuth 2.1 compliance and all endpoints
chmod +x verify_project.sh && ./verify_project.sh

# Test API endpoints with enhanced features
chmod +x tests/api/test_api.sh && ./tests/api/test_api.sh

# Test JWE encryption specifically  
chmod +x tests/api/test_jwe.sh && ./tests/api/test_jwe.sh

# Test password security
chmod +x scripts/security/verify_password_security.sh && ./scripts/security/verify_password_security.sh
```

The `verify_project.sh` script now includes comprehensive OAuth 2.1 compliance testing:
- OAuth 2.1 configuration validation
- PKCE enforcement testing
- Deprecated grant rejection verification
- Authorization endpoint testing
## 📡 API Endpoints

### Authentication Endpoints

#### `GET /authorize`
OAuth 2.1 authorization endpoint with mandatory PKCE.

**Query Parameters**:
```
response_type=code          (required)
client_id=your-client-id    (required)
redirect_uri=callback-url   (required)
code_challenge=CHALLENGE    (required, base64url-encoded SHA256)
code_challenge_method=S256  (required, only S256 supported)
scope=openid+email+profile  (optional)
state=random-state          (recommended)
```

**Response**: Redirects to `redirect_uri` with authorization code:
```
http://callback-url?code=AUTHORIZATION_CODE&state=random-state
```

#### `POST /authorize`
Complete authorization flow with user credentials (internal form submission).

**Form Data**:
```
email=user@example.com
password=userpassword
(plus all original query parameters)
```

#### `POST /oauth/token`
OAuth 2.1 token endpoint for authorization code exchange with PKCE validation.

**Request** (form-encoded):
```
grant_type=authorization_code
code=AUTHORIZATION_CODE
client_id=your-client-id
code_verifier=CODE_VERIFIER
redirect_uri=http://localhost:3000/callback
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2In0...", 
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2In0...",
  "scope": "openid profile email"
}
```
*Note: Tokens are JWE encrypted for enhanced security*

#### `GET /userinfo`
Get authenticated user information (Protected endpoint).

**Headers**:
```
Authorization: Bearer <access_token>
```

**Response**:
```json
{
  "sub": "user_id",
  "email": "user@example.com", 
  "email_verified": true,
  "name": "User Name",
  "nickname": "user",
  "picture": ""
}
```

#### `POST /dbconnections/signup`
Register a new user with enhanced validation.

**Request**:
```json
{
  "email": "newuser@example.com",
  "password": "securepassword",
  "name": "New User"
}
```

**Response**:
```json
{
  "_id": "user_id",
  "email": "newuser@example.com",
  "name": "New User", 
  "email_verified": true,
  "created_at": "2025-07-04T12:00:00Z"
}
```

#### `GET /api/v2/users`
List users (Protected endpoint - requires authentication).

**Headers**:
```
Authorization: Bearer <access_token>
```

### Discovery & Monitoring Endpoints

#### `GET /.well-known/openid_configuration`
OpenID Connect discovery document with enhanced metadata.

#### `GET /health`
Comprehensive health check with system metrics.

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2025-07-04T12:00:00Z", 
  "checks": { "user_repository": "healthy", "cache": "healthy" },
  "system": { "goroutines": 45, "memory": { "alloc_mb": 12 } },
  "metrics": { "requests": { "total": 1247 } }
}
```

#### `GET /metrics`
Prometheus-compatible metrics endpoint (when ENABLE_METRICS=true).

#### `GET /debug/config` 
Debug configuration endpoint (development mode only).

## 🔧 Usage Examples

### Register a New User
```bash
curl -X POST http://localhost:8080/dbconnections/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "mypassword123",
    "name": "Test User"
  }'
```

### OAuth 2.1 Authorization Flow

#### 1. Generate PKCE Parameters
```bash
# Generate code verifier and challenge for PKCE
CODE_VERIFIER=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-43)
CODE_CHALLENGE=$(echo -n $CODE_VERIFIER | openssl dgst -binary -sha256 | openssl base64 | tr -d "=+/" | cut -c1-43)
```

#### 2. Start Authorization (Redirect User)
```bash
# User visits this URL in browser to login
http://localhost:8080/authorize?response_type=code&client_id=your-client-id&redirect_uri=http://localhost:3000/callback&code_challenge=$CODE_CHALLENGE&code_challenge_method=S256&scope=openid+email+profile&state=random-state
```

#### 3. Exchange Authorization Code for Tokens
```bash
# After user completes login, exchange the code for tokens
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=AUTHORIZATION_CODE&client_id=your-client-id&code_verifier=$CODE_VERIFIER&redirect_uri=http://localhost:3000/callback"
```

### Get User Information
```bash
curl -X GET http://localhost:8080/userinfo \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN_HERE"
```

## 🔐 Password Security

The OAuth 2.1 compliant server implements **enterprise-grade password security** for internal user authentication (login form):

- **bcrypt Hashing**: Industry-standard password hashing with built-in salt
- **Configurable Cost**: Default bcrypt cost factor of 10 (can be adjusted)
- **No Plaintext Storage**: Passwords are never stored in plaintext anywhere
- **Memory Safety**: Object pooling and secure memory handling
- **Validation**: Minimum 8-character password requirement
- **OAuth 2.1 Usage Only**: Passwords only used for authorization flow login form (not password grant)

**Verification**: Use `./verify_password_security.sh` to test password hashing functionality.

**Documentation**: See [PASSWORD_SECURITY.md](./PASSWORD_SECURITY.md) for detailed implementation.

## Configuration

Configure the server using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `JWE_SECRET` | Secret key for JWE token encryption (required) | - |
| `SERVER_ADDRESS` | Server listening address | `:8080` |
| `ISSUER` | JWT issuer claim | `auth0-server` |
| `DOMAIN` | Domain for OIDC discovery | `localhost:8080` |

### 🔐 Generating Secure JWE Secrets

Use the built-in generators to create cryptographically secure secrets:

**Shell Script (Recommended)**:
```bash
# Generate and optionally update .env automatically
./generate_jwe_secret.sh

# Generate with specific key size (default: 32 bytes/256-bit)
./generate_jwe_secret.sh 32
```

**Go Program**:
```bash
# Generate with default settings
go run generate_jwe_secret.go

# Generate with specific key size
go run generate_jwe_secret.go 32
```

Both generators create multiple format options:
- **Hexadecimal** (recommended for .env files)
- **Base64** (alternative format)
- **Base64 URL-safe** (web-safe encoding)

**Security Notes**:
- Use at least 256-bit (32 bytes) keys for production
- Store secrets securely (environment variables, key management systems)
- Never commit secrets to version control
- Rotate secrets regularly
- Use different secrets for different environments

### 🗄️ PostgreSQL Database Configuration

The server uses PostgreSQL 17+ as the primary database with the following default settings:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL server host | `localhost` |
| `DB_PORT` | PostgreSQL server port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `` (empty) |
| `DB_NAME` | Database name | `Auth0_DB` |
| `DB_SSL_MODE` | SSL mode | `disable` |

### 🗄️ Database Configuration

The server supports both PostgreSQL and in-memory storage:

**PostgreSQL (Production)**:
```bash
# Set in .env or as environment variables
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=Auth0_DB
```

**In-Memory Storage (Development/Testing)**:
```bash
# Set in .env or as environment variables
DB_DRIVER=memory
```

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_DRIVER` | Database driver (`postgres` or `memory`) | `memory` |
| `DB_HOST` | PostgreSQL server host | `localhost` |
| `DB_PORT` | PostgreSQL server port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `` (empty) |
| `DB_NAME` | Database name | `Auth0_DB` |
| `DB_SSL_MODE` | SSL mode | `disable` |

**Automatic Fallback**: If PostgreSQL connection fails, the server automatically falls back to in-memory storage with a warning message.

**Database Features**:
- **Schema Management**: Tables must be created using `setup_database.sh` before starting the server
- **Connection Pooling**: Configured for high-performance with 25 max connections  
- **Health Checks**: Database connectivity monitoring
- **Migration Ready**: Schema can be easily extended for new features

**Setup PostgreSQL**:
```bash
# REQUIRED: Run the setup script to create database and tables
./setup_database.sh

# Or create manually with psql
createdb -U postgres Auth0_DB
psql -U postgres -d Auth0_DB -f schema.sql
```

**For Testing/Development without PostgreSQL**:
```bash
# The server will use in-memory storage automatically
export DB_DRIVER=memory
go run cmd/auth0-server/main.go
```

## Project Structure

```
auth0-server/
├── cmd/auth0-server/           # Application entry point
│   └── main.go
├── internal/                   # Private application code
│   ├── application/           # Application layer (use cases)
│   │   └── usecases/         # Business logic use cases
│   ├── domain/               # Domain layer (entities & interfaces)
│   │   ├── auth/            # Authentication domain
│   │   └── user/            # User domain
│   ├── infrastructure/       # Infrastructure layer
│   │   ├── crypto/          # Cryptographic services (JWE, passwords)
│   │   └── storage/         # Data storage implementations
│   └── interfaces/          # Interface layer
│       └── http/           # HTTP interface
│           ├── handlers/   # HTTP request handlers
│           └── middleware/ # HTTP middleware
├── pkg/                     # Public packages
│   ├── errors/             # Error definitions
│   ├── logger/             # Logging interface
│   └── server/             # HTTP server utilities
├── .env.example           # Environment configuration template
├── setup_database.sh      # PostgreSQL database and schema setup script
├── schema.sql             # Database schema definition
├── DATABASE.md            # Database management documentation
├── go.mod                # Go module definition
└── README.md            # This file
```

## Architecture Highlights

- **Clean Architecture**: Follows Domain-Driven Design principles with clear separation of concerns
- **High Concurrency**: Optimized for handling thousands of concurrent requests with worker pools
- **Dependency Injection**: Loose coupling through interface-based dependency injection
- **Context Propagation**: Proper context usage for request tracing and cancellation
- **Concurrent-Safe Storage**: Thread-safe repository with dedicated write workers
- **Object Pooling**: JWE service uses object pools for better memory efficiency
- **Graceful Shutdown**: Proper cleanup and graceful server shutdown
- **Middleware Chain**: Composable middleware for cross-cutting concerns
- **Rate Limiting**: Built-in rate limiting for API protection

## Security Features

- **OAuth 2.1 Compliance**: Full compliance with [draft-ietf-oauth-v2-1-14](https://datatracker.ietf.org/doc/draft-ietf-oauth-v2-1/) (October 2025)
- **RFC 9700 Compliance**: Implements [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/rfc9700/) (January 2025)
- **PKCE Mandatory**: S256 code challenges required for all authorization flows (OAuth 2.1 Section 7.6)
- **No Deprecated Flows**: Password and implicit grants removed per OAuth 2.1 specification
- **Password Hashing**: Uses bcrypt with configurable cost for internal authentication
- **JWE Security**: JSON Web Encryption with AES-256-GCM encryption and HMAC-SHA256 signing
- **Authorization Codes**: Temporary, one-time use codes with 10-minute expiration
- **Exact Redirect URI Matching**: Per OAuth 2.1 Section 7.5.3 requirements
- **No Query String Tokens**: Bearer tokens only in Authorization header (RFC 9700 Section 4.3.2)
- **CORS Protection**: Configurable CORS policies
- **Input Validation**: Proper request validation and sanitization
- **Error Handling**: Secure error responses without information leakage

## Performance Characteristics

- **High Concurrency**: Handles thousands of concurrent requests efficiently
- **Low Memory Footprint**: Minimal resource usage
- **Fast Startup**: Quick server initialization
- **Scalable Design**: Ready for horizontal scaling

## 📋 OAuth 2.1 Compliance Documentation

This server is fully compliant with **OAuth 2.1 (draft-ietf-oauth-v2-1-14)** and **RFC 9700** security standards. See the comprehensive documentation:

- **[OAUTH_2_1_COMPLIANCE.md](./OAUTH_2_1_COMPLIANCE.md)** - Complete compliance checklist and implementation details
- **[OAUTH_2_1_TESTING.md](./OAUTH_2_1_TESTING.md)** - OAuth 2.1 flow testing guide with curl examples
- **[docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)** - OAuth 2.1 architecture and security design

### OAuth 2.1 Compliance Features (per draft-ietf-oauth-v2-1-14)
- ✅ **Authorization Code Flow with PKCE** - Only secure flow supported
- ✅ **No Password Grant** - Resource Owner Password Credentials grant removed (Section 2.4)
- ✅ **No Implicit Grant** - `response_type=token` removed (Section 2.1.2)
- ✅ **PKCE Mandatory** - S256 challenge method required for all clients (Section 7.6)
- ✅ **Exact Redirect URI Matching** - Per Section 7.5.3
- ✅ **No Bearer Tokens in Query String** - Tokens only in Authorization header (Section 5.2)
- ✅ **One-Time Authorization Codes** - Codes expire after single use (Section 4.1.2)
- ✅ **Secure by Default** - Implements RFC 9700 security best practices

## Development

### Using the Makefile

The project includes a comprehensive Makefile for development:

```bash
# Quick setup
make setup          # Install dependencies and setup database
make run             # Run with live reload (if air is installed)
make test            # Run all tests

# Database operations
make db-setup       # Set up PostgreSQL database
make db-debug       # Debug database connection
make db-reset       # Reset database schema

# Testing
make test-api       # API integration tests
make test-pass      # Password security verification
make test-unit      # Unit tests

# Security
make jwe-key        # Generate JWE secret key
make verify         # Verify security implementation

# Code quality
make lint           # Run linters
make format         # Format code
make clean          # Clean build artifacts
```

Run `make help` for all available commands.

### Manual Commands

```bash
# Install dependencies
go mod download

# Run with live reload (if air is installed)
air

# Generate new JWE secret
./scripts/security/generate_jwe_secret.sh

# Set up database
./scripts/database/setup_database.sh
```

### Running Tests
```bash
go test ./...
```

### Building for Production
```bash
go build -o auth0-server ./cmd/auth0-server
```

### VS Code Integration
Use the included VS Code task "Run Auth0 Server" or press `Ctrl+Shift+P` and search for "Tasks: Run Task".

## Extending the Server

The modular architecture makes it easy to extend:

1. **Add new endpoints**: Create handlers in `internal/interfaces/http/handlers/`
2. **Add middleware**: Implement in `internal/interfaces/http/middleware/`
3. **Add storage backends**: Implement interfaces in `internal/application/ports/`
4. **Add authentication methods**: Extend the authentication logic

## Production Considerations

For production deployment:

1. **Use a real database** instead of in-memory storage
2. **Implement rate limiting** for API endpoints  
3. **Add logging and monitoring** capabilities
4. **Use HTTPS** with proper TLS certificates
5. **Implement token rotation** and revocation
6. **Add user management features** (password reset, etc.)
7. **Configure proper CORS policies** for your domains

## License

MIT License - feel free to use this in your projects!
