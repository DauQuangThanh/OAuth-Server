package ports

import (
	"context"

	"auth0-server/internal/domain/account"
)

// AccountRepository defines the interface for account persistence operations
type AccountRepository interface {
	// Create stores a new account
	Create(ctx context.Context, account *account.Account) error

	// GetByID retrieves an account by their ID
	GetByID(ctx context.Context, id string) (*account.Account, error)

	// GetByEmail retrieves an account by their email
	GetByEmail(ctx context.Context, email string) (*account.Account, error)

	// Update modifies an existing account
	Update(ctx context.Context, account *account.Account) error

	// Delete removes an account by ID
	Delete(ctx context.Context, id string) error

	// List retrieves accounts with pagination
	List(ctx context.Context, offset, limit int) ([]*account.Account, int, error)

	// Close closes any resources used by the repository
	Close() error
}

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	// Set stores a value with expiration
	Set(ctx context.Context, key string, value interface{}, ttl int64) error

	// Get retrieves a value by key
	Get(ctx context.Context, key string) (interface{}, error)

	// Delete removes a value by key
	Delete(ctx context.Context, key string) error

	// Close closes the cache connection
	Close() error
}
