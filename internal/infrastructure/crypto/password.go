package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"auth0-server/internal/domain/account"
)

// BcryptPasswordHasher implements password hashing using bcrypt
// optimized for concurrent operations
type BcryptPasswordHasher struct {
	cost int
	pool sync.Pool
}

// NewBcryptPasswordHasher creates a new bcrypt password hasher
func NewBcryptPasswordHasher(cost int) *BcryptPasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &BcryptPasswordHasher{
		cost: cost,
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 72) // bcrypt max password length
			},
		},
	}
}

// Hash generates a bcrypt hash of the password
func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Use pool to reduce allocations
	buf := h.pool.Get().([]byte)
	defer h.pool.Put(buf)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// Compare verifies a password against its hash
func (h *BcryptPasswordHasher) Compare(hashedPassword, password string) error {
	if len(password) == 0 || len(hashedPassword) == 0 {
		return fmt.Errorf("password and hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	return nil
}

// DefaultPasswordHasher returns a password hasher with default settings
func DefaultPasswordHasher() account.PasswordHasher {
	return NewBcryptPasswordHasher(bcrypt.DefaultCost)
}

// IDGenerator provides thread-safe ID generation
type IDGenerator struct {
	mutex sync.Mutex
}

// NewIDGenerator creates a new ID generator
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// Generate creates a new random hex ID
func (g *IDGenerator) Generate() (string, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
