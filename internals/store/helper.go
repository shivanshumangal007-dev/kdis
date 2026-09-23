package store

import (
	"fmt"
	"time"
)

func isExpired(v valueStore) bool {
	return !v.expiresAt.IsZero() && time.Now().After(v.expiresAt)
}

func checktype(v valueStore, wanted ValueType) error{
	if v.kind != wanted{
		return fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return nil
}

func ExpiredKeysRemover(s *InMemoryStore) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	fmt.Println("started ticker for removing expired keys evry minute")
	for range ticker.C {
		s.mu.Lock()
		for key, val := range s.items {
			fmt.Printf("checking for %s \n", key)
			if isExpired(val) {
				delete(s.items, key)
				fmt.Println("expired key found")
			}
		}
		s.mu.Unlock()
	}
}
