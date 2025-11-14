package vector

import (
	"reflect"
	"testing"
)

// Basic functionality tests

// creates a new Vector clock and verifies initial state is zero.
func TestNewVector(t *testing.T) {
	v := NewVector(0, 3)

	clock := v.Clock()
	expected := []int64{0, 0, 0}

	if !reflect.DeepEqual(clock, expected) {
		t.Errorf("Expected %v, got %v", expected, clock)
	}
}

// verifies tick increments only the process's own counter.
func TestVectorTick(t *testing.T) {
	v := NewVector(1, 3) // process 1 in 3-process system

	clock := v.Tick()
	expected := []int64{0, 1, 0}

	if !reflect.DeepEqual(clock, expected) {
		t.Errorf("Expected %v, got %v", expected, clock)
	}

	// second tick
	clock = v.Tick()
	expected = []int64{0, 2, 0}

	if !reflect.DeepEqual(clock, expected) {
		t.Errorf("Expected %v after second tick, got %v", expected, clock)
	}
}

// verifies send increments and returns correct timestamp.
func TestVectorSend(t *testing.T) {
	v := NewVector(0, 2)

	clock := v.Send()
	expected := []int64{1, 0}

	if !reflect.DeepEqual(clock, expected) {
		t.Errorf("Expected %v, got %v", expected, clock)
	}

	// send should increment like tick
	clock = v.Send()
	expected = []int64{2, 0}

	if !reflect.DeepEqual(clock, expected) {
		t.Errorf("Expected %v after second send, got %v", expected, clock)
	}
}

// verifies receive merges clocks and increments own counter.
func TestVectorReceive(t *testing.T) {
	v := NewVector(1, 3)
	v.Tick() // [0, 1, 0]

	// receive message with vector [2, 0, 1]
	received := []int64{2, 0, 1}
	clock := v.Receive(received)

	// should be [2, 2, 1] (max of each + increment own)
	expected := []int64{2, 2, 1}

	if !reflect.DeepEqual(clock, expected) {
		t.Errorf("Expected %v, got %v", expected, clock)
	}
}

// verifies receive handles lower timestamp correctly.
func TestVectorReceiveLowerTimestamp(t *testing.T) {
	v := NewVector(0, 3)
	v.Tick() // [1, 0, 0]
	v.Tick() // [2, 0, 0]
	v.Tick() // [3, 0, 0]

	// receive message with lower timestamp
	received := []int64{1, 0, 0}
	clock := v.Receive(received)

	// should be [3, 0, 0] + increment own = [4, 0, 0]
	expected := []int64{4, 0, 0}

	if !reflect.DeepEqual(clock, expected) {
		t.Errorf("Expected %v, got %v", expected, clock)
	}
}

// verifies reset sets all components to zero.
func TestVectorReset(t *testing.T) {
	v := NewVector(0, 3)
	v.Tick()
	v.Tick()

	v.Reset()

	clock := v.Clock()
	expected := []int64{0, 0, 0}

	if !reflect.DeepEqual(clock, expected) {
		t.Errorf("Expected %v after reset, got %v", expected, clock)
	}
}

