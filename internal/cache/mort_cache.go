package cache

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/velvetriddles/mortgage-calc/internal/model"
)

var (
	// ErrEmptyCache occurs when attempting to get data from an empty cache
	ErrEmptyCache = errors.New("empty cache")
)

// CachedItem represents a cache element with a unique identifier
type CachedItem struct {
	ID int `json:"id"`
	model.ExecuteResponse
}

// Cache defines an interface for caching mortgage calculation results
type Cache interface {
	Save(resp model.ExecuteResponse) int
	GetAll() ([]CachedItem, error)
	Clear()
	Size() int
}

// MortCache represents a thread-safe implementation of cache for mortgage calculation results
type MortCache struct {
	items  map[int]model.ExecuteResponse
	mu     sync.RWMutex
	nextID int32 // Using atomic for generating unique IDs
}

// NewMortCache creates a new cache instance
func NewMortCache() *MortCache {
	return &MortCache{
		items:  make(map[int]model.ExecuteResponse),
		nextID: -1, // Initialize with -1 so the first ID will be 0
	}
}

// Save saves the calculation result to the cache and returns its ID
func (c *MortCache) Save(resp model.ExecuteResponse) int {
	// Generate unique ID atomically
	id := int(atomic.AddInt32(&c.nextID, 1))

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[id] = resp

	return id
}

// GetAll returns all stored in the cache elements
func (c *MortCache) GetAll() ([]CachedItem, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.items) == 0 {
		return nil, ErrEmptyCache
	}

	result := make([]CachedItem, 0, len(c.items))

	for id, resp := range c.items {
		result = append(result, CachedItem{
			ID:              id,
			ExecuteResponse: resp,
		})
	}

	return result, nil
}

// Clear clears the cache
func (c *MortCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a new map instead of clearing the existing one
	// this is more efficient for GC
	c.items = make(map[int]model.ExecuteResponse)
}

// Size returns the number of elements in the cache
func (c *MortCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}
