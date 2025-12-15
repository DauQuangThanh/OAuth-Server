#!/bin/bash

# Comprehensive PostgreSQL Database Setup Script for Auth0-Server
# Combines database creation, schema setup, and troubleshooting capabilities
# Handles both initial setup and schema fixes

set -e

# Default configuration (matches .env defaults)
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-}
DB_NAME=${DB_NAME:-auth0_db}

# Script options
FORCE_RECREATE=${FORCE_RECREATE:-false}
VERBOSE=${VERBOSE:-false}

# Helper functions
log_info() {
    echo "ℹ️  $1"
}

log_success() {
    echo "✅ $1"
}

log_error() {
    echo "❌ $1"
}

log_warning() {
    echo "⚠️  $1"
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Comprehensive PostgreSQL database setup for Auth0-Server"
    echo ""
    echo "Options:"
    echo "  --force-recreate    Drop and recreate existing database and tables"
    echo "  --verbose          Show detailed output"
    echo "  --help             Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  DB_HOST            PostgreSQL host (default: localhost)"
    echo "  DB_PORT            PostgreSQL port (default: 5432)"
    echo "  DB_USER            PostgreSQL user (default: postgres)"
    echo "  DB_PASSWORD        PostgreSQL password (default: empty)"
    echo "  DB_NAME            Database name (default: Auth0_DB)"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Normal setup"
    echo "  $0 --force-recreate                  # Force recreate everything"
    echo "  VERBOSE=true $0                      # Verbose mode"
    echo "  DB_NAME=myauth $0                    # Custom database name"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --force-recreate)
            FORCE_RECREATE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main setup
echo "🗄️  PostgreSQL Database Setup for Auth0-Server"
echo "=============================================="
echo

if [[ "$VERBOSE" == "true" ]]; then
    log_info "Running in verbose mode"
fi

if [[ "$FORCE_RECREATE" == "true" ]]; then
    log_warning "Force recreate mode enabled - will drop existing data!"
fi

echo "Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo "  Force recreate: $FORCE_RECREATE"
echo

# Step 1: Check PostgreSQL client tools
echo "1. Checking PostgreSQL client tools..."
if ! command -v psql >/dev/null 2>&1; then
    log_error "psql command not found. Please install PostgreSQL client tools."
    echo "   On macOS: brew install postgresql"
    echo "   On Ubuntu/Debian: sudo apt-get install postgresql-client"
    echo "   On CentOS/RHEL: sudo yum install postgresql"
    exit 1
fi
log_success "PostgreSQL client tools found"

# Step 2: Test connection to PostgreSQL server
echo "2. Testing PostgreSQL server connection..."
if [[ "$VERBOSE" == "true" ]]; then
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "SELECT version();"
else
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "SELECT version();" >/dev/null 2>&1
fi

if [ $? -ne 0 ]; then
    log_error "Cannot connect to PostgreSQL server."
    echo "   Please ensure PostgreSQL is running and connection parameters are correct."
    echo "   Connection string: postgresql://$DB_USER@$DB_HOST:$DB_PORT/postgres"
    echo ""
    echo "   Common fixes:"
    echo "   - Start PostgreSQL: brew services start postgresql (macOS)"
    echo "   - Check if PostgreSQL is running: ps aux | grep postgres"
    echo "   - Verify port is open: netstat -an | grep $DB_PORT"
    exit 1
fi
log_success "PostgreSQL server connection successful"

# Step 3: Handle database creation/recreation
echo "3. Managing database '$DB_NAME'..."
DB_EXISTS=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")

if [[ "$FORCE_RECREATE" == "true" && "$DB_EXISTS" == "1" ]]; then
    log_warning "Dropping existing database '$DB_NAME'..."
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS \"$DB_NAME\";"
    DB_EXISTS=""
fi

if [ "$DB_EXISTS" = "1" ]; then
    log_success "Database '$DB_NAME' already exists"
else
    log_info "Creating database '$DB_NAME'..."
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE \"$DB_NAME\";"
    log_success "Database '$DB_NAME' created successfully"
fi

# Step 4: Test connection to target database
echo "4. Testing connection to database '$DB_NAME'..."
if [[ "$VERBOSE" == "true" ]]; then
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT current_database(), current_user;"
else
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT current_database(), current_user;" >/dev/null 2>&1
fi