// verifies CompareClocks correctly identifies all relationships.
func TestCompareClocks(t *testing.T) {
	tests := []struct {
		name     string
		v1       []int64
		v2       []int64
		expected Ordering
	}{
		{
			name:     "equal clocks",
			v1:       []int64{1, 2, 3},
			v2:       []int64{1, 2, 3},
			expected: Equal,
		},
		{
			name:     "before relationship",
			v1:       []int64{1, 2, 0},
			v2:       []int64{2, 3, 1},
			expected: Before,
		},
		{
			name:     "after relationship",
			v1:       []int64{3, 2, 1},
			v2:       []int64{1, 1, 0},
			expected: After,
		},
		{
			name:     "concurrent clocks",
			v1:       []int64{1, 0, 2},
			v2:       []int64{0, 3, 1},
			expected: Concurrent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareClocks(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("CompareClocks(%v, %v) = %v, expected %v",
					tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

// verifies String() method for all Ordering values.
func TestOrderingString(t *testing.T) {
	tests := []struct {
		ordering Ordering
		expected string
	}{
		{Before, "Before"},
		{After, "After"},
		{Concurrent, "Concurrent"},
		{Equal, "Equal"},
		{Ordering(999), "Unknown"}, // test default case
	}

	for _, tt := range tests {
		result := tt.ordering.String()
		if result != tt.expected {
			t.Errorf("Ordering(%d).String() = %s, expected %s",
				tt.ordering, result, tt.expected)
		}
	}
}

// verifies panic on mismatched vector lengths.
func TestCompareClocksPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when comparing different length vectors")
		}
	}()

	v1 := []int64{1, 2}
	v2 := []int64{1, 2, 3}
	CompareClocks(v1, v2)
}

// Scenario-based tests

// simulates message exchange between two processes.
func TestVectorClockMessagePassing(t *testing.T) {
	processA := NewVector(0, 2)
	processB := NewVector(1, 2)

	// process A does local work
	clockA := processA.Tick()
	expectedA := []int64{1, 0}
	if !reflect.DeepEqual(clockA, expectedA) {
		t.Errorf("Process A: Expected %v, got %v", expectedA, clockA)
	}

	// process A sends message
	msgTimestamp := processA.Send()
	expectedMsg := []int64{2, 0}
	if !reflect.DeepEqual(msgTimestamp, expectedMsg) {
		t.Errorf("Message timestamp: Expected %v, got %v", expectedMsg, msgTimestamp)
	}

	// process B receives the message
	clockB := processB.Receive(msgTimestamp)
	expectedB := []int64{2, 1} // max(0,2), max(0,0)+1
	if !reflect.DeepEqual(clockB, expectedB) {
		t.Errorf("Process B: Expected %v, got %v", expectedB, clockB)
	}

	// process B sends reply
	replyTimestamp := processB.Send()
	expectedReply := []int64{2, 2}
	if !reflect.DeepEqual(replyTimestamp, expectedReply) {
		t.Errorf("Reply timestamp: Expected %v, got %v", expectedReply, replyTimestamp)
	}

	// process A receives reply
	clockA = processA.Receive(replyTimestamp)
	expectedA = []int64{3, 2} // max(2,2)+1, max(0,2)
	if !reflect.DeepEqual(clockA, expectedA) {
		t.Errorf("Process A after receive: Expected %v, got %v", expectedA, clockA)
	}
}

// tests message chain across three processes (P1 → P2 → P3 → P1).
func TestVectorClockMultipleProcesses(t *testing.T) {
	p1 := NewVector(0, 3)
	p2 := NewVector(1, 3)
	p3 := NewVector(2, 3)

	// P1 → P2
	ts1 := p1.Send() // [1, 0, 0]
	p2.Receive(ts1)  // [1, 1, 0]

	// P2 → P3
	ts2 := p2.Send() // [1, 2, 0]
	p3.Receive(ts2)  // [1, 2, 1]

	// P3 → P1
	ts3 := p3.Send() // [1, 2, 2]
	p1.Receive(ts3)  // [2, 2, 2]

	// verify final states
	clock1 := p1.Clock()
	expected1 := []int64{2, 2, 2}
	if !reflect.DeepEqual(clock1, expected1) {
		t.Errorf("P1: Expected %v, got %v", expected1, clock1)
	}

	clock2 := p2.Clock()
	expected2 := []int64{1, 2, 0}
	if !reflect.DeepEqual(clock2, expected2) {
		t.Errorf("P2: Expected %v, got %v", expected2, clock2)
	}

	clock3 := p3.Clock()
	expected3 := []int64{1, 2, 2}
	if !reflect.DeepEqual(clock3, expected3) {
		t.Errorf("P3: Expected %v, got %v", expected3, clock3)
	}
}

// verifies clock jumps forward when receiving timestamp from "future".
func TestVectorClockReceiveFromFuture(t *testing.T) {
	localClock := NewVector(0, 2)
	remoteClock := NewVector(1, 2)

	// local does a few ticks
	localClock.Tick() // [1, 0]
	localClock.Tick() // [2, 0]
	localClock.Tick() // [3, 0]

	// remote does many ticks
	for i := 0; i < 10; i++ {
		remoteClock.Tick()
	}

	// remote sends
	msgTimestamp := remoteClock.Send() // [0, 11]

	// local receives from "future"
	clock := localClock.Receive(msgTimestamp)

	// should be [max(3,0)+1, max(0,11)] = [4, 11]
	expected := []int64{4, 11}
	if !reflect.DeepEqual(clock, expected) {
		t.Errorf("Expected %v, got %v", expected, clock)
	}
}

// verifies detection of concurrent events.
func TestVectorClockConcurrencyDetection(t *testing.T) {
	p1 := NewVector(0, 2)
	p2 := NewVector(1, 2)

	// both processes do local work independently
	clock1 := p1.Tick() // [1, 0]
	clock2 := p2.Tick() // [0, 1]

	// these should be concurrent
	ordering := CompareClocks(clock1, clock2)
	if ordering != Concurrent {
		t.Errorf("Expected Concurrent, got %v", ordering)
	}
}

// verifies causality ordering: send happens before receive.
func TestVectorClockCausalityOrdering(t *testing.T) {
	sender := NewVector(0, 2)
	receiver := NewVector(1, 2)

	sendTime := sender.Send()                 // [1, 0]
	receiveTime := receiver.Receive(sendTime) // [1, 1]

	// receive must happen after send
	ordering := CompareClocks(sendTime, receiveTime)
	if ordering != Before {
		t.Errorf("Causality violation: send should happen Before receive, got %v", ordering)
	}
}

// verifies transitive causality (P1 → P2 → P3, then P1 before P3).
func TestVectorClockTransitivity(t *testing.T) {
	p1 := NewVector(0, 3)
	p2 := NewVector(1, 3)
	p3 := NewVector(2, 3)

	// P1 → P2 → P3 (transitive causality chain)
	ts1 := p1.Send()       // [1, 0, 0]
	p2.Receive(ts1)        // [1, 1, 0]
	ts2 := p2.Send()       // [1, 2, 0]
	ts3 := p3.Receive(ts2) // [1, 2, 1]

	// ts1 should happen before ts3 (transitivity)
	ordering := CompareClocks(ts1, ts3)
	if ordering != Before {
		t.Errorf("Expected transitive Before relationship, got %v", ordering)
	}
}
