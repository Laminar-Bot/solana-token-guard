package tokenguard

import (
	"context"
	"sync"
	"time"
)

// DefaultCacheTTL is the default time-to-live for cached screening results.
const DefaultCacheTTL = 5 * time.Minute

// DefaultMaxCacheSize is the default maximum number of entries in the cache.
// With ~1KB per entry, 10,000 entries = ~10MB max memory.
const DefaultMaxCacheSize = 10000

// ============================================================================
// In-Memory Cache Implementation
// ============================================================================

// cacheEntry holds a cached screening result with expiration.
type cacheEntry struct {
	result    *TokenScreeningResult
	expiresAt time.Time
}

// isExpired returns true if the entry has passed its expiration time.
func (e *cacheEntry) isExpired() bool {
	return time.Now().After(e.expiresAt)
}

// InMemoryCache provides a simple in-memory cache for screening results.
//
// This implementation is suitable for single-instance deployments.
// For distributed deployments, use a custom cache implementation.
//
// Features:
//   - Thread-safe
//   - Configurable TTL
//   - Configurable max size (prevents unbounded memory growth)
//   - Automatic expiration checks on read
//   - Periodic cleanup goroutine (optional)
type InMemoryCache struct {
	entries map[string]*cacheEntry
	ttl     time.Duration
	maxSize int
	mu      sync.RWMutex

	// For cleanup goroutine
	done chan struct{}
}

// InMemoryCacheConfig holds configuration for InMemoryCache.
type InMemoryCacheConfig struct {
	// TTL is the time-to-live for cached entries.
	// Defaults to 5 minutes if zero.
	TTL time.Duration

	// MaxSize is the maximum number of entries in the cache.
	// When the cache is full, expired entries are evicted first,
	// then the oldest entries are evicted to make room.
	// Defaults to 10,000 if zero.
	MaxSize int

	// CleanupInterval is how often expired entries are removed.
	// Set to 0 to disable background cleanup.
	// If enabled, you must call Close() to stop the cleanup goroutine.
	CleanupInterval time.Duration
}

// NewInMemoryCache creates a new in-memory cache.
func NewInMemoryCache(cfg InMemoryCacheConfig) *InMemoryCache {
	if cfg.TTL == 0 {
		cfg.TTL = DefaultCacheTTL
	}
	if cfg.MaxSize == 0 {
		cfg.MaxSize = DefaultMaxCacheSize
	}

	c := &InMemoryCache{
		entries: make(map[string]*cacheEntry),
		ttl:     cfg.TTL,
		maxSize: cfg.MaxSize,
		done:    make(chan struct{}),
	}

	// Start background cleanup if interval is set
	if cfg.CleanupInterval > 0 {
		go c.cleanupLoop(cfg.CleanupInterval)
	}

	return c
}

// Get retrieves a cached screening result.
//
// Returns the result and true if found and not expired.
// Returns nil and false if not found or expired.
func (c *InMemoryCache) Get(_ context.Context, tokenMint string) (*TokenScreeningResult, bool) {
	c.mu.RLock()
	entry, ok := c.entries[tokenMint]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if entry.isExpired() {
		// Lazy deletion on read
		c.mu.Lock()
		delete(c.entries, tokenMint)
		c.mu.Unlock()
		return nil, false
	}

	return entry.result, true
}

// Set stores a screening result in the cache.
//
// If the cache is at maximum capacity, it will:
// 1. First evict all expired entries
// 2. If still at capacity, evict one entry (effectively random due to map iteration)
//
// This prevents unbounded memory growth if cleanup goroutine fails or TTL is extended.
func (c *InMemoryCache) Set(_ context.Context, result *TokenScreeningResult) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if this is an update to an existing entry (no size increase)
	_, isUpdate := c.entries[result.TokenMint]

	// If at max size and this is a new entry, we need to make room
	if !isUpdate && len(c.entries) >= c.maxSize {
		c.evictLocked()
	}

	c.entries[result.TokenMint] = &cacheEntry{
		result:    result,
		expiresAt: time.Now().Add(c.ttl),
	}

	return nil
}

// evictLocked removes entries to make room for new ones.
// Must be called with c.mu held (write lock).
//
// Strategy:
// 1. First, evict all expired entries (they're useless anyway)
// 2. If still at max capacity, evict one entry (random due to map iteration order)
func (c *InMemoryCache) evictLocked() {
	now := time.Now()

	// Phase 1: Evict all expired entries
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}

	// Phase 2: If still at capacity, evict one entry
	// Map iteration in Go is randomized, so this effectively evicts a random entry.
	// This is simpler than maintaining LRU order and sufficient for a screening cache.
	if len(c.entries) >= c.maxSize {
		for key := range c.entries {
			delete(c.entries, key)
			break // Only evict one
		}
	}
}

// Delete removes an entry from the cache.
func (c *InMemoryCache) Delete(_ context.Context, tokenMint string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, tokenMint)
}

// Clear removes all entries from the cache.
func (c *InMemoryCache) Clear(_ context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
}

// Size returns the number of entries in the cache (including expired).
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// Close stops the background cleanup goroutine.
// Must be called if CleanupInterval was set.
func (c *InMemoryCache) Close() error {
	close(c.done)
	return nil
}

// cleanupLoop periodically removes expired entries.
func (c *InMemoryCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.done:
			return
		}
	}
}

// cleanup removes all expired entries.
func (c *InMemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}

// ============================================================================
// No-Op Cache Implementation
// ============================================================================

// NoOpCache is a cache implementation that doesn't cache anything.
// Useful for testing or when caching should be disabled.
type NoOpCache struct{}

// NewNoOpCache creates a new no-op cache.
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns nil and false (cache miss).
func (c *NoOpCache) Get(_ context.Context, _ string) (*TokenScreeningResult, bool) {
	return nil, false
}

// Set does nothing.
func (c *NoOpCache) Set(_ context.Context, _ *TokenScreeningResult) error {
	return nil
}
