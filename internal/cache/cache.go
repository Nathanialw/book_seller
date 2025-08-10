package cache

import (
	"log"
	"sync"

	"github.com/nathanialw/ecommerce/internal/db"
)

var (
	authors []string
	mu      sync.RWMutex
)

// LoadAuthors queries the DB once and caches the result.
func LoadCache() error {
	gs, err := db.GetCache()
	if err != nil {
		return err
	}
	mu.Lock()
	authors = gs
	mu.Unlock()
	log.Printf("Loaded %d authors into cache", len(gs))
	return nil
}

func GetCache() []string {
	mu.RLock()
	defer mu.RUnlock()
	return authors
}

func UpdateCache() {
	if err := updateCache(); err != nil {
		log.Printf("Failed to update authors cache: %v", err)
	}
}

func updateCache() error {
	newAuthors, err := db.GetCache()
	if err != nil {
		return err
	}
	mu.Lock()
	authors = newAuthors
	mu.Unlock()
	log.Printf("Updated authors cache with %d entries", len(newAuthors))
	return nil
}
