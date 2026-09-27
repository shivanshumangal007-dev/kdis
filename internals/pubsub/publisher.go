package pubsub

import "fmt"

func (ps *PubsubStore) Publish(channelName, data string) (int, error) {
	ps.mu.RLock()
	if len(data) < 1 {
		return 0, fmt.Errorf("LENGTH length of the sent data should be greater than 0")
	}
	c, found := ps.channels[channelName]
	ps.mu.RUnlock()
	if !found {
		return 0, nil
	}
	count := 0
	for _, cha := range c {
		select {
		case cha <- data:
			count++
			// Data was sent successfully
		default:
			// fmt.Print("error sending to channel: ", cha)
			// Channel was full or not ready; handle the failure here without blocking
		}
	}
	return count, nil
}
