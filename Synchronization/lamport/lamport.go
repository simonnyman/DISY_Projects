package lamport

import (
	"sync"
)

// Lamport's logical clock
// thread-safe for concurrent use.
type LamportClock struct {
	mu   sync.Mutex
	time int64
}

// creates a new Lamport clock initialized to zero.
func NewLamportClock() *LamportClock {
	return &LamportClock{
		time: 0,
	}
}

// tick increments the clock for a local event.
func (lc *LamportClock) Tick() int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time++
	return lc.time
}

// send increments the clock and returns the timestamp for the outgoing message.
func (lc *LamportClock) Send() int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time++
	return lc.time
}

// updates the clock based on received timestamp.
// sets time to max(local, received) + 1.
func (lc *LamportClock) Receive(receivedTime int64) int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time = max(lc.time, receivedTime) + 1
	return lc.time
}

// returns the current clock value.
func (lc *LamportClock) Time() int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.time
}

// sets the clock back to zero.
func (lc *LamportClock) Reset() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time = 0
}
