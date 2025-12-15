package account

import (
	"context"
	"time"
)

// Account represents the account domain entity
type Account struct {
	ID        string    `json:"account_id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never serialize
	Name      string    `json:"name,omitempty"`
	Nickname  string    `json:"nickname,omitempty"`
	Picture   string    `json:"picture,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Verified  bool      `json:"email_verified"`
	Blocked   bool      `json:"blocked"`
}

// Repository defines the interface for account storage operations
type Repository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Account, error)
}

// Service defines the interface for account business logic
type Service interface {
	CreateAccount(ctx context.Context, email, password, name string) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	ValidateCredentials(ctx context.Context, email, password string) (*Account, error)
	UpdateAccount(ctx context.Context, account *Account) error
	ListAccounts(ctx context.Context, limit, offset int) ([]*Account, error)
}

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

// CreateAccountRequest represents an account creation request
type CreateAccountRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name,omitempty"`
}

// AccountProfile represents public account information (compatible with user profile for Auth0)
type AccountProfile struct {
	ID            string `json:"sub"` // Keep 'sub' for Auth0 compatibility
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name,omitempty"`
	Nickname      string `json:"nickname,omitempty"`
	Picture       string `json:"picture,omitempty"`
}

// ToUserProfile converts AccountProfile to maintain backward compatibility
func (ap *AccountProfile) ToUserProfile() *AccountProfile {
	return ap // They have the same structure
}
