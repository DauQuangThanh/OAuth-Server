package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"auth0-server/internal/domain/account"
	"auth0-server/pkg/logger"
)

// InMemoryAccountRepository implements account repository using in-memory storage
type InMemoryAccountRepository struct {
	accounts map[string]*account.Account
	mutex    sync.RWMutex
	logger   logger.Logger
}

// NewInMemoryAccountRepository creates a new in-memory account repository
func NewInMemoryAccountRepository(logger logger.Logger) *InMemoryAccountRepository {
	return &InMemoryAccountRepository{
		accounts: make(map[string]*account.Account),
		mutex:    sync.RWMutex{},
		logger:   logger,
	}
}

// Create stores a new account in memory
func (r *InMemoryAccountRepository) Create(ctx context.Context, acc *account.Account) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if account already exists by ID
	if _, exists := r.accounts[acc.ID]; exists {
		return fmt.Errorf("account with ID %s already exists", acc.ID)
	}

	// Check if account already exists by email
	for _, existing := range r.accounts {
		if existing.Email == acc.Email {
			return fmt.Errorf("account with email %s already exists", acc.Email)
		}
	}

	// Store account
	r.accounts[acc.ID] = &account.Account{
		ID:        acc.ID,
		Email:     acc.Email,
		Password:  acc.Password,
		Name:      acc.Name,
		Nickname:  acc.Nickname,
		Picture:   acc.Picture,
		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
		Verified:  acc.Verified,
		Blocked:   acc.Blocked,
	}

	r.logger.Info("Account created successfully", map[string]interface{}{
		"component":  "in_memory_account_repository",
		"account_id": acc.ID,
		"email":      acc.Email,
	})

	return nil
}

// GetByID retrieves an account by their ID
func (r *InMemoryAccountRepository) GetByID(ctx context.Context, id string) (*account.Account, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	acc, exists := r.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account not found")
	}

	// Return a copy to prevent external modification
	return &account.Account{
		ID:        acc.ID,
		Email:     acc.Email,
		Password:  acc.Password,
		Name:      acc.Name,
		Nickname:  acc.Nickname,
		Picture:   acc.Picture,
		CreatedAt: acc.CreatedAt,
		UpdatedAt: acc.UpdatedAt,
		Verified:  acc.Verified,
		Blocked:   acc.Blocked,
	}, nil
}

// GetByEmail retrieves an account by their email address
func (r *InMemoryAccountRepository) GetByEmail(ctx context.Context, email string) (*account.Account, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, acc := range r.accounts {
		if acc.Email == email {
			// Return a copy to prevent external modification
			return &account.Account{
				ID:        acc.ID,
				Email:     acc.Email,
				Password:  acc.Password,
				Name:      acc.Name,
				Nickname:  acc.Nickname,
				Picture:   acc.Picture,
				CreatedAt: acc.CreatedAt,
				UpdatedAt: acc.UpdatedAt,
				Verified:  acc.Verified,
				Blocked:   acc.Blocked,
			}, nil
		}
	}

	return nil, fmt.Errorf("account not found")
}

// Update modifies an existing account in memory
func (r *InMemoryAccountRepository) Update(ctx context.Context, acc *account.Account) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	existing, exists := r.accounts[acc.ID]
	if !exists {
		return fmt.Errorf("account not found")
	}

	// Update the account
	existing.Email = acc.Email
	existing.Password = acc.Password
	existing.Name = acc.Name
	existing.Nickname = acc.Nickname
	existing.Picture = acc.Picture
	existing.UpdatedAt = time.Now()
	existing.Verified = acc.Verified
	existing.Blocked = acc.Blocked

	r.logger.Info("Account updated successfully", map[string]interface{}{
		"component":  "in_memory_account_repository",
		"account_id": acc.ID,
	})

	return nil
}

// Delete removes an account by ID
func (r *InMemoryAccountRepository) Delete(ctx context.Context, id string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.accounts[id]; !exists {
		return fmt.Errorf("account not found")
	}

	delete(r.accounts, id)

	r.logger.Info("Account deleted successfully", map[string]interface{}{
		"component":  "in_memory_account_repository",
		"account_id": id,
	})

	return nil
}

// List retrieves accounts with pagination
func (r *InMemoryAccountRepository) List(ctx context.Context, limit, offset int) ([]*account.Account, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Convert map to slice for consistent ordering
	var accounts []*account.Account
	for _, acc := range r.accounts {
		accounts = append(accounts, &account.Account{
			ID:        acc.ID,
			Email:     acc.Email,
			Password:  acc.Password,
			Name:      acc.Name,
			Nickname:  acc.Nickname,
			Picture:   acc.Picture,
			CreatedAt: acc.CreatedAt,
			UpdatedAt: acc.UpdatedAt,
			Verified:  acc.Verified,
			Blocked:   acc.Blocked,
		})
	}

	// Apply pagination
	start := offset
	if start > len(accounts) {
		start = len(accounts)
	}

	end := start + limit
	if end > len(accounts) {
		end = len(accounts)
	}

	result := accounts[start:end]

	r.logger.Info("Listed accounts successfully", map[string]interface{}{
		"component": "in_memory_account_repository",
		"count":     len(result),
		"limit":     limit,
		"offset":    offset,
	})

	return result, nil
}
