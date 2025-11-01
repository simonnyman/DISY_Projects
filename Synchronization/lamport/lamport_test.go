package lamport

import (
	"sync"
	"testing"
)

func TestNewClock(t *testing.T) {
	clock := NewClock()
	if clock.Time() != 0 {
		t.Errorf("Expected initial time to be 0, got %d", clock.Time())
	}
}

func TestTick(t *testing.T) {
	clock := NewClock()

	time1 := clock.Tick()
	if time1 != 1 {
		t.Errorf("Expected first tick to return 1, got %d", time1)
	}

	time2 := clock.Tick()
	if time2 != 2 {
		t.Errorf("Expected second tick to return 2, got %d", time2)
	}
}

func TestSend(t *testing.T) {
	clock := NewClock()

	timestamp := clock.Send()
	if timestamp != 1 {
		t.Errorf("Expected first send to return 1, got %d", timestamp)
	}

	if clock.Time() != 1 {
		t.Errorf("Expected clock time to be 1 after send, got %d", clock.Time())
	}
}

func TestReceive(t *testing.T) {
	clock := NewClock()

	// Receive a message with timestamp 5
	newTime := clock.Receive(5)
	if newTime != 6 {
		t.Errorf("Expected time after receive to be 6, got %d", newTime)
	}

	// Receive a message with smaller timestamp
	newTime = clock.Receive(3)
	if newTime != 7 {
		t.Errorf("Expected time after receive to be 7, got %d", newTime)
	}
}

func TestReceiveMaxLogic(t *testing.T) {
	clock := NewClock()
	clock.Tick() // time = 1
	clock.Tick() // time = 2
	clock.Tick() // time = 3

	// Receive message with smaller timestamp
	newTime := clock.Receive(1)
	expected := int64(4) // max(3, 1) + 1
	if newTime != expected {
		t.Errorf("Expected %d, got %d", expected, newTime)
	}

	// Receive message with larger timestamp
	newTime = clock.Receive(10)
	expected = 11 // max(4, 10) + 1
	if newTime != expected {
		t.Errorf("Expected %d, got %d", expected, newTime)
	}
}

func TestReset(t *testing.T) {
	clock := NewClock()
	clock.Tick()
	clock.Tick()
	clock.Reset()

	if clock.Time() != 0 {
		t.Errorf("Expected time to be 0 after reset, got %d", clock.Time())
	}
}

func TestConcurrency(t *testing.T) {
	clock := NewClock()
	var wg sync.WaitGroup

	// Simulate concurrent ticks
	numGoroutines := 100
	ticksPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < ticksPerGoroutine; j++ {
				clock.Tick()
			}
		}()
	}

	wg.Wait()

	expectedTime := int64(numGoroutines * ticksPerGoroutine)
	if clock.Time() != expectedTime {
		t.Errorf("Expected time to be %d after concurrent ticks, got %d", expectedTime, clock.Time())
	}
}

func TestMessageHappensBefore(t *testing.T) {
	msg1 := NewMessage(5, 1, "data1")
	msg2 := NewMessage(10, 2, "data2")
	msg3 := NewMessage(10, 1, "data3")

	if !msg1.HappensBefore(msg2) {
		t.Error("Message with timestamp 5 should happen before message with timestamp 10")
	}

	if msg2.HappensBefore(msg1) {
		t.Error("Message with timestamp 10 should not happen before message with timestamp 5")
	}

	// Test tiebreaker with process ID
	if !msg3.HappensBefore(msg2) {
		t.Error("Message with same timestamp but lower process ID should happen before")
	}
}

func TestMessageOrdering(t *testing.T) {
	tests := []struct {
		name     string
		msg1     *Message
		msg2     *Message
		expected bool
	}{
		{
			name:     "Different timestamps",
			msg1:     NewMessage(1, 1, nil),
			msg2:     NewMessage(2, 1, nil),
			expected: true,
		},
		{
			name:     "Same timestamp, different process",
			msg1:     NewMessage(5, 1, nil),
			msg2:     NewMessage(5, 2, nil),
			expected: true,
		},
		{
			name:     "Reverse order",
			msg1:     NewMessage(10, 1, nil),
			msg2:     NewMessage(5, 1, nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg1.HappensBefore(tt.msg2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Benchmark tests
func BenchmarkTick(b *testing.B) {
	clock := NewClock()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clock.Tick()
	}
}

func BenchmarkSend(b *testing.B) {
	clock := NewClock()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clock.Send()
	}
}

func BenchmarkReceive(b *testing.B) {
	clock := NewClock()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clock.Receive(int64(i))
	}
}

func BenchmarkConcurrentOperations(b *testing.B) {
	clock := NewClock()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			clock.Tick()
		}
	})
}
