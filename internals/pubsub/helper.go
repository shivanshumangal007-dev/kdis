package pubsub

import (
	"sync"
)

type PubsubStore struct {
	mu       sync.RWMutex
	channels map[string][]chan string
}

func NewPubsubStore() *PubsubStore {
	store := &PubsubStore{
		channels: make(map[string][]chan string),
	}

	return store
}

// func dataSender(ps *pubsub, channelName string, data string) (int, error) {

// }
