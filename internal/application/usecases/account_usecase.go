package usecases

import (
	"context"
	"fmt"
	"time"

	"auth0-server/internal/domain/account"
	"auth0-server/internal/infrastructure/crypto"
)

// AccountUseCase handles account-related business logic
type AccountUseCase struct {
	accountRepo    account.Repository
	passwordHasher account.PasswordHasher
	idGenerator    *crypto.IDGenerator
}

// NewAccountUseCase creates a new account use case
func NewAccountUseCase(
	accountRepo account.Repository,
	passwordHasher account.PasswordHasher,
	idGenerator *crypto.IDGenerator,
) *AccountUseCase {
	return &AccountUseCase{
		accountRepo:    accountRepo,
		passwordHasher: passwordHasher,
		idGenerator:    idGenerator,
	}
}

// CreateAccount creates a new account with validation
func (uc *AccountUseCase) CreateAccount(ctx context.Context, email, password, name string) (*account.Account, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Validate input
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	// Check if account already exists by email
	existingAccount, err := uc.accountRepo.GetByEmail(ctx, email)
	if err == nil && existingAccount != nil {
		return nil, fmt.Errorf("account with email %s already exists", email)
	}

	// Generate account ID
	accountID, err := uc.idGenerator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate account ID: %w", err)
	}

	// Hash password
	hashedPassword, err := uc.passwordHasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create account
	newAccount := &account.Account{
		ID:        accountID,
		Email:     email,
		Password:  hashedPassword,
		Name:      name,
		Nickname:  name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Verified:  true, // Auto-verify for local development
		Blocked:   false,
	}

	// Save account
	err = uc.accountRepo.Create(ctx, newAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return newAccount, nil
}

// GetAccount retrieves an account by ID
func (uc *AccountUseCase) GetAccount(ctx context.Context, id string) (*account.Account, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if id == "" {
		return nil, fmt.Errorf("account ID is required")
	}

	return uc.accountRepo.GetByID(ctx, id)
}

// GetAccountByEmail retrieves an account by email
func (uc *AccountUseCase) GetAccountByEmail(ctx context.Context, email string) (*account.Account, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	return uc.accountRepo.GetByEmail(ctx, email)
}

// ValidateCredentials validates account credentials for authentication
func (uc *AccountUseCase) ValidateCredentials(ctx context.Context, email, password string) (*account.Account, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	// Get account by email
	acc, err := uc.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is blocked
	if acc.Blocked {
		return nil, fmt.Errorf("account is blocked")
	}

	// Verify password
	err = uc.passwordHasher.Compare(acc.Password, password)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return acc, nil
}

// UpdateAccount updates an existing account
func (uc *AccountUseCase) UpdateAccount(ctx context.Context, acc *account.Account) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if acc == nil {
		return fmt.Errorf("account is required")
	}
	if acc.ID == "" {
		return fmt.Errorf("account ID is required")
	}

	// Update timestamp
	acc.UpdatedAt = time.Now()

	return uc.accountRepo.Update(ctx, acc)
}

// ListAccounts retrieves accounts with pagination
func (uc *AccountUseCase) ListAccounts(ctx context.Context, limit, offset int) ([]*account.Account, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}
	if offset < 0 {
		offset = 0
	}

	return uc.accountRepo.List(ctx, limit, offset)
}

// VerifyPassword verifies if the provided password matches the hashed password
func (uc *AccountUseCase) VerifyPassword(hashedPassword, password string) bool {
	return uc.passwordHasher.Compare(hashedPassword, password) == nil
}
