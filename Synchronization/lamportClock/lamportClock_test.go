package lamportClock

import (
	"testing"
)

// Message represents a message with a timestamp
// type Message struct {
// 	Content   string
// 	Timestamp int64
// }

func TestLamportClock(t *testing.T) {
	lc := NewLamportClock()

	lc.Tick()
	if lc.Time() != 1 {
		t.Errorf("Expected time 1, got %d", lc.Time())
	}

	lc.Send()
	if lc.Time() != 2 {
		t.Errorf("Expected time 2, got %d", lc.Time())
	}

	lc.Receive(1)
	if lc.Time() != 3 {
		t.Errorf("Expected time 3, got %d", lc.Time())
	}

	lc.Reset()
	if lc.Time() != 0 {
		t.Errorf("Expected time 0, got %d", lc.Time())
	}
}

// TestMessagePassing simulates message passing between two processes
func TestMessagePassing(t *testing.T) {
	// Create two processes with their own clocks
	processA := NewLamportClock()
	processB := NewLamportClock()

	// Process A does some local work
	processA.Tick()
	if processA.Time() != 1 {
		t.Errorf("Process A: Expected time 1, got %d", processA.Time())
	}

	// Process A sends a message to Process B
	msgTimestamp := processA.Send()
	msg := message{
		Data:      "Hello from A",
		timestamp: msgTimestamp,
	}
	if msg.timestamp != 2 {
		t.Errorf("Message timestamp should be 2, got %d", msg.timestamp)
	}

	// Process B receives the message
	processB.Receive(msg.timestamp)
	if processB.Time() != 3 {
		t.Errorf("Process B: Expected time 3 after receiving, got %d", processB.Time())
	}

	// Process B does local work
	processB.Tick()
	if processB.Time() != 4 {
		t.Errorf("Process B: Expected time 4, got %d", processB.Time())
	}

	// Process B sends a reply to Process A
	replyTimestamp := processB.Send()
	reply := message{
		Data:      "Reply from B",
		timestamp: replyTimestamp,
	}
	if reply.timestamp != 5 {
		t.Errorf("Reply timestamp should be 5, got %d", reply.timestamp)
	}

	// Process A receives the reply
	processA.Receive(reply.timestamp)
	if processA.Time() != 6 {
		t.Errorf("Process A: Expected time 6 after receiving reply, got %d", processA.Time())
	}
}

// TestMultipleProcesses simulates three processes exchanging messages
func TestMultipleProcesses(t *testing.T) {
	p1 := NewLamportClock()
	p2 := NewLamportClock()
	p3 := NewLamportClock()

	// P1 sends to P2
	ts1 := p1.Send() // P1: 1
	p2.Receive(ts1)  // P2: 2

	// P2 sends to P3
	ts2 := p2.Send() // P2: 3
	p3.Receive(ts2)  // P3: 4

	// P3 sends to P1
	ts3 := p3.Send() // P3: 5
	p1.Receive(ts3)  // P1: 6

	// Verify final times
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

// TestConcurrentMessages tests when a process receives a message with a much larger timestamp
func TestReceiveFromFuture(t *testing.T) {
	localClock := NewLamportClock()
	remoteClock := NewLamportClock()

	// Local process does a few ticks
	localClock.Tick() // 1
	localClock.Tick() // 2
	localClock.Tick() // 3

	// Remote process does many ticks
	for i := 0; i < 10; i++ {
		remoteClock.Tick()
	}

	// Remote sends message with timestamp 11
	msgTimestamp := remoteClock.Send()
	if msgTimestamp != 11 {
		t.Errorf("Remote clock should be 11, got %d", msgTimestamp)
	}

	// Local receives message from "future"
	localClock.Receive(msgTimestamp)
	// Should jump to max(3, 11) + 1 = 12
	if localClock.Time() != 12 {
		t.Errorf("Local clock should jump to 12, got %d", localClock.Time())
	}
}

// TestCausalityOrdering ensures send always happens before receive
func TestCausalityOrdering(t *testing.T) {
	sender := NewLamportClock()
	receiver := NewLamportClock()

	// Sender sends a message
	sendTime := sender.Send()

	// Receiver receives it
	receiveTime := receiver.Receive(sendTime)

	// Receive time must be greater than send time
	if receiveTime <= sendTime {
		t.Errorf("Causality violation: receive time %d should be > send time %d", receiveTime, sendTime)
	}
}

// TestCreateMessage tests the CreateMessage function
func TestCreateMessage(t *testing.T) {
	clock := NewLamportClock()

	// Send a message and create message object
	timestamp := clock.Send()
	msg := CreateMessage(timestamp, "test data", 1)

	if msg == nil {
		t.Fatal("CreateMessage returned nil")
	}
	if msg.timestamp != timestamp {
		t.Errorf("Expected timestamp %d, got %d", timestamp, msg.timestamp)
	}
	if msg.Data != "test data" {
		t.Errorf("Expected data 'test data', got %v", msg.Data)
	}
	if msg.PID != 1 {
		t.Errorf("Expected PID 1, got %d", msg.PID)
	}
}

// TestHappensBefore tests the HappensBefore method
func TestHappensBefore(t *testing.T) {
	clock1 := NewLamportClock()
	clock2 := NewLamportClock()

	// Create first message
	ts1 := clock1.Send()
	msg1 := CreateMessage(ts1, "first message", 1)

	// Create second message with later timestamp
	clock2.Receive(ts1) // Sync with clock1
	ts2 := clock2.Send()
	msg2 := CreateMessage(ts2, "second message", 2)

	// msg1 should happen before msg2
	if !msg1.HappensBefore(msg2) {
		t.Errorf("msg1 (ts=%d) should happen before msg2 (ts=%d)", msg1.timestamp, msg2.timestamp)
	}

	// msg2 should NOT happen before msg1
	if msg2.HappensBefore(msg1) {
		t.Errorf("msg2 (ts=%d) should NOT happen before msg1 (ts=%d)", msg2.timestamp, msg1.timestamp)
	}

	// Same timestamp should return false
	msg3 := CreateMessage(ts1, "concurrent message", 3)
	if msg1.HappensBefore(msg3) {
		t.Errorf("msg1 should NOT happen before msg3 when they have the same timestamp")
	}
}

// TestMessageOrdering tests complete message ordering scenario
func TestMessageOrdering(t *testing.T) {
	p1 := NewLamportClock()
	p2 := NewLamportClock()
	p3 := NewLamportClock()

	// Create messages from different processes
	ts1 := p1.Send()
	msg1 := CreateMessage(ts1, "Message from P1", 1)

	p2.Receive(ts1)
	ts2 := p2.Send()
	msg2 := CreateMessage(ts2, "Message from P2", 2)

	p3.Receive(ts2)
	ts3 := p3.Send()
	msg3 := CreateMessage(ts3, "Message from P3", 3)

	// Verify ordering
	if !msg1.HappensBefore(msg2) {
		t.Error("msg1 should happen before msg2")
	}
	if !msg2.HappensBefore(msg3) {
		t.Error("msg2 should happen before msg3")
	}
	if !msg1.HappensBefore(msg3) {
		t.Error("msg1 should happen before msg3 (transitivity)")
	}
}
