package quota

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type cachedQuota struct {
	data      Usage
	expiresAt time.Time
}

// Cache provides a TTL-based in-memory cache for per-user quota Usage
// responses. Concurrent-safe via a read-write mutex.
type Cache struct {
	mu    sync.RWMutex
	items map[uuid.UUID]*cachedQuota
	ttl   time.Duration
}

// NewCache creates a Cache that evicts entries after ttl.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[uuid.UUID]*cachedQuota),
		ttl:   ttl,
	}
}

// Get returns the cached Usage for userID if present and not expired.
func (c *Cache) Get(userID uuid.UUID) (Usage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.items[userID]
	if !ok || time.Now().After(entry.expiresAt) {
		return Usage{}, false
	}
	return entry.data, true
}

// Set stores a Usage snapshot for userID with the configured TTL.
func (c *Cache) Set(userID uuid.UUID, data Usage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[userID] = &cachedQuota{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a single user's cached entry (e.g. after a clone
// changes disk usage).
func (c *Cache) Invalidate(userID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, userID)
}
