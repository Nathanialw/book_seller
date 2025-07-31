package cache

import (
	"log"
	"sync"

	"bookmaker.ca/internal/db"
)

var (
	authors []string
	mu      sync.RWMutex
)

// LoadAuthors queries the DB once and caches the result.
func LoadAuthors() error {
	gs, err := db.GetAuthors()
	if err != nil {
		return err
	}
	mu.Lock()
	authors = gs
	mu.Unlock()
	log.Printf("Loaded %d authors into cache", len(gs))
	return nil
}

func GetAuthors() []string {
	mu.RLock()
	defer mu.RUnlock()
	return authors
}
