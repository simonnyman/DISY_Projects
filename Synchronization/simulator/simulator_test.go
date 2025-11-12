package simulator

import (
	"testing"
	"time"
)

// Basic functionality tests
func TestNewSimulator(t *testing.T) {
	sim := NewSimulator(5)

	if sim.NumProcesses != 5 {
		t.Errorf("Expected 5 processes, got %d", sim.NumProcesses)
	}

	if len(sim.Processes) != 5 {
		t.Errorf("Expected 5 processes in array, got %d", len(sim.Processes))
	}

	// Check each process is initialized
	for i, p := range sim.Processes {
		if p.ID != i {
			t.Errorf("Process %d has wrong ID: %d", i, p.ID)
		}
		if p.LamportClock == nil {
			t.Errorf("Process %d has nil Lamport clock", i)
		}
		if p.VectorClock == nil {
			t.Errorf("Process %d has nil Vector clock", i)
		}
	}
}

func TestLocalEvent(t *testing.T) {
	sim := NewSimulator(3)

	sim.generateLocalEvent(0)

	if len(sim.Processes[0].Events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(sim.Processes[0].Events))
	}

	event := sim.Processes[0].Events[0]

	if event.EventType != "local" {
		t.Errorf("Expected 'local' event type, got %s", event.EventType)
	}

	if event.Timestamp != 1 {
		t.Errorf("Expected Lamport timestamp 1, got %d", event.Timestamp)
	}

	if event.VectorTime[0] != 1 {
		t.Errorf("Expected vector clock [1,0,0], got %v", event.VectorTime)
	}

	if event.TargetID != -1 {
		t.Errorf("Local event should have TargetID -1, got %d", event.TargetID)
	}
}

func TestSendMessage(t *testing.T) {
	sim := NewSimulator(3)

	sim.sendMessage(0, 1)

	// Check send event on P0
	if len(sim.Processes[0].Events) != 1 {
		t.Fatalf("Expected 1 event on P0, got %d", len(sim.Processes[0].Events))
	}

	sendEvent := sim.Processes[0].Events[0]

	if sendEvent.EventType != "send" {
		t.Errorf("Expected 'send' event type, got %s", sendEvent.EventType)
	}

	if sendEvent.TargetID != 1 {
		t.Errorf("Expected target ID 1, got %d", sendEvent.TargetID)
	}

	if sendEvent.MessageID != 0 {
		t.Errorf("Expected first message ID 0, got %d", sendEvent.MessageID)
	}
}

func TestReceiveMessage(t *testing.T) {
	sim := NewSimulator(2)

	// Send and wait for receive
	sim.sendMessage(0, 1)
	time.Sleep(20 * time.Millisecond)

	// P1 should have received
	if len(sim.Processes[1].Events) == 0 {
		t.Fatal("P1 should have received message")
	}

	receiveEvent := sim.Processes[1].Events[0]

	if receiveEvent.EventType != "receive" {
		t.Errorf("Expected 'receive' event type, got %s", receiveEvent.EventType)
	}

	if receiveEvent.TargetID != 0 {
		t.Errorf("Expected message from P0, got P%d", receiveEvent.TargetID)
	}

	// Vector clock should show synchronization
	if receiveEvent.VectorTime[0] == 0 {
		t.Error("P1 should have knowledge of P0's events")
	}
}

func TestMultipleMessages(t *testing.T) {
	sim := NewSimulator(3)

	// Send multiple messages
	sim.sendMessage(0, 1)
	sim.sendMessage(0, 2)
	sim.sendMessage(1, 2)

	time.Sleep(30 * time.Millisecond)

	// Check message IDs are unique
	if sim.Processes[0].Events[0].MessageID == sim.Processes[0].Events[1].MessageID {
		t.Error("Message IDs should be unique")
	}
}

func TestGetStatistics(t *testing.T) {
	sim := NewSimulator(2)

	sim.generateLocalEvent(0)
	sim.sendMessage(0, 1)
	time.Sleep(20 * time.Millisecond)

	stats := sim.GetStatistics()

	if stats["local_events"].(int) != 1 {
		t.Errorf("Expected 1 local event, got %d", stats["local_events"])
	}

	if stats["send_events"].(int) != 1 {
		t.Errorf("Expected 1 send event, got %d", stats["send_events"])
	}

	if stats["receive_events"].(int) != 1 {
		t.Errorf("Expected 1 receive event, got %d", stats["receive_events"])
	}

	totalEvents := stats["total_events"].(int)
	if totalEvents != 3 {
		t.Errorf("Expected 3 total events, got %d", totalEvents)
	}
}

func TestConcurrencyDetection(t *testing.T) {
	sim := NewSimulator(2)

	// Create concurrent events (no communication)
	sim.generateLocalEvent(0) // [1, 0]
	sim.generateLocalEvent(1) // [0, 1]

	concurrent := sim.CountConcurrentEvents()

	if concurrent != 1 {
		t.Errorf("Expected 1 concurrent pair, got %d", concurrent)
	}
}

