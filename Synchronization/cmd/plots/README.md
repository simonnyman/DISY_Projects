# Distributed System Clock Analysis Plots

This directory contains a program that generates plots analyzing how the distributed system's behavior changes as the number of processes increases.

## Running the Plot Generator

```bash
cd Synchronization
go run cmd/plots/main.go
```

## Generated Plots

### 1. **event_count.png** - Total Events vs Number of Processes
Shows how the total number of events scales with the number of processes. This demonstrates the overall activity in the system.

**Expected behavior:** Should increase roughly linearly or quadratically depending on communication patterns.

### 2. **event_types.png** - Event Types Breakdown
Shows the breakdown of different event types (Local, Send, Receive) as processes increase.

**Expected behavior:** 
- Local events should scale linearly with processes
- Send/Receive events should scale faster (more communication opportunities)

### 3. **concurrency_rate.png** - Concurrency Rate vs Number of Processes
Shows the percentage of concurrent event pairs as the system scales.

**Expected behavior:** Should increase with more processes as independent events happen simultaneously in different processes.

### 4. **memory_usage.png** - Lamport vs Vector Clock Memory
Compares the memory usage of Lamport clocks (constant per process) versus Vector clocks (linear with number of processes).

**Expected behavior:**
- Lamport: Linear growth (O(n) where n = processes)
- Vector: Quadratic growth (O(n²) as each of n processes stores n timestamps)

### 5. **message_count.png** - Total Messages vs Number of Processes
Shows how communication scales with the number of processes.

**Expected behavior:** Should increase significantly as more processes create more potential communication paths (O(n²) possible pairs).

### 6. **memory_overhead.png** - Vector Clock Overhead Ratio
Shows the ratio of Vector clock memory to Lamport clock memory.

**Expected behavior:** Should increase linearly with the number of processes, demonstrating that Vector clocks have O(n) space overhead per process compared to Lamport's O(1).

## Configuration

You can modify the following constants in `cmd/plots/main.go`:

```go
const (
    simulationTime = 1 * time.Second  // How long each simulation runs
    localEventProb = 0.3              // Probability of local events (30%)
    sendEventProb  = 0.4              // Probability of send events (40%)
    numRuns        = 3                // Runs per process count (for averaging)
)

var processCounts = []int{2, 4, 6, 8, 10, 15, 20, 25, 30}
```

## Key Insights

These plots help visualize the fundamental trade-offs in distributed systems:

1. **Lamport vs Vector Clocks Trade-off:**
   - Lamport: O(1) space, but cannot detect concurrency
   - Vector: O(n) space, can detect concurrent events

2. **Scalability:**
   - As processes increase, coordination complexity grows
   - Communication overhead scales with O(n²) possible connections

3. **Concurrency:**
   - More processes → more concurrent events
   - Important for understanding system parallelism

## Additional Analyses

### Time Analysis

Analyzes how **simulation duration** affects the system. Run:

```bash
go run cmd/time_analysis/main.go
```

This generates plots showing:
1. **time_event_growth.png** - Linear growth of events over time
2. **time_event_rate.png** - Events per second (steady-state behavior)
3. **time_event_types.png** - How event types accumulate
4. **time_concurrency.png** - Concurrency rate evolution
5. **time_clock_values.png** - Logical time progression
6. **time_efficiency.png** - Normalized throughput

**Key Insights:**
- Shows temporal behavior and steady-state characteristics
- Validates linear growth assumptions
- Demonstrates how logical time relates to physical time

### Probability Analysis

Analyzes how **event probabilities** affect the system. Run:

```bash
go run cmd/prob_analysis/main.go
```

This generates plots showing:
1. **prob_total_events.png** - Activity by probability configuration
2. **prob_event_types.png** - Event type breakdown
3. **prob_concurrency.png** - How probabilities affect concurrency
4. **prob_distribution.png** - Actual vs configured percentages
5. **prob_message_balance.png** - Send/receive ratio validation
6. **prob_efficiency.png** - Communication overhead analysis

**Key Insights:**
- Shows how local vs communication mix affects concurrency
- Demonstrates communication overhead (2× cost: send + receive)
- Reveals different workload patterns (compute-heavy vs communication-heavy)

## Use Cases

These plots are useful for:
- Understanding the scalability characteristics of distributed clock algorithms
- Demonstrating theoretical complexity in practice
- Visualizing the trade-offs between different clock implementations
- Analyzing how system behavior evolves over time
- Understanding steady-state behavior vs initialization effects
- Academic presentations and reports on distributed systems
