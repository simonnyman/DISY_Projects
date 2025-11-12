package simulator

import vector "github.com/simonnyman/DISY_Projects/Synchronization/vectorCustom"

// returns statistics about the simulation
func (s *Simulator) GetStatistics() map[string]interface{} {
	totalEvents := len(s.Events)
	localEvents := 0
	sendEvents := 0
	receiveEvents := 0

	for _, e := range s.Events {
		switch e.EventType {
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
	}
}

// counts how many event pairs are concurrent
func (s *Simulator) CountConcurrentEvents() int {
	concurrent := 0
	events := s.Events

	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if areConcurrent(events[i].VectorTime, events[j].VectorTime) {
				concurrent++
			}
		}
	}

	return concurrent
}

// checks if two vector clock timestamps are concurrent
func areConcurrent(v1, v2 []int64) bool {
	return vector.CompareClocks(v1, v2) == vector.Concurrent
}

// checks if event with v1 happened before event with v2
func HappenedBefore(v1, v2 []int64) bool {
	return vector.CompareClocks(v1, v2) == vector.Before
}

// checks if two vector timestamps are identical
func AreEqual(v1, v2 []int64) bool {
	return vector.CompareClocks(v1, v2) == vector.Equal
}
