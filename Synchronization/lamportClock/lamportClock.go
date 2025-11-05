package lamportClock

import (
	"sync"
)

type LamportClock struct {
	mu   sync.Mutex
	time int64
}

func NewLamportClock() *LamportClock {
	return &LamportClock{
		time: 0,
	}
}

func (lc *LamportClock) Tick() int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time++
	return lc.time
}

func (lc *LamportClock) Send() int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time++
	return lc.time
}

func (lc *LamportClock) Receive(receivedTime int64) int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time = max(lc.time, receivedTime) + 1
	return lc.time
}

func (lc *LamportClock) Time() int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.time
}

func (lc *LamportClock) Reset() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time = 0
}

type message struct {
	timestamp int64
	Data      interface{}
	PID       int
}

func CreateMessage(timestamp int64, data interface{}, pid int) *message {
	return &message{
		timestamp: timestamp,
		Data:      data,
		PID:       pid,
	}
}

func (a *message) HappensBefore(b *message) bool {
	return a.timestamp < b.timestamp
}
