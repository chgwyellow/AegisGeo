package store

import (
	"AegisGeo/internal/models"
	"sync"
)

// Define Mutex object, like blueprint
type MemoryCache struct {
	mu     sync.RWMutex            // Read & Write Mutex
	events map[string]models.Event // Event ID is key, struct is value
}

// Initialize factory, like building a tower via blueprint, the constructor
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		events: make(map[string]models.Event),
	}
}

// Set function, write or update (WLock)
func (c *MemoryCache) Set(event models.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.events[event.ID] = event
}

// Get function, get single event (RLock)
func (c *MemoryCache) Get(id string) (models.Event, bool) {
	c.mu.RLock() // Many people can read the same date but prohibiting write
	defer c.mu.RUnlock()

	event, exists := c.events[id]
	return event, exists
}

// Get function, get all events (RLock)
func (c *MemoryCache) GetAll() []models.Event {
	c.mu.RLock()
	defer c.mu.RUnlock()

	list := make([]models.Event, 0, len(c.events))
	for _, event := range c.events {
		list = append(list, event)
	}
	return list
}
