package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"auth0-server/internal/domain/account"
	"auth0-server/pkg/logger"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresAccountRepository implements account repository using PostgreSQL
type PostgresAccountRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresAccountRepository creates a new PostgreSQL account repository
func NewPostgresAccountRepository(db *sql.DB, logger logger.Logger) *PostgresAccountRepository {
	return &PostgresAccountRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new account into the database
func (r *PostgresAccountRepository) Create(ctx context.Context, a *account.Account) error {
	query := `
		INSERT INTO accounts (id, email, password, name, nickname, picture, created_at, updated_at, verified, blocked)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	// Debug logging
	r.logger.Info("Executing Create query", map[string]interface{}{
		"query":      query,
		"account_id": a.ID,
		"email":      a.Email,
	})

	_, err := r.db.ExecContext(ctx, query,
		a.ID, a.Email, a.Password, a.Name, a.Nickname, a.Picture,
		a.CreatedAt, a.UpdatedAt, a.Verified, a.Blocked,
	)

	if err != nil {
		r.logger.Error("Failed to create account", err, map[string]interface{}{
			"component":  "postgres_account_repository",
			"account_id": a.ID,
			"email":      a.Email,
		})
		return fmt.Errorf("failed to create account: %w", err)
	}

	r.logger.Info("Account created successfully", map[string]interface{}{
		"component":  "postgres_account_repository",
		"account_id": a.ID,
		"email":      a.Email,
	})

	return nil
}

// GetByID retrieves an account by their ID
func (r *PostgresAccountRepository) GetByID(ctx context.Context, id string) (*account.Account, error) {
	query := `
		SELECT id, email, password, name, nickname, picture, created_at, updated_at, verified, blocked
		FROM accounts WHERE id = $1
	`

	a := &account.Account{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.Email, &a.Password, &a.Name, &a.Nickname, &a.Picture,
		&a.CreatedAt, &a.UpdatedAt, &a.Verified, &a.Blocked,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}

	if err != nil {
		r.logger.Error("Failed to get account by ID", err, map[string]interface{}{
			"component":  "postgres_account_repository",
			"account_id": id,
		})
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}

	return a, nil
}

// GetByEmail retrieves an account by their email address
func (r *PostgresAccountRepository) GetByEmail(ctx context.Context, email string) (*account.Account, error) {
	query := `
		SELECT id, email, password, name, nickname, picture, created_at, updated_at, verified, blocked
		FROM accounts WHERE email = $1
	`

	// Debug logging
	r.logger.Info("Executing GetByEmail query", map[string]interface{}{
		"query": query,
		"email": email,
	})

	a := &account.Account{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&a.ID, &a.Email, &a.Password, &a.Name, &a.Nickname, &a.Picture,
		&a.CreatedAt, &a.UpdatedAt, &a.Verified, &a.Blocked,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}

	if err != nil {
		r.logger.Error("Failed to get account by email", err, map[string]interface{}{
			"component": "postgres_account_repository",
			"email":     email,
		})
		return nil, fmt.Errorf("failed to get account by email: %w", err)
	}

	return a, nil
}

// Update updates an existing account in the database
func (r *PostgresAccountRepository) Update(ctx context.Context, a *account.Account) error {
	query := `
		UPDATE accounts 
		SET email = $2, password = $3, name = $4, nickname = $5, picture = $6,
		    updated_at = $7, verified = $8, blocked = $9
		WHERE id = $1
	`

	a.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		a.ID, a.Email, a.Password, a.Name, a.Nickname, a.Picture,
		a.UpdatedAt, a.Verified, a.Blocked,
	)

	if err != nil {
		r.logger.Error("Failed to update account", err, map[string]interface{}{
			"component":  "postgres_account_repository",
			"account_id": a.ID,
		})
		return fmt.Errorf("failed to update account: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	r.logger.Info("Account updated successfully", map[string]interface{}{
		"component":  "postgres_account_repository",
		"account_id": a.ID,
	})

	return nil
}

// Delete removes an account from the database
func (r *PostgresAccountRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM accounts WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete account", err, map[string]interface{}{
			"component":  "postgres_account_repository",
			"account_id": id,
		})
		return fmt.Errorf("failed to delete account: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	r.logger.Info("Account deleted successfully", map[string]interface{}{
		"component":  "postgres_account_repository",
		"account_id": id,
	})

	return nil
}

// List retrieves accounts with pagination
func (r *PostgresAccountRepository) List(ctx context.Context, limit, offset int) ([]*account.Account, error) {
	query := `
		SELECT id, email, password, name, nickname, picture, created_at, updated_at, verified, blocked
		FROM accounts 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to list accounts", err, map[string]interface{}{
			"component": "postgres_account_repository",
			"limit":     limit,
			"offset":    offset,
		})
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*account.Account
	for rows.Next() {
		a := &account.Account{}
		err := rows.Scan(
			&a.ID, &a.Email, &a.Password, &a.Name, &a.Nickname, &a.Picture,
			&a.CreatedAt, &a.UpdatedAt, &a.Verified, &a.Blocked,
		)
		if err != nil {
			r.logger.Error("Failed to scan account row", err, map[string]interface{}{
				"component": "postgres_account_repository",
			})
			return nil, fmt.Errorf("failed to scan account row: %w", err)
		}
		accounts = append(accounts, a)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error iterating account rows", err, map[string]interface{}{
			"component": "postgres_account_repository",
		})
		return nil, fmt.Errorf("error iterating account rows: %w", err)
	}

	r.logger.Info("Listed accounts successfully", map[string]interface{}{
		"component": "postgres_account_repository",
		"count":     len(accounts),
		"limit":     limit,
		"offset":    offset,
	})

	return accounts, nil
}
