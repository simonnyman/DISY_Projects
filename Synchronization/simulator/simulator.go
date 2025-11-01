package simulator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/simonnyman/DISY_Projects/synchronization/lamport"
	"github.com/simonnyman/DISY_Projects/synchronization/vectorclock"
)

// Event represents an event in the distributed system
type Event struct {
	ProcessID  int
	EventType  string // "local", "send", "receive"
	Timestamp  int64
	VectorTime []int64
	MessageID  int
}

// Process represents a process in the distributed system
type Process struct {
	ID           int
	LamportClock *lamport.Clock
	VectorClock  *vectorclock.VectorClock
	Events       []Event
	mutex        sync.Mutex
}

// Simulator simulates a distributed system with multiple processes
type Simulator struct {
	Processes       []*Process
	NumProcesses    int
	MessageChannels []chan *Message
	Events          []Event
	eventsMutex     sync.RWMutex
}

// Message represents a message sent between processes
type Message struct {
	From        int
	To          int
	ID          int
	LamportTime int64
	VectorTime  []int64
	Data        string
}

// NewSimulator creates a new simulator with the specified number of processes
func NewSimulator(numProcesses int) *Simulator {
	processes := make([]*Process, numProcesses)
	channels := make([]chan *Message, numProcesses)

	for i := 0; i < numProcesses; i++ {
		processes[i] = &Process{
			ID:           i,
			LamportClock: lamport.NewClock(),
			VectorClock:  vectorclock.NewVectorClock(i, numProcesses),
			Events:       make([]Event, 0),
		}
		channels[i] = make(chan *Message, 100)
	}

	return &Simulator{
		Processes:       processes,
		NumProcesses:    numProcesses,
		MessageChannels: channels,
		Events:          make([]Event, 0),
	}
}

// RunSimulation runs a simulation with specified parameters
func (s *Simulator) RunSimulation(duration time.Duration, localEventRate, messageRate float64) {
	var generatorWg sync.WaitGroup
	var receiverWg sync.WaitGroup

	// Start receiver goroutines
	for i := 0; i < s.NumProcesses; i++ {
		receiverWg.Add(1)
		go s.processReceiver(i, &receiverWg)
	}

	// Start event generator goroutines
	for i := 0; i < s.NumProcesses; i++ {
		generatorWg.Add(1)
		go s.eventGenerator(i, duration, localEventRate, messageRate, &generatorWg)
	}

	// Wait for generators to finish
	generatorWg.Wait()

	// Close all channels
	for i := 0; i < s.NumProcesses; i++ {
		close(s.MessageChannels[i])
	}

	// Wait for receivers to finish processing remaining messages
	receiverWg.Wait()
}

// eventGenerator generates random events for a process
func (s *Simulator) eventGenerator(processID int, duration time.Duration, localEventRate, messageRate float64, wg *sync.WaitGroup) {
	defer wg.Done()

	process := s.Processes[processID]
	startTime := time.Now()
	messageID := 0

	for time.Since(startTime) < duration {
		// Decide event type based on rates
		r := rand.Float64()

		if r < localEventRate {
			// Local event
			lamportTime := process.LamportClock.Tick()
			vectorTime := process.VectorClock.Tick()

			event := Event{
				ProcessID:  processID,
				EventType:  "local",
				Timestamp:  lamportTime,
				VectorTime: vectorTime,
			}

			process.recordEvent(event)
			s.recordGlobalEvent(event)

		} else if r < localEventRate+messageRate {
			// Send message to random process
			targetProcess := rand.Intn(s.NumProcesses)
			if targetProcess == processID {
				targetProcess = (targetProcess + 1) % s.NumProcesses
			}

			lamportTime := process.LamportClock.Send()
			vectorTime := process.VectorClock.Send()

			msg := &Message{
				From:        processID,
				To:          targetProcess,
				ID:          messageID,
				LamportTime: lamportTime,
				VectorTime:  append([]int64{}, vectorTime...),
				Data:        fmt.Sprintf("msg_%d_from_%d", messageID, processID),
			}
			messageID++

			event := Event{
				ProcessID:  processID,
				EventType:  "send",
				Timestamp:  lamportTime,
				VectorTime: vectorTime,
				MessageID:  msg.ID,
			}

			process.recordEvent(event)
			s.recordGlobalEvent(event)

			// Send message (non-blocking)
			select {
			case s.MessageChannels[targetProcess] <- msg:
			default:
				// Channel full, skip message
			}
		}

		// Sleep for a short random duration
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)+1))
	}
}

// processReceiver receives and processes messages for a process
func (s *Simulator) processReceiver(processID int, wg *sync.WaitGroup) {
	defer wg.Done()

	process := s.Processes[processID]

	for msg := range s.MessageChannels[processID] {
		lamportTime := process.LamportClock.Receive(msg.LamportTime)
		vectorTime := process.VectorClock.Receive(msg.VectorTime)

		event := Event{
			ProcessID:  processID,
			EventType:  "receive",
			Timestamp:  lamportTime,
			VectorTime: vectorTime,
			MessageID:  msg.ID,
		}

		process.recordEvent(event)
		s.recordGlobalEvent(event)
	}
}

// recordEvent records an event for a process
func (p *Process) recordEvent(event Event) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.Events = append(p.Events, event)
}

// recordGlobalEvent records an event in the global event list
func (s *Simulator) recordGlobalEvent(event Event) {
	s.eventsMutex.Lock()
	defer s.eventsMutex.Unlock()
	s.Events = append(s.Events, event)
}

// GetStatistics returns statistics about the simulation
func (s *Simulator) GetStatistics() map[string]interface{} {
	s.eventsMutex.RLock()
	defer s.eventsMutex.RUnlock()

	totalEvents := len(s.Events)
	localEvents := 0
	sendEvents := 0
	receiveEvents := 0

	for _, event := range s.Events {
		switch event.EventType {
		case "local":
			localEvents++
		case "send":
			sendEvents++
		case "receive":
			receiveEvents++
		}
	}

	return map[string]interface{}{
		"total_events":   totalEvents,
		"local_events":   localEvents,
		"send_events":    sendEvents,
		"receive_events": receiveEvents,
		"num_processes":  s.NumProcesses,
	}
}

// AnalyzeOrdering analyzes the ordering guarantees
func (s *Simulator) AnalyzeOrdering() map[string]interface{} {
	s.eventsMutex.RLock()
	defer s.eventsMutex.RUnlock()

	// Count concurrent events in vector clocks
	concurrentPairs := 0
	totalPairs := 0

	// Sample pairs of events to analyze
	sampleSize := 1000
	if len(s.Events) < sampleSize {
		sampleSize = len(s.Events)
	}

	for i := 0; i < sampleSize; i++ {
		for j := i + 1; j < sampleSize && j < len(s.Events); j++ {
			e1 := s.Events[i]
			e2 := s.Events[j]

			totalPairs++

			if len(e1.VectorTime) > 0 && len(e2.VectorTime) > 0 {
				ordering := vectorclock.CompareVectorClocks(e1.VectorTime, e2.VectorTime)
				if ordering == vectorclock.Concurrent {
					concurrentPairs++
				}
			}
		}
	}

	concurrencyRate := 0.0
	if totalPairs > 0 {
		concurrencyRate = float64(concurrentPairs) / float64(totalPairs)
	}

	return map[string]interface{}{
		"total_pairs_analyzed": totalPairs,
		"concurrent_pairs":     concurrentPairs,
		"concurrency_rate":     concurrencyRate,
	}
}
