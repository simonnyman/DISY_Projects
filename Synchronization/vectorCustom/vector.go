package vector

import (
	"sync"
)

type Vector struct {
	processID int
	clock     []int64
	mutex     sync.RWMutex
}

func NewVector(processID, numProcesses int) *Vector {
	return &Vector{
		processID: processID,
		clock:     make([]int64, numProcesses),
	}
}

func (vector *Vector) Tick() []int64 {
	vector.mutex.Lock()
	defer vector.mutex.Unlock()
	vector.clock[vector.processID]++
	return vector.copyClock()
}

func (vector *Vector) copyClock() []int64 {
	clockCopy := make([]int64, len(vector.clock))
	copy(clockCopy, vector.clock)
	return clockCopy
}

func (vector *Vector) Receive(receivedClock []int64) []int64 {
	vector.mutex.Lock()
	defer vector.mutex.Unlock()

	for i := 0; i < len(vector.clock); i++ {
		if receivedClock[i] > vector.clock[i] {
			vector.clock[i] = receivedClock[i]
		}
	}
	vector.clock[vector.processID]++
	return vector.copyClock()
}

func (vector *Vector) Reset() {
	vector.mutex.Lock()
	defer vector.mutex.Unlock()

	for i := 0; i < len(vector.clock); i++ {
		vector.clock[i] = 0
	}
}

func (vector *Vector) Clock() []int64 {
	vector.mutex.RLock()
	defer vector.mutex.RUnlock()
	return vector.copyClock()
}

func (vector *Vector) Send() []int64 {
	vector.mutex.Lock()
	defer vector.mutex.Unlock()
	vector.clock[vector.processID]++
	return vector.copyClock()
}

type Ordering int

const (
	Before Ordering = iota
	After
	Concurrent
	Equal
)

// CompareClocks determines the causal relationship between two vector clocks
func CompareClocks(vector1, vector2 []int64) Ordering {
	if len(vector1) != len(vector2) {
		panic("vectors must have the same length")
	}

	equal := true
	before := false
	after := false

	for i := 0; i < len(vector1); i++ {
		if vector1[i] < vector2[i] {
			equal = false
			before = true
		} else if vector1[i] > vector2[i] {
			equal = false
			after = true
		}
	}

	if equal {
		return Equal
	} else if before && !after {
		return Before
	} else if after && !before {
		return After
	}
	return Concurrent
}

// type Message struct {
// 	vectorTime []int64
// 	processID  int
// 	Data       interface{}
// }

// func newMessage(mVectorTime []int64, mProcessID int, mData interface{}) *Message {
// 	time := make([]int64, len(mVectorTime))

// 	copy(time, mVectorTime)
// 	return &Message{
// 		vectorTime: time,
// 		processID:  mProcessID,
// 		Data:       mData,
// 	}
// }

// func (m *Message) compareTo(other *Message) Ordering {
// 	return compareClocks(m.vectorTime, other.vectorTime)
// }

// func (m *Message) happenedBefore(other *Message) bool {
// 	return m.compareTo(other) == Before
// }

// func (m *Message) isConcurrent(other *Message) bool {
// 	return m.compareTo(other) == Concurrent
// }
