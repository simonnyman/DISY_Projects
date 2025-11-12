package simulator

import (
	"math/rand"
	"sync"
	"time"

	lamport "github.com/simonnyman/DISY_Projects/Synchronization/lamportClock"
	vector "github.com/simonnyman/DISY_Projects/Synchronization/vectorCustom"
)

type Event struct {
	ProcessID  int
	EventType  string  // "local", "send", "receive"
	Timestamp  int64   // lamport time
	VectorTime []int64 // vector clock time
	TargetID   int     // for send: who receives, For receive: who sent
	MessageID  int     // unique message identifier
}

type Process struct {
	ID           int
	LamportClock *lamport.LamportClock
	VectorClock  *vector.Vector
	Events       []Event
	inbox        chan *Message
}

type Simulator struct {
	Processes        []*Process
	NumProcesses     int
	Events           []Event
	messageIDCounter int        // counter for unique message IDs
	mu               sync.Mutex // protect message ID counter
}

type Message struct {
	From        int
	To          int
	LamportTime int64
	VectorTime  []int64
	MessageID   int
}

// creates a new simulator with the specified number of processes
func NewSimulator(numProcesses int) *Simulator {
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

// process does a local event (no messaging)
func (s *Simulator) generateLocalEvent(processID int) {
	p := s.Processes[processID]

	lt := p.LamportClock.Tick()
	vt := p.VectorClock.Tick()

	e := Event{
		ProcessID:  processID,
		EventType:  "local",
		Timestamp:  lt,
		VectorTime: vt,
		TargetID:   -1, // no target for local events
		MessageID:  -1, // no message for local events
	}

	p.Events = append(p.Events, e)
	s.Events = append(s.Events, e)
}

// process sends a message to another process
func (s *Simulator) sendMessage(fromID, toID int) {
	sender := s.Processes[fromID]
	receiver := s.Processes[toID]

	// update sender's clocks
	lt := sender.LamportClock.Send()
	vt := sender.VectorClock.Send()

	// get unique message ID
	s.mu.Lock()
	msgID := s.messageIDCounter
	s.messageIDCounter++
	s.mu.Unlock()

	// create the message
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
	s.Events = append(s.Events, e)
}

// process receives a message from inbox
func (s *Simulator) receiveMessage(processID int, msg *Message) {
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
	s.Events = append(s.Events, e)
}

// runs the simulation for a given duration with specified event probabilities
func (s *Simulator) RunSimulation(duration time.Duration, localEventProb, sendEventProb float64) {
	var wg sync.WaitGroup
	stopChan := make(chan bool)

	// start a goroutine for each process
	for i := 0; i < s.NumProcesses; i++ {
		wg.Add(2) // one for event generator, one for receiver

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
						// send message to random process
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

	// let simulation run for the specified duration
	time.Sleep(duration)

	// stop all goroutines
	close(stopChan)
	wg.Wait()
}
