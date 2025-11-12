package lamportClock

import (
	"testing"
)

// Basic functionality tests
func TestNewLamportClock(t *testing.T) {
	clock := NewLamportClock()

	if clock.Time() != 0 {
		t.Errorf("Expected initial time 0, got %d", clock.Time())
	}
}

func TestTick(t *testing.T) {
	lc := NewLamportClock()

	lc.Tick()
	if lc.Time() != 1 {
		t.Errorf("Expected time 1, got %d", lc.Time())
	}

	lc.Tick()
	if lc.Time() != 2 {
		t.Errorf("Expected time 2, got %d", lc.Time())
	}
}

func TestSend(t *testing.T) {
	lc := NewLamportClock()

	timestamp := lc.Send()
	if timestamp != 1 {
		t.Errorf("Expected timestamp 1, got %d", timestamp)
	}

	if lc.Time() != 1 {
		t.Errorf("Expected time 1, got %d", lc.Time())
	}
}

func TestReceive(t *testing.T) {
	lc := NewLamportClock()
	lc.Tick() // time = 1

	// Receive message with higher timestamp
	lc.Receive(5)

	// Should be max(1, 5) + 1 = 6
	if lc.Time() != 6 {
		t.Errorf("Expected time 6, got %d", lc.Time())
	}
}

func TestReceiveLowerTimestamp(t *testing.T) {
	lc := NewLamportClock()
	lc.Tick() // 1
	lc.Tick() // 2
	lc.Tick() // 3

	// Receive message with lower timestamp
	lc.Receive(1)

	// Should be max(3, 1) + 1 = 4
	if lc.Time() != 4 {
		t.Errorf("Expected time 4, got %d", lc.Time())
	}
}

func TestReset(t *testing.T) {
	lc := NewLamportClock()
	lc.Tick()
	lc.Tick()

	lc.Reset()
	if lc.Time() != 0 {
		t.Errorf("Expected time 0 after reset, got %d", lc.Time())
	}
}

// Scenario-based tests (from your existing file)
func TestMessagePassing(t *testing.T) {
	processA := NewLamportClock()
	processB := NewLamportClock()

	// Process A does local work
	processA.Tick()
	if processA.Time() != 1 {
		t.Errorf("Process A: Expected time 1, got %d", processA.Time())
	}

	// Process A sends message
	msgTimestamp := processA.Send()
	if msgTimestamp != 2 {
		t.Errorf("Message timestamp should be 2, got %d", msgTimestamp)
	}

	// Process B receives the message
	processB.Receive(msgTimestamp)
	if processB.Time() != 3 {
		t.Errorf("Process B: Expected time 3, got %d", processB.Time())
	}

	// Process B sends reply
	replyTimestamp := processB.Send()
	if replyTimestamp != 4 {
		t.Errorf("Reply timestamp should be 4, got %d", replyTimestamp)
	}

	// Process A receives reply
	processA.Receive(replyTimestamp)
	if processA.Time() != 5 {
		t.Errorf("Process A: Expected time 5, got %d", processA.Time())
	}
}

func TestMultipleProcesses(t *testing.T) {
	p1 := NewLamportClock()
	p2 := NewLamportClock()
	p3 := NewLamportClock()

	// P1 → P2
	ts1 := p1.Send() // P1: 1
	p2.Receive(ts1)  // P2: 2

	// P2 → P3
	ts2 := p2.Send() // P2: 3
	p3.Receive(ts2)  // P3: 4

	// P3 → P1
	ts3 := p3.Send() // P3: 5
	p1.Receive(ts3)  // P1: 6

	if p1.Time() != 6 {
		t.Errorf("P1: Expected time 6, got %d", p1.Time())
	}
	if p2.Time() != 3 {
		t.Errorf("P2: Expected time 3, got %d", p2.Time())
	}
	if p3.Time() != 5 {
		t.Errorf("P3: Expected time 5, got %d", p3.Time())
	}
}

func TestReceiveFromFuture(t *testing.T) {
	localClock := NewLamportClock()
	remoteClock := NewLamportClock()

	// Local does a few ticks
	localClock.Tick() // 1
	localClock.Tick() // 2
	localClock.Tick() // 3

	// Remote does many ticks
	for i := 0; i < 10; i++ {
		remoteClock.Tick()
	}

	// Remote sends
	msgTimestamp := remoteClock.Send() // 11

	// Local receives from "future"
	localClock.Receive(msgTimestamp)

	// Should jump to max(3, 11) + 1 = 12
	if localClock.Time() != 12 {
		t.Errorf("Expected time 12, got %d", localClock.Time())
	}
}

func TestCausalityOrdering(t *testing.T) {
	sender := NewLamportClock()
	receiver := NewLamportClock()

	sendTime := sender.Send()
	receiveTime := receiver.Receive(sendTime)

	// Receive must happen after send
	if receiveTime <= sendTime {
		t.Errorf("Causality violation: receive(%d) should be > send(%d)",
			receiveTime, sendTime)
	}
}

func TestConcurrentUpdates(t *testing.T) {
	clock := NewLamportClock()
	done := make(chan bool)

	// Concurrent ticks
	for i := 0; i < 100; i++ {
		go func() {
			clock.Tick()
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < 100; i++ {
		<-done
	}

	if clock.Time() != 100 {
		t.Errorf("Expected time 100, got %d", clock.Time())
	}
}