if [ $? -eq 0 ]; then
    log_success "Successfully connected to database '$DB_NAME'"
else
    log_error "Failed to connect to database '$DB_NAME'"
    exit 1
fi

# Step 5: Check if schema file exists
echo "5. Locating database schema..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA_FILE="$SCRIPT_DIR/../../database/schema.sql"

if [ ! -f "$SCHEMA_FILE" ]; then
    log_error "Schema file not found at $SCHEMA_FILE"
    log_info "Looking for alternative locations..."
    
    # Try alternative paths
    ALT_PATHS=(
        "$(pwd)/database/schema.sql"
        "$(dirname "$0")/database/schema.sql"
        "./schema.sql"
    )
    
    for alt_path in "${ALT_PATHS[@]}"; do
        if [ -f "$alt_path" ]; then
            SCHEMA_FILE="$alt_path"
            log_success "Found schema file at $alt_path"
            break
        fi
    done
    
    if [ ! -f "$SCHEMA_FILE" ]; then
        log_error "Could not locate schema.sql file"
        echo "   Please ensure the database/schema.sql file exists"
        exit 1
    fi
fi

log_success "Schema file found at $SCHEMA_FILE"

# Step 6: Handle table creation/recreation
echo "6. Managing database schema..."
TABLE_EXISTS=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users';")

if [[ "$FORCE_RECREATE" == "true" && "$TABLE_EXISTS" == "1" ]]; then
    log_warning "Dropping existing tables..."
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "DROP TABLE IF EXISTS users CASCADE;"
    TABLE_EXISTS=""
fi

log_info "Creating/updating database schema..."
if [[ "$VERBOSE" == "true" ]]; then
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SCHEMA_FILE"
else
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SCHEMA_FILE" >/dev/null 2>&1
fi

if [ $? -eq 0 ]; then
    log_success "Database schema applied successfully"
else
    log_error "Failed to create database schema"
    exit 1
fi

# Step 7: Verify table creation
echo "7. Verifying table structure..."
TABLE_COUNT=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'accounts';")

if [ "$TABLE_COUNT" = "1" ]; then
    log_success "Accounts table verified successfully"
else
    log_error "Accounts table verification failed"
    exit 1
fi

# Step 8: Show detailed verification
echo "8. Database verification and information..."

# Show column structure
log_info "Table structure:"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
    SELECT 
        column_name,
        data_type,
        is_nullable,
        column_default
    FROM information_schema.columns 
    WHERE table_schema = 'public' AND table_name = 'accounts'
    ORDER BY ordinal_position;
"

# Show indexes
log_info "Table indexes:"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
    SELECT 
        indexname,
        indexdef
    FROM pg_indexes 
    WHERE tablename = 'users';
"

# Show user count
USER_COUNT=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM users;")
log_info "Current user count: $USER_COUNT"

# Step 9: Connection test
echo "9. Final connection test..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
    SELECT 
        current_database() as database,
        current_user as user,
        inet_server_addr() as server_ip,
        inet_server_port() as server_port,
        version() as postgresql_version;
" 2>/dev/null

echo
echo "🎉 Database setup completed successfully!"
echo
log_success "All database components are ready"
echo
echo "🔐 Security Information:"
echo "   ✅ Passwords use bcrypt hashing (cost factor: 10)"
echo "   ✅ No plaintext passwords stored"
echo "   ✅ Unique salt per password"
echo "   ✅ Timing attack protection"
echo
echo "📋 Next Steps:"
echo "1. Start the Auth0-Server:"
echo "   export DB_DRIVER=postgres"
echo "   go run cmd/auth0-server/main.go"
echo ""
echo "2. Test the setup:"
echo "   ./verify_project.sh"
echo ""
echo "3. Run API tests:"
echo "   ./tests/api/test_api.sh"
echo
echo "🔧 Connection Information:"
echo "   Connection string: postgresql://$DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
echo "   Manual connection: psql postgresql://$DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
echo
echo "🔨 Troubleshooting:"
echo "   - For schema issues: $0 --force-recreate"
echo "   - For verbose output: $0 --verbose"
echo "   - For help: $0 --help"
echo
