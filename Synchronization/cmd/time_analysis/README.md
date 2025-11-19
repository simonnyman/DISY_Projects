# Time Analysis - Simulation Duration Impact

This program analyzes how the **duration** of the simulation affects system behavior, keeping the number of processes fixed at 10.

## Running the Analysis

```bash
cd Synchronization
go run cmd/time_analysis/main.go
```

## Generated Plots

### 1. **time_event_growth.png** - Event Accumulation Over Time
Shows how total events accumulate as simulation time increases.

**Expected behavior:** Linear growth (more time = more events)

**Key insight:** The slope of this line indicates the overall event generation rate of the system.

### 2. **time_event_rate.png** - Event Generation Rate Stability
Shows events per second across different simulation durations.

**Expected behavior:** Should stabilize after initialization, showing consistent event generation regardless of duration.

**Key insight:** If this stabilizes, it means the system reaches a steady state. Variations might indicate initialization overhead or system saturation.

### 3. **time_event_types.png** - Event Type Distribution Over Time
Shows how local, send, and receive events accumulate independently.

**Expected behavior:** All should grow linearly with similar slopes (due to fixed probabilities).

**Key insight:** 
- Local events scale with process count × time
- Send/Receive should be roughly equal (each send creates a receive)
- Ratio between event types should remain constant

### 4. **time_concurrency.png** - Concurrency Rate Evolution
Shows how the percentage of concurrent event pairs changes with simulation duration.

**Expected behavior:** May slightly decrease as simulations run longer.

**Key insight:** Longer simulations create more causal dependencies (messages create ordering), which can reduce the proportion of concurrent events. In very short simulations, processes are more independent.

### 5. **time_clock_values.png** - Logical Time Progression
Shows average Lamport and Vector clock values as simulation progresses.

**Expected behavior:** Linear growth with time.

**Key insight:**
- Shows how "logical time" advances in the system
- Lamport values track total event count at each process
- Vector clock sums show similar growth but account for all processes
- The rate of growth indicates system activity level

### 6. **time_efficiency.png** - Normalized Event Rate
Shows events per 100ms time window - a normalized efficiency metric.

**Expected behavior:** Should be relatively flat after system initialization.

**Key insight:**
- Flat line = consistent throughput
- If decreasing: system saturation or resource constraints
- If increasing: system warming up or initialization effects
- Measures whether the system maintains efficiency over time

## Configuration

Modify these constants in `cmd/time_analysis/main.go`:

```go
const (
    numProcesses   = 10  // fixed for time analysis
    localEventProb = 0.3
    sendEventProb  = 0.4
    numRuns        = 3   // runs per duration (for averaging)
)

var simulationDurations = []int{100, 250, 500, 750, 1000, 1500, 2000, 3000, 5000}
```

## Key Questions Answered

### 1. **Does the system reach steady state?**
Check the event rate plot - if it stabilizes, the system operates consistently.

### 2. **How does logical time relate to real time?**
The clock values plot shows logical time progression. The ratio of logical to physical time indicates system activity.

### 3. **Is there initialization overhead?**
Compare short vs long simulations in the efficiency plot. If short simulations have different rates, there's initialization cost.

### 4. **How do causal relationships evolve?**
The concurrency plot shows whether the system becomes more or less concurrent over time.

### 5. **Are event rates stable?**
The event rate and efficiency plots reveal whether the system maintains consistent throughput.

## Practical Implications

**For System Design:**
- **Stable rates** mean the system is predictable and can be capacity-planned
- **Linear growth** confirms the system scales well temporally
- **Consistent efficiency** indicates no performance degradation over time

**For Understanding Distributed Systems:**
- Shows that logical time (clocks) advances independently of physical time
- Demonstrates that event generation in distributed systems follows probabilistic patterns
- Illustrates how causal dependencies accumulate over time

**For Academic Analysis:**
- Validates theoretical models (O(1) operations per event)
- Shows empirical behavior matches expected complexity
- Provides data for discussing system properties in papers/presentations

## Comparison with Process Scaling Analysis

**Process Scaling** (`cmd/plots/`):
- Varies number of processes (2-30)
- Fixed simulation time
- Shows **spatial** scalability (more nodes)
- Demonstrates O(n²) communication complexity

**Time Analysis** (this):
- Fixed number of processes (10)
- Varies simulation duration (100ms-5s)
- Shows **temporal** behavior (over time)
- Demonstrates steady-state characteristics

Together, these analyses provide a complete picture of system scalability and behavior.
