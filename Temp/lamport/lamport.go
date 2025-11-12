package lamport

import (
	"sync"
)

// Clock represents a Lamport logical clock
type Clock struct {
	time  int64
	mutex sync.Mutex
}

// NewClock creates a new Lamport clock initialized to 0
func NewClock() *Clock {
	return &Clock{
		time: 0,
	}
}

// Tick increments the logical clock (local event)
// Time complexity: O(1)
// Space complexity: O(1)
func (c *Clock) Tick() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.time++
	return c.time
}

// Send returns the current timestamp to include in a message
// and increments the clock
// Time complexity: O(1)
// Space complexity: O(1)
func (c *Clock) Send() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.time++
	return c.time
}

// Receive updates the clock based on received timestamp
// Implements: c.time = max(c.time, receivedTime) + 1
// Time complexity: O(1)
// Space complexity: O(1)
func (c *Clock) Receive(receivedTime int64) int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if receivedTime > c.time {
		c.time = receivedTime
	}
	c.time++
	return c.time
}

// Time returns the current logical clock value
func (c *Clock) Time() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.time
}

// Reset resets the clock to 0
func (c *Clock) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.time = 0
}

// Message represents a message with Lamport timestamp
type Message struct {
	Timestamp int64
	ProcessID int
	Data      interface{}
}

// NewMessage creates a new message with the given timestamp
func NewMessage(timestamp int64, processID int, data interface{}) *Message {
	return &Message{
		Timestamp: timestamp,
		ProcessID: processID,
		Data:      data,
	}
}

// HappensBefore determines if this message happened before another
// Returns true if this message's timestamp is less than the other's
// For equal timestamps, uses process ID as tiebreaker
func (m *Message) HappensBefore(other *Message) bool {
	if m.Timestamp < other.Timestamp {
		return true
	}
	if m.Timestamp == other.Timestamp {
		return m.ProcessID < other.ProcessID
	}
	return false
}
