package pubsub

import "slices"

func (ps *PubsubStore) NewSubsciber(channelName string) <-chan string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	c, found := ps.channels[channelName]
	subChannel := make(chan string)
	if !found { // channel not exist already
		c = make([]chan string, 0)
	}
	// if channel exist
	c = append(c, subChannel)
	return subChannel
}

func (ps *PubsubStore) QuitSubsciber(subChannel chan string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	defer close(subChannel)
	for key, c := range ps.channels {
		idx := slices.Index(c, subChannel)
		if idx != -1 {
			// 2. Swap target with the last element, then truncate the slice
			c[idx] = c[len(c)-1]
			c = c[:len(c)-1]

			if len(c) == 0 {
				delete(ps.channels, key) // Safe to do inside a range loop in Go
			} else {
				ps.channels[key] = c // CRUCIAL: Write the shortened slice back to the map
			}
		}
	}
}
