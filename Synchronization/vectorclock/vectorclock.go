package vectorclock

import (
	"sync"
)

// VectorClock represents a vector clock for distributed systems
type VectorClock struct {
	processID int
	clock     []int64
	mutex     sync.RWMutex
}

// NewVectorClock creates a new vector clock for a process
// processID: the ID of this process (0-indexed)
// numProcesses: total number of processes in the system
func NewVectorClock(processID, numProcesses int) *VectorClock {
	return &VectorClock{
		processID: processID,
		clock:     make([]int64, numProcesses),
	}
}

// Tick increments the local process's clock entry (local event)
// Time complexity: O(1)
// Space complexity: O(1)
func (vc *VectorClock) Tick() []int64 {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()
	vc.clock[vc.processID]++
	return vc.copy()
}

// Send returns a copy of the current vector clock to include in a message
// and increments the local process's clock entry
// Time complexity: O(n) where n is number of processes
// Space complexity: O(n)
func (vc *VectorClock) Send() []int64 {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()
	vc.clock[vc.processID]++
	return vc.copy()
}

// Receive updates the vector clock based on received vector
// Implements: vc[i] = max(vc[i], received[i]) for all i, then vc[processID]++
// Time complexity: O(n) where n is number of processes
// Space complexity: O(1) (not counting the return value)
func (vc *VectorClock) Receive(received []int64) []int64 {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()

	for i := 0; i < len(vc.clock) && i < len(received); i++ {
		if received[i] > vc.clock[i] {
			vc.clock[i] = received[i]
		}
	}
	vc.clock[vc.processID]++
	return vc.copy()
}

// Clock returns a copy of the current vector clock
func (vc *VectorClock) Clock() []int64 {
	vc.mutex.RLock()
	defer vc.mutex.RUnlock()
	return vc.copy()
}

// Reset resets all entries in the vector clock to 0
func (vc *VectorClock) Reset() {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()
	for i := range vc.clock {
		vc.clock[i] = 0
	}
}

// copy creates a copy of the clock (must be called with lock held)
func (vc *VectorClock) copy() []int64 {
	result := make([]int64, len(vc.clock))
	copy(result, vc.clock)
	return result
}

// Message represents a message with vector clock timestamp
type Message struct {
	VectorTime []int64
	ProcessID  int
	Data       interface{}
}

// NewMessage creates a new message with the given vector timestamp
func NewMessage(vectorTime []int64, processID int, data interface{}) *Message {
	// Create a copy to avoid aliasing issues
	timeCopy := make([]int64, len(vectorTime))
	copy(timeCopy, vectorTime)
	return &Message{
		VectorTime: timeCopy,
		ProcessID:  processID,
		Data:       data,
	}
}

// Ordering represents the relationship between two vector clocks
type Ordering int

const (
	Before     Ordering = iota // This happened before other
	After                      // This happened after other
	Concurrent                 // Events are concurrent
	Equal                      // Clocks are equal
)

// CompareTo determines the ordering relationship with another message
// Time complexity: O(n) where n is number of processes
func (m *Message) CompareTo(other *Message) Ordering {
	return CompareVectorClocks(m.VectorTime, other.VectorTime)
}

// CompareVectorClocks compares two vector clocks
// Returns: Before, After, Concurrent, or Equal
// Time complexity: O(n) where n is number of processes
func CompareVectorClocks(v1, v2 []int64) Ordering {
	if len(v1) != len(v2) {
		panic("vector clocks must have same length")
	}

	allEqual := true
	someLess := false
	someGreater := false

	for i := 0; i < len(v1); i++ {
		if v1[i] < v2[i] {
			someLess = true
			allEqual = false
		} else if v1[i] > v2[i] {
			someGreater = true
			allEqual = false
		}
	}

	if allEqual {
		return Equal
	}
	if someLess && !someGreater {
		return Before
	}
	if someGreater && !someLess {
		return After
	}
	return Concurrent
}

// HappensBefore determines if this message happened before another
// Returns true only if this definitely happened before (not concurrent)
func (m *Message) HappensBefore(other *Message) bool {
	return m.CompareTo(other) == Before
}

// IsConcurrent determines if this message is concurrent with another
func (m *Message) IsConcurrent(other *Message) bool {
	return m.CompareTo(other) == Concurrent
}
