package vector

import (
	"sync"
	"testing"
)

func TestNewVectorClock(t *testing.T) {
	vc := NewVector(0, 3)
	clock := vc.Clock()

	if len(clock) != 3 {
		t.Errorf("Expected clock length 3, got %d", len(clock))
	}

	for i, val := range clock {
		if val != 0 {
			t.Errorf("Expected clock[%d] to be 0, got %d", i, val)
		}
	}
}

func TestTick(t *testing.T) {
	vc := NewVector(1, 3)

	clock := vc.Tick()
	if clock[1] != 1 {
		t.Errorf("Expected clock[1] to be 1, got %d", clock[1])
	}

	clock = vc.Tick()
	if clock[1] != 2 {
		t.Errorf("Expected clock[1] to be 2, got %d", clock[1])
	}
}

func TestSend(t *testing.T) {
	vc := NewVector(0, 3)

	timestamp := vc.Send()
	if timestamp[0] != 1 {
		t.Errorf("Expected timestamp[0] to be 1, got %d", timestamp[0])
	}

	// Verify clock was incremented
	clock := vc.Clock()
	if clock[0] != 1 {
		t.Errorf("Expected clock[0] to be 1 after send, got %d", clock[0])
	}
}

func TestReceive(t *testing.T) {
	vc := NewVector(0, 3)
	vc.Tick() // [1, 0, 0]

	received := []int64{2, 3, 1}
	newClock := vc.Receive(received)

	// After receive: max([1,0,0], [2,3,1]) + increment own = [3, 3, 1]
	// Process 0's counter should be incremented
	expected := []int64{3, 3, 1}

	for i, val := range newClock {
		if val != expected[i] {
			t.Errorf("After receive, expected clock[%d]=%d, got %d", i, expected[i], val)
		}
	}
}

func TestReset(t *testing.T) {
	vc := NewVector(1, 3)
	vc.Tick()
	vc.Tick()
	vc.Reset()

	clock := vc.Clock()
	for i, val := range clock {
		if val != 0 {
			t.Errorf("Expected clock[%d] to be 0 after reset, got %d", i, val)
		}
	}
}

func TestCompareVectorClocks(t *testing.T) {
	tests := []struct {
		name     string
		v1       []int64
		v2       []int64
		expected Ordering
	}{
		{
			name:     "Equal clocks",
			v1:       []int64{1, 2, 3},
			v2:       []int64{1, 2, 3},
			expected: Equal,
		},
		{
			name:     "v1 before v2",
			v1:       []int64{1, 1, 1},
			v2:       []int64{2, 2, 2},
			expected: Before,
		},
		{
			name:     "v1 after v2",
			v1:       []int64{3, 3, 3},
			v2:       []int64{1, 1, 1},
			expected: After,
		},
		{
			name:     "Concurrent events",
			v1:       []int64{2, 1, 0},
			v2:       []int64{1, 2, 0},
			expected: Concurrent,
		},
		{
			name:     "Partial order - before",
			v1:       []int64{1, 2, 3},
			v2:       []int64{2, 2, 3},
			expected: Before,
		},
		{
			name:     "Partial order - after",
			v1:       []int64{2, 3, 4},
			v2:       []int64{2, 2, 3},
			expected: After,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareClocks(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for v1=%v, v2=%v", tt.expected, result, tt.v1, tt.v2)
			}
		})
	}
}

func TestMessageCompareTo(t *testing.T) {
	msg1 := NewMessage([]int64{1, 2, 0}, 0, "msg1")
	msg2 := NewMessage([]int64{2, 2, 0}, 1, "msg2")
	msg3 := NewMessage([]int64{2, 1, 0}, 2, "msg3")

	// msg1 happened before msg2
	if msg1.CompareTo(msg2) != Before {
		t.Error("msg1 should happen before msg2")
	}

	// msg1 and msg3 are concurrent (msg1[0]=1<2, msg1[1]=2>1)
	if msg1.CompareTo(msg3) != Concurrent {
		t.Error("msg1 and msg3 should be concurrent")
	}
}

func TestMessageHappensBefore(t *testing.T) {
	msg1 := NewMessage([]int64{1, 1, 1}, 0, nil)
	msg2 := NewMessage([]int64{2, 2, 2}, 1, nil)
	msg3 := NewMessage([]int64{2, 0, 1}, 2, nil)

	if !msg1.HappensBefore(msg2) {
		t.Error("msg1 should happen before msg2")
	}

	// msg1=[1,1,1] vs msg3=[2,0,1]: msg1[0]<msg3[0] but msg1[1]>msg3[1], so concurrent
	if msg1.HappensBefore(msg3) {
		t.Error("msg1 should be concurrent with msg3, not before")
	}
}

func TestMessageIsConcurrent(t *testing.T) {
	msg1 := NewMessage([]int64{2, 1, 0}, 0, nil)
	msg2 := NewMessage([]int64{1, 2, 0}, 1, nil)
	msg3 := NewMessage([]int64{3, 3, 0}, 2, nil)

	if !msg1.IsConcurrent(msg2) {
		t.Error("msg1 and msg2 should be concurrent")
	}

	if msg1.IsConcurrent(msg3) {
		t.Error("msg1 and msg3 should not be concurrent")
	}
}

func TestConcurrency(t *testing.T) {
	numProcesses := 3
	vc := NewVectorClock(0, numProcesses)
	var wg sync.WaitGroup

	numGoroutines := 50
	ticksPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < ticksPerGoroutine; j++ {
				vc.Tick()
			}
		}()
	}

	wg.Wait()

	clock := vc.Clock()
	expectedTime := int64(numGoroutines * ticksPerGoroutine)
	if clock[0] != expectedTime {
		t.Errorf("Expected clock[0] to be %d after concurrent ticks, got %d", expectedTime, clock[0])
	}
}

func TestMessageCopy(t *testing.T) {
	original := []int64{1, 2, 3}
	msg := NewMessage(original, 0, nil)

	// Modify original
	original[0] = 999

	// Message should have its own copy
	if msg.VectorTime[0] == 999 {
		t.Error("Message should have independent copy of vector time")
	}
}

// Benchmark tests
func BenchmarkTick(b *testing.B) {
	vc := NewVectorClock(0, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vc.Tick()
	}
}

func BenchmarkSend(b *testing.B) {
	vc := NewVectorClock(0, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vc.Send()
	}
}

func BenchmarkReceive(b *testing.B) {
	vc := NewVectorClock(0, 10)
	received := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vc.Receive(received)
	}
}

func BenchmarkCompare(b *testing.B) {
	v1 := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	v2 := []int64{2, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareVectorClocks(v1, v2)
	}
}

func BenchmarkConcurrentOperations(b *testing.B) {
	vc := NewVectorClock(0, 10)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			vc.Tick()
		}
	})
}