// Scenario-based tests
func TestVectorClockSynchronization(t *testing.T) {
	sim := NewSimulator(2)

	// P0: local event
	sim.generateLocalEvent(0) // P0: [1, 0]

	// P0: send to P1
	sim.sendMessage(0, 1) // P0: [2, 0]

	time.Sleep(20 * time.Millisecond)

	// P1 should have synchronized vector clock
	receiveEvent := sim.Processes[1].Events[0]

	// P1's vector should be [2, 1] (knows P0 did 2 things, P1 did 1)
	if receiveEvent.VectorTime[0] != 2 {
		t.Errorf("P1 should know P0 has done 2 events, got %d", receiveEvent.VectorTime[0])
	}

	if receiveEvent.VectorTime[1] != 1 {
		t.Errorf("P1 should have incremented own clock to 1, got %d", receiveEvent.VectorTime[1])
	}
}

func TestMessageChain(t *testing.T) {
	sim := NewSimulator(3)

	// Create a message chain: P0 → P1 → P2
	sim.sendMessage(0, 1) // P0: [1, 0, 0]
	time.Sleep(20 * time.Millisecond)

	sim.sendMessage(1, 2) // P1: [1, 2, 0] (after receiving from P0)
	time.Sleep(20 * time.Millisecond)

	// P2 should have transitive knowledge
	if len(sim.Processes[2].Events) == 0 {
		t.Fatal("P2 should have received message")
	}

	receiveEvent := sim.Processes[2].Events[0]

	// P2 should know about P0 and P1
	if receiveEvent.VectorTime[0] == 0 {
		t.Error("P2 should have transitive knowledge of P0's events")
	}
	if receiveEvent.VectorTime[1] == 0 {
		t.Error("P2 should have knowledge of P1's events")
	}
}

func TestBroadcastPattern(t *testing.T) {
	sim := NewSimulator(4)

	// P0 broadcasts to all others
	sim.sendMessage(0, 1)
	sim.sendMessage(0, 2)
	sim.sendMessage(0, 3)

	time.Sleep(30 * time.Millisecond)

	// All other processes should have received
	for i := 1; i < 4; i++ {
		if len(sim.Processes[i].Events) == 0 {
			t.Errorf("Process %d should have received message", i)
		}
	}
}

func TestCausalOrderingPreserved(t *testing.T) {
	sim := NewSimulator(2)

	// P0 does two sends
	sim.sendMessage(0, 1) // msg0: [1, 0]
	sim.sendMessage(0, 1) // msg1: [2, 0]

	time.Sleep(30 * time.Millisecond)

	// P1 should receive both in order
	if len(sim.Processes[1].Events) != 2 {
		t.Fatalf("Expected 2 receive events, got %d", len(sim.Processes[1].Events))
	}

	event1 := sim.Processes[1].Events[0]
	event2 := sim.Processes[1].Events[1]

	// Second receive should have higher timestamp
	if event2.Timestamp <= event1.Timestamp {
		t.Error("Causal ordering not preserved in receive timestamps")
	}
}

func TestRunSimulation(t *testing.T) {
	sim := NewSimulator(3)

	// Run short simulation
	sim.RunSimulation(100*time.Millisecond, 0.3, 0.4)

	stats := sim.GetStatistics()
	totalEvents := stats["total_events"].(int)

	if totalEvents == 0 {
		t.Error("Simulation should have generated events")
	}

	// Each process should have some events
	for i, p := range sim.Processes {
		if len(p.Events) == 0 {
			t.Errorf("Process %d should have generated events", i)
		}
	}
}

func TestSimulationWithDifferentProbabilities(t *testing.T) {
	tests := []struct {
		name      string
		localProb float64
		sendProb  float64
	}{
		{"HighLocal", 0.8, 0.1},
		{"HighSend", 0.1, 0.8},
		{"Balanced", 0.4, 0.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := NewSimulator(3)
			sim.RunSimulation(100*time.Millisecond, tt.localProb, tt.sendProb)

			stats := sim.GetStatistics()
			if stats["total_events"].(int) == 0 {
				t.Error("Simulation should generate events")
			}
		})
	}
}

func TestProcessStatistics(t *testing.T) {
	sim := NewSimulator(3)

	sim.generateLocalEvent(0)
	sim.generateLocalEvent(0)
	sim.sendMessage(0, 1)
	time.Sleep(20 * time.Millisecond)

	processStats := sim.GetProcessStatistics()

	if len(processStats) != 3 {
		t.Errorf("Expected stats for 3 processes, got %d", len(processStats))
	}

	// P0 should have 3 events (2 local + 1 send)
	p0Stats := processStats[0]

	if p0Stats["total_events"].(int) != 3 {
		t.Errorf("P0 should have 3 events, got %d", p0Stats["total_events"])
	}

	if p0Stats["local_events"].(int) != 2 {
		t.Errorf("P0 should have 2 local events, got %d", p0Stats["local_events"])
	}

	if p0Stats["send_events"].(int) != 1 {
		t.Errorf("P0 should have 1 send event, got %d", p0Stats["send_events"])
	}
}
