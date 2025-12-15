package ports

import (
	"context"

	"auth0-server/internal/domain/account"
	"auth0-server/internal/domain/auth"
)

// TokenService defines the interface for token operations
type TokenService interface {
	// GenerateAccessToken creates a new access token for an account
	GenerateAccessToken(ctx context.Context, account *account.Account, scopes []string) (string, error)

	// GenerateRefreshToken creates a new refresh token for an account
	GenerateRefreshToken(ctx context.Context, account *account.Account) (string, error)

	// ValidateToken validates and parses a token
	ValidateToken(ctx context.Context, token string) (*auth.Claims, error)

	// RefreshToken generates a new access token from a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (string, error)

	// RevokeToken invalidates a token
	RevokeToken(ctx context.Context, token string) error
}

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	// Hash creates a hash of the given password
	Hash(ctx context.Context, password string) (string, error)

	// Verify checks if a password matches the given hash
	Verify(ctx context.Context, password, hash string) bool
}

// IDGenerator defines the interface for generating unique identifiers
type IDGenerator interface {
	// Generate creates a new unique ID
	Generate() string

	// GenerateSecure creates a cryptographically secure ID
	GenerateSecure() (string, error)
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	// Publish sends an event to the event bus
	Publish(ctx context.Context, event interface{}) error

	// Subscribe registers a handler for specific event types
	Subscribe(eventType string, handler func(ctx context.Context, event interface{}) error)

	// Close closes the event publisher
	Close() error
}
