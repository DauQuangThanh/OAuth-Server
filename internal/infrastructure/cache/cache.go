package cache

import (
	"context"
	"sync"
	"time"

	"auth0-server/internal/application/ports"
)

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// InMemoryCache implements a thread-safe in-memory cache
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	maxSize int
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(maxSize int) *InMemoryCache {
	cache := &InMemoryCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
	}

	// Start background cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set implements ports.CacheRepository
func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If cache is at max size, remove oldest entries
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	expiresAt := time.Now().Add(time.Duration(ttl) * time.Second)
	c.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	return nil
}

// Get implements ports.CacheRepository
func (c *InMemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists || entry.IsExpired() {
		return nil, ErrCacheKeyNotFound
	}

	return entry.Value, nil
}

// Delete implements ports.CacheRepository
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	return nil
}

// Close implements ports.CacheRepository
func (c *InMemoryCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	return nil
}

// cleanup removes expired entries periodically
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if entry.IsExpired() {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// evictOldest removes the oldest entry (simple FIFO for demo)
func (c *InMemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.ExpiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.ExpiresAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// GetStats returns cache statistics
func (c *InMemoryCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expired := 0
	for _, entry := range c.entries {
		if entry.IsExpired() {
			expired++
		}
	}

	return map[string]interface{}{
		"total_entries":   len(c.entries),
		"expired_entries": expired,
		"max_size":        c.maxSize,
		"utilization":     float64(len(c.entries)) / float64(c.maxSize),
	}
}

// CachedTokenService wraps a token service with caching
type CachedTokenService struct {
	tokenService ports.TokenService
	cache        ports.CacheRepository
	cacheTTL     int64
}

// NewCachedTokenService creates a new cached token service
func NewCachedTokenService(tokenService ports.TokenService, cache ports.CacheRepository, cacheTTL int64) *CachedTokenService {
	return &CachedTokenService{
		tokenService: tokenService,
		cache:        cache,
		cacheTTL:     cacheTTL,
	}
}

// ValidateToken implements ports.TokenService with caching
func (c *CachedTokenService) ValidateToken(ctx context.Context, token string) (interface{}, error) {
	// Try to get from cache first
	cacheKey := "token:" + token
	if cached, err := c.cache.Get(ctx, cacheKey); err == nil {
		return cached, nil
	}

	// Not in cache, validate with underlying service
	claims, err := c.tokenService.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Cache the result
	c.cache.Set(ctx, cacheKey, claims, c.cacheTTL)

	return claims, nil
}

// Error definitions
var (
	ErrCacheKeyNotFound = &CacheError{Message: "cache key not found"}
)

// CacheError represents a cache-related error
type CacheError struct {
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}

// Ensure InMemoryCache implements the interface
var _ ports.CacheRepository = (*InMemoryCache)(nil)
