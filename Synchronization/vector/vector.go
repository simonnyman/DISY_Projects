package vector

import (
	"sync"
)

// vector clock
// thread-safe for concurrent use.
type Vector struct {
	processID int
	clock     []int64
	mu        sync.RWMutex
}

// creates a new Vector clock for the specified process.
func NewVector(processID, numProcesses int) *Vector {
	return &Vector{
		processID: processID,
		clock:     make([]int64, numProcesses),
	}
}

// increments the clock for a local event.
func (v *Vector) Tick() []int64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.clock[v.processID]++
	return v.copyClock()
}

// increments the clock and returns timestamp for outgoing message.
func (v *Vector) Send() []int64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.clock[v.processID]++
	return v.copyClock()
}

// updates the clock based on received timestamp.
// merges by taking component-wise max, then increments own counter.
func (v *Vector) Receive(receivedClock []int64) []int64 {
	v.mu.Lock()
	defer v.mu.Unlock()

	for i := range v.clock {
		v.clock[i] = max(v.clock[i], receivedClock[i])
	}
	v.clock[v.processID]++
	return v.copyClock()
}

// returns a copy of the current vector clock.
func (v *Vector) Clock() []int64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.copyClock()
}

// sets all components to zero.
func (v *Vector) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()

	for i := range v.clock {
		v.clock[i] = 0
	}
}

// creates a copy of the clock slice.
// must be called with lock held.
func (v *Vector) copyClock() []int64 {
	clockCopy := make([]int64, len(v.clock))
	copy(clockCopy, v.clock)
	return clockCopy
}

// ordering represents the causal relationship between two events.
type Ordering int

const (
	Before     Ordering = iota // v1 happened before v2
	After                      // v1 happened after v2
	Concurrent                 // v1 and v2 are concurrent
	Equal                      // v1 and v2 are identical
)

// string returns readable representation of the ordering.
func (o Ordering) String() string {
	switch o {
	case Before:
		return "Before"
	case After:
		return "After"
	case Concurrent:
		return "Concurrent"
	case Equal:
		return "Equal"
	default:
		return "Unknown"
	}
}

// determines the causal relationship between two vector clocks.
// panics if vectors have different lengths.
func CompareClocks(v1, v2 []int64) Ordering {
	if len(v1) != len(v2) {
		panic("vector: cannot compare clocks of different lengths")
	}

	allEqual := true
	hasLess := false
	hasGreater := false

	for i := range v1 {
		if v1[i] < v2[i] {
			allEqual = false
			hasLess = true
		} else if v1[i] > v2[i] {
			allEqual = false
			hasGreater = true
		}
	}

	if allEqual {
		return Equal
	}
	if hasLess && !hasGreater {
		return Before
	}
	if hasGreater && !hasLess {
		return After
	}
	return Concurrent
}
