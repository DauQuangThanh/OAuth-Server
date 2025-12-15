#!/bin/bash

# Generate JWE Secret Script
# Generates a cryptographically secure 256-bit (32-byte) key for JWE token encryption
# This script creates a random hex-encoded secret suitable for production use

set -e

# Function to generate secure random hex string
generate_hex_secret() {
    local bytes=${1:-32}  # Default to 32 bytes (256 bits)
    
    # Try different methods based on available tools
    if command -v openssl >/dev/null 2>&1; then
        # Use OpenSSL (most reliable, available on most systems)
        openssl rand -hex "$bytes"
    elif command -v head >/dev/null 2>&1 && [[ -c /dev/urandom ]]; then
        # Use /dev/urandom with head and od (Unix/Linux systems)
        head -c "$bytes" /dev/urandom | od -An -tx1 | tr -d ' \n'
    elif command -v python3 >/dev/null 2>&1; then
        # Use Python3 as fallback
        python3 -c "import secrets; print(secrets.token_hex($bytes))"
    elif command -v python >/dev/null 2>&1; then
        # Use Python2 as fallback
        python -c "import os, binascii; print(binascii.hexlify(os.urandom($bytes)).decode())"
    else
        echo "Error: No suitable random generator found." >&2
        echo "Please install openssl, python, or ensure /dev/urandom is available." >&2
        exit 1
    fi
}

# Function to validate the generated secret
validate_secret() {
    local secret="$1"
    
    # Check if secret is exactly 64 characters (32 bytes in hex)
    if [[ ${#secret} -ne 64 ]]; then
        echo "Error: Generated secret has incorrect length (${#secret}, expected 64)" >&2
        return 1
    fi
    
    # Check if secret contains only valid hex characters
    if [[ ! "$secret" =~ ^[0-9a-fA-F]+$ ]]; then
        echo "Error: Generated secret contains invalid characters" >&2
        return 1
    fi
    
    return 0
}

# Main execution
main() {
    local output_format="${1:-hex}"
    local bytes=32  # 256 bits
    
    case "$output_format" in
        "hex")
            secret=$(generate_hex_secret "$bytes")
            ;;
        "base64")
            if command -v openssl >/dev/null 2>&1; then
                secret=$(openssl rand -base64 "$bytes")
            else
                echo "Error: base64 format requires openssl" >&2
                exit 1
            fi
            ;;
        *)
            echo "Error: Unsupported format '$output_format'. Use 'hex' or 'base64'." >&2
            exit 1
            ;;
    esac
    
    # Validate the secret (only for hex format)
    if [[ "$output_format" == "hex" ]]; then
        if ! validate_secret "$secret"; then
            exit 1
        fi
    fi
    
    # Output the secret
    echo "$secret"
}

# Show usage if help is requested
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    cat << EOF
Usage: $0 [FORMAT]

Generate a cryptographically secure JWE secret key.

FORMATS:
    hex     Generate 64-character hexadecimal secret (default)
    base64  Generate base64-encoded secret

EXAMPLES:
    $0                    # Generate hex secret
    $0 hex               # Generate hex secret
    $0 base64            # Generate base64 secret
    
    # Use in environment variable
    export JWE_SECRET=\$($0)
    
    # Save to file
    $0 > .env.secret

SECURITY NOTES:
    - Keep the generated secret secure and never commit it to version control
    - Use different secrets for different environments (dev/staging/prod)
    - Store secrets in secure environment variable management systems
    - Rotate secrets periodically for enhanced security

EOF
    exit 0
fi

# Execute main function
main "$@"
