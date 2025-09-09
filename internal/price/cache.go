package price

import (
	"sync"
	"time"
)

// CacheEntry represents a cached price entry with expiration
type CacheEntry struct {
	// Value is the cached price data
	Value *SOLUSDPrice

	// ExpiresAt indicates when this entry expires
	ExpiresAt time.Time

	// CreatedAt indicates when this entry was created
	CreatedAt time.Time
}

// PriceCache provides thread-safe caching for price data with TTL support
type PriceCache struct {
	// entries holds the cached data
	entries map[string]*CacheEntry

	// ttl is the time-to-live for cache entries
	ttl time.Duration

	// mu protects concurrent access to the cache
	mu sync.RWMutex

	// lastCleanup tracks when we last cleaned up expired entries
	lastCleanup time.Time

	// cleanupInterval determines how often to clean up expired entries
	cleanupInterval time.Duration
}

// NewPriceCache creates a new price cache with the specified TTL
func NewPriceCache(ttl time.Duration) *PriceCache {
	cache := &PriceCache{
		entries:         make(map[string]*CacheEntry),
		ttl:             ttl,
		cleanupInterval: ttl / 2, // Clean up twice as often as TTL
	}

	// Start background cleanup goroutine if TTL > 0
	if ttl > 0 {
		go cache.cleanupLoop()
	}

	return cache
}

// Get retrieves a price from cache if it exists and hasn't expired
func (c *PriceCache) Get(key string) (*SOLUSDPrice, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// Set stores a price in cache with TTL expiration
func (c *PriceCache) Set(key string, price *SOLUSDPrice) {
	if price == nil {
		return
	}

	now := time.Now()
	entry := &CacheEntry{
		Value:     price,
		ExpiresAt: now.Add(c.ttl),
		CreatedAt: now,
	}

	c.mu.Lock()
	c.entries[key] = entry
	c.mu.Unlock()
}

// GetStale retrieves a price from cache even if it has expired
// Returns the price and a boolean indicating if it's stale (expired)
func (c *PriceCache) GetStale(key string) (*SOLUSDPrice, bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false, false
	}

	isStale := time.Now().After(entry.ExpiresAt)
	return entry.Value, true, isStale
}

// Delete removes a price from cache
func (c *PriceCache) Delete(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Clear removes all entries from cache
func (c *PriceCache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]*CacheEntry)
	c.mu.Unlock()
}

// Size returns the number of entries currently in cache
func (c *PriceCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// GetTTL returns the cache TTL duration
func (c *PriceCache) GetTTL() time.Duration {
	return c.ttl
}

// GetStats returns cache statistics
func (c *PriceCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	stats := CacheStats{
		TotalEntries: len(c.entries),
	}

	for _, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			stats.ExpiredEntries++
		} else {
			stats.ValidEntries++
		}

		age := now.Sub(entry.CreatedAt)
		if stats.OldestEntry == 0 || age > stats.OldestEntry {
			stats.OldestEntry = age
		}
		if stats.NewestEntry == 0 || age < stats.NewestEntry {
			stats.NewestEntry = age
		}
	}

	return stats
}

// CacheStats provides statistics about cache usage
type CacheStats struct {
	TotalEntries   int           `json:"total_entries"`
	ValidEntries   int           `json:"valid_entries"`
	ExpiredEntries int           `json:"expired_entries"`
	OldestEntry    time.Duration `json:"oldest_entry"`
	NewestEntry    time.Duration `json:"newest_entry"`
}

// cleanupLoop runs in background to remove expired entries
func (c *PriceCache) cleanupLoop() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired entries from cache
func (c *PriceCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Track cleanup time to avoid excessive cleanup calls
	if now.Sub(c.lastCleanup) < c.cleanupInterval/2 {
		return // Skip if we cleaned up recently
	}

	// Remove expired entries
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}

	c.lastCleanup = now
}

// Standard cache key constants
const (
	SOLUSDCacheKey = "sol_usd_price"
)

// PriceCacheManager manages multiple price caches with different TTLs
type PriceCacheManager struct {
	// solCache caches SOL/USD prices
	solCache *PriceCache

	// config holds cache configuration
	config *PriceConfig
}

// NewPriceCacheManager creates a new price cache manager
func NewPriceCacheManager(config *PriceConfig) *PriceCacheManager {
	return &PriceCacheManager{
		solCache: NewPriceCache(config.CacheTTL),
		config:   config,
	}
}

// GetSOLPrice retrieves SOL/USD price from cache
func (m *PriceCacheManager) GetSOLPrice() (*SOLUSDPrice, bool) {
	return m.solCache.Get(SOLUSDCacheKey)
}

// SetSOLPrice stores SOL/USD price in cache
func (m *PriceCacheManager) SetSOLPrice(price *SOLUSDPrice) {
	m.solCache.Set(SOLUSDCacheKey, price)
}

// GetSOLPriceStale retrieves SOL/USD price from cache even if expired
func (m *PriceCacheManager) GetSOLPriceStale() (*SOLUSDPrice, bool, bool) {
	return m.solCache.GetStale(SOLUSDCacheKey)
}

// ClearSOLPrice removes SOL/USD price from cache
func (m *PriceCacheManager) ClearSOLPrice() {
	m.solCache.Delete(SOLUSDCacheKey)
}

// ClearAll clears all caches
func (m *PriceCacheManager) ClearAll() {
	m.solCache.Clear()
}

// GetSOLCacheStats returns statistics for SOL price cache
func (m *PriceCacheManager) GetSOLCacheStats() CacheStats {
	return m.solCache.GetStats()
}

// IsSOLPriceStale checks if cached SOL price is stale according to config
func (m *PriceCacheManager) IsSOLPriceStale() bool {
	price, exists, isExpired := m.GetSOLPriceStale()
	if !exists {
		return true // No cache = stale
	}

	if isExpired {
		return true // Cache expired = stale
	}

	// Check against max staleness from config
	maxStale := m.config.GetMaxStaleness()
	if time.Since(price.Timestamp) > maxStale {
		return true // Data too old = stale
	}

	return false
}

// Close stops background cleanup goroutines
func (m *PriceCacheManager) Close() error {
	// Note: Currently cleanup goroutines don't have a clean shutdown mechanism
	// This could be enhanced with context cancellation if needed
	return nil
}
