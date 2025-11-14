package simulator

import (
	"math/rand"
	"sync"
	"time"

	lamport "github.com/simonnyman/DISY_Projects/Synchronization/lamport"
	vector "github.com/simonnyman/DISY_Projects/Synchronization/vector"
)

// Event represents a single event in the distributed system.
type Event struct {
	ProcessID  int     // process that generated the event
	EventType  string  // "local", "send", or "receive"
	Timestamp  int64   // Lamport timestamp
	VectorTime []int64 // Vector clock timestamp
	TargetID   int     // for send: receiver, for receive: sender, -1 for local
	MessageID  int     // unique message identifier, -1 for local events
}

// Process represents a single process in the distributed system.
type Process struct {
	ID           int
	LamportClock *lamport.LamportClock
	VectorClock  *vector.Vector
	Events       []Event
	inbox        chan *Message
}

// Simulator manages the distributed system simulation.
type Simulator struct {
	Processes        []*Process
	NumProcesses     int
	Events           []Event
	messageIDCounter int
	counterMu        sync.Mutex // protects messageIDCounter
	eventsMu         sync.Mutex // protects Events slice
}

// Message represents a message sent between processes.
type Message struct {
	From        int
	To          int
	LamportTime int64
	VectorTime  []int64
	MessageID   int
}

// creates a new simulator with the specified number of processes.
// panics if numProcesses is less than 1.
func NewSimulator(numProcesses int) *Simulator {
	if numProcesses < 1 {
		panic("simulator: number of processes must be at least 1")
	}

	processes := make([]*Process, numProcesses)

	for i := 0; i < numProcesses; i++ {
		processes[i] = &Process{
			ID:           i,
			LamportClock: lamport.NewLamportClock(),
			VectorClock:  vector.NewVector(i, numProcesses),
			Events:       make([]Event, 0),
			inbox:        make(chan *Message, 100),
		}
	}

	return &Simulator{
		Processes:        processes,
		NumProcesses:     numProcesses,
		Events:           make([]Event, 0),
		messageIDCounter: 0,
	}
}

// generates a local event for the specified process.
// panics if processID is out of bounds.
func (s *Simulator) generateLocalEvent(processID int) {
	if processID < 0 || processID >= s.NumProcesses {
		panic("simulator: processID out of bounds")
	}

	p := s.Processes[processID]

	lt := p.LamportClock.Tick()
	vt := p.VectorClock.Tick()

	e := Event{
		ProcessID:  processID,
		EventType:  "local",
		Timestamp:  lt,
		VectorTime: vt,
		TargetID:   -1,
		MessageID:  -1,
	}

	p.Events = append(p.Events, e)
	s.appendEvent(e)
}

// sends a message from one process to another.
// panics if fromID or toID is out of bounds.
func (s *Simulator) sendMessage(fromID, toID int) {
	if fromID < 0 || fromID >= s.NumProcesses {
		panic("simulator: fromID out of bounds")
	}
	if toID < 0 || toID >= s.NumProcesses {
		panic("simulator: toID out of bounds")
	}

	sender := s.Processes[fromID]
	receiver := s.Processes[toID]

	// update sender's clocks
	lt := sender.LamportClock.Send()
	vt := sender.VectorClock.Send()

	// get unique message ID
	s.counterMu.Lock()
	msgID := s.messageIDCounter
	s.messageIDCounter++
	s.counterMu.Unlock()

	// create and send the message
	msg := &Message{
		From:        fromID,
		To:          toID,
		LamportTime: lt,
		VectorTime:  vt,
		MessageID:   msgID,
	}
	receiver.inbox <- msg

	// record the send event
	e := Event{
		ProcessID:  fromID,
		EventType:  "send",
		Timestamp:  lt,
		VectorTime: vt,
		TargetID:   toID,
		MessageID:  msgID,
	}

	sender.Events = append(sender.Events, e)
	s.appendEvent(e)
}

// processes a received message and updates clocks.
// panics if processID is out of bounds.
func (s *Simulator) receiveMessage(processID int, msg *Message) {
	if processID < 0 || processID >= s.NumProcesses {
		panic("simulator: processID out of bounds")
	}

	receiver := s.Processes[processID]

	// update receiver's clocks with message timestamps
	lt := receiver.LamportClock.Receive(msg.LamportTime)
	vt := receiver.VectorClock.Receive(msg.VectorTime)

	// record the receive event
	e := Event{
		ProcessID:  processID,
		EventType:  "receive",
		Timestamp:  lt,
		VectorTime: vt,
		TargetID:   msg.From,
		MessageID:  msg.MessageID,
	}

	receiver.Events = append(receiver.Events, e)
	s.appendEvent(e)
}

// appends an event to the global event list in a thread-safe manner.
func (s *Simulator) appendEvent(e Event) {
	s.eventsMu.Lock()
	s.Events = append(s.Events, e)
	s.eventsMu.Unlock()
}

// runs the simulation for the specified duration.
func (s *Simulator) RunSimulation(duration time.Duration, localEventProb, sendEventProb float64) {
	if localEventProb < 0 || localEventProb > 1 {
		panic("simulator: localEventProb must be between 0 and 1")
	}
	if sendEventProb < 0 || sendEventProb > 1 {
		panic("simulator: sendEventProb must be between 0 and 1")
	}
	if duration <= 0 {
		panic("simulator: duration must be positive")
	}

	var wg sync.WaitGroup
	stopChan := make(chan bool)

	// start goroutines for each process
	for i := 0; i < s.NumProcesses; i++ {
		wg.Add(2)

		// event generator goroutine
		go func(processID int) {
			defer wg.Done()
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-stopChan:
					return
				case <-ticker.C:
					r := rand.Float64()
					if r < localEventProb {
						s.generateLocalEvent(processID)
					} else if r < localEventProb+sendEventProb {
						// send to random process (not self)
						toID := rand.Intn(s.NumProcesses)
						if toID != processID {
							s.sendMessage(processID, toID)
						}
					}
				}
			}
		}(i)

		// message receiver goroutine
		go func(processID int) {
			defer wg.Done()
			process := s.Processes[processID]

			for {
				select {
				case <-stopChan:
					return
				case msg := <-process.inbox:
					s.receiveMessage(processID, msg)
				}
			}
		}(i)
	}

	// run simulation for specified duration
	time.Sleep(duration)

	// stop event generation
	close(stopChan)

	// give receivers time to drain inboxes
	time.Sleep(50 * time.Millisecond)

	// wait for all goroutines to finish
	wg.Wait()
}
