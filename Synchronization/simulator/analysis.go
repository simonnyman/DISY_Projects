package simulator

import vector "github.com/simonnyman/DISY_Projects/Synchronization/vector"

// returns detailed statistics for each process
func (s *Simulator) GetProcessStatistics() []map[string]interface{} {
	stats := make([]map[string]interface{}, s.NumProcesses)

	for i := 0; i < s.NumProcesses; i++ {
		p := s.Processes[i]
		local, send, receive := 0, 0, 0

		for _, e := range p.Events {
			switch e.EventType {
			case "local":
				local++
			case "send":
				send++
			case "receive":
				receive++
			}
		}

		stats[i] = map[string]interface{}{
			"process_id":     i,
			"total_events":   len(p.Events),
			"local_events":   local,
			"send_events":    send,
			"receive_events": receive,
		}
	}

	return stats
}

// returns aggregate statistics across all processes
func (s *Simulator) GetStatistics() map[string]interface{} {
	processStats := s.GetProcessStatistics()

	totalEvents := 0
	localEvents := 0
	sendEvents := 0
	receiveEvents := 0

	for _, pStats := range processStats {
		totalEvents += pStats["total_events"].(int)
		localEvents += pStats["local_events"].(int)
		sendEvents += pStats["send_events"].(int)
		receiveEvents += pStats["receive_events"].(int)
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

// analyzes all causal relationships
func (s *Simulator) CountCausalRelationships() map[string]int {
	before := 0
	after := 0
	concurrent := 0
	equal := 0
	events := s.Events

	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			ordering := vector.CompareClocks(events[i].VectorTime, events[j].VectorTime)
			switch ordering {
			case vector.Before:
				before++
			case vector.After:
				after++
			case vector.Concurrent:
				concurrent++
			case vector.Equal:
				equal++
			}
		}
	}

	return map[string]int{
		"before":     before,
		"after":      after,
		"concurrent": concurrent,
		"equal":      equal,
	}
}

// returns who communicated with whom
func (s *Simulator) GetCommunicationMatrix() [][]int {
	matrix := make([][]int, s.NumProcesses)
	for i := range matrix {
		matrix[i] = make([]int, s.NumProcesses)
	}

	for _, event := range s.Events {
		if event.EventType == "send" {
			matrix[event.ProcessID][event.TargetID]++
		}
	}

	return matrix
}
