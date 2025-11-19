# Complete Analysis Suite - Distributed Clock Simulation

This directory contains three complementary analysis programs that visualize different aspects of distributed system behavior.

## Quick Start

Run all three analyses:

```bash
cd Synchronization

# Process scaling analysis
go run cmd/plots/main.go

# Time analysis  
go run cmd/time_analysis/main.go

# Probability analysis
go run cmd/prob_analysis/main.go
```

## Three Dimensions of Analysis

### 📊 1. Process Scaling Analysis (`cmd/plots/`)

**What it varies:** Number of processes (2 → 30)  
**What's fixed:** Time (1s), Probabilities (30% local, 40% send)

**Answers:**
- How does the system scale with more nodes?
- What's the memory overhead of Vector vs Lamport clocks?
- How does communication complexity grow (O(n²))?

**6 Plots Generated:**
- `event_count.png` - Total events vs processes
- `event_types.png` - Event type breakdown
- `concurrency_rate.png` - Concurrency percentage
- `memory_usage.png` - Lamport vs Vector memory (O(n) vs O(n²))
- `message_count.png` - Communication growth
- `memory_overhead.png` - Vector/Lamport ratio

**Key Insight:** Shows **spatial scalability** - adding more nodes increases complexity quadratically for communication and memory.

---

### ⏱️ 2. Time Analysis (`cmd/time_analysis/`)

**What it varies:** Simulation duration (100ms → 5s)  
**What's fixed:** Processes (10), Probabilities (30% local, 40% send)

**Answers:**
- Does the system reach steady state?
- How does logical time relate to physical time?
- Is event generation consistent over time?

**6 Plots Generated:**
- `time_event_growth.png` - Event accumulation (linear)
- `time_event_rate.png` - Events/second (steady state)
- `time_event_types.png` - Type accumulation
- `time_concurrency.png` - Concurrency evolution
- `time_clock_values.png` - Logical time progression
- `time_efficiency.png` - Normalized throughput

**Key Insight:** Shows **temporal behavior** - the system operates consistently over time with predictable event generation rates.

---

### 🎲 3. Probability Analysis (`cmd/prob_analysis/`)

**What it varies:** Event probabilities (11 different configurations)  
**What's fixed:** Processes (10), Time (1s)

**Answers:**
- How do local vs communication probabilities affect behavior?
- What's the overhead of communication?
- Which configurations maximize concurrency?

**6 Plots Generated:**
- `prob_total_events.png` - Activity by configuration
- `prob_event_types.png` - Type breakdown
- `prob_concurrency.png` - Concurrency by config
- `prob_distribution.png` - Actual percentages
- `prob_message_balance.png` - Send/receive validation
- `prob_efficiency.png` - Communication overhead

**Key Insight:** Shows **configurational impact** - local-heavy configurations have more concurrency, communication-heavy have more coordination.

---

## Summary Table

| Analysis | Variable | Fixed | Key Metric | Main Finding |
|----------|----------|-------|------------|--------------|
| **Process Scaling** | # Processes | Time, Probs | Memory Growth | O(n²) for vectors |
| **Time** | Duration | Processes, Probs | Event Rate | Linear, stable |
| **Probability** | Event Probs | Processes, Time | Concurrency | Local ↑ = Concurrent ↑ |

## What Each Analysis Reveals

### Process Scaling → **Scalability**
- Can the system handle more nodes?
- What are the resource costs?
- How does complexity grow?

### Time Analysis → **Stability**
- Does the system behave consistently?
- Is performance predictable?
- Are there initialization effects?

### Probability Analysis → **Workload Patterns**
- How do different workloads behave?
- What's the cost of communication?
- How to optimize for concurrency?

## Complete Understanding

Running all three analyses provides:

1. **Horizontal Scaling** (Process): Adding more nodes
2. **Vertical Scaling** (Time): Running longer
3. **Workload Tuning** (Probability): Different usage patterns

Together, they answer:
- ✅ How the system scales spatially (more processes)
- ✅ How the system scales temporally (longer runs)
- ✅ How configuration affects behavior (different workloads)

## Use Cases

**Academic Research:**
- Demonstrate distributed systems concepts
- Validate theoretical complexity (O(n), O(n²))
- Compare Lamport vs Vector clocks empirically

**System Design:**
- Choose appropriate clock algorithm for scale
- Understand communication patterns
- Plan capacity and resources

**Performance Analysis:**
- Identify bottlenecks
- Optimize for specific workloads
- Predict behavior at scale

## Configuration

Each analysis can be customized:

**Process Scaling:**
```go
var processCounts = []int{2, 4, 6, 8, 10, 15, 20, 25, 30}
```

**Time Analysis:**
```go
var simulationDurations = []int{100, 250, 500, 750, 1000, 1500, 2000, 3000, 5000}
```

**Probability Analysis:**
```go
var probConfigs = []ProbConfig{
    {0.1, 0.1, "Low Activity"},
    {0.3, 0.3, "Medium"},
    {0.5, 0.5, "High Activity"},
    // ... add more
}
```

## Requirements

```bash
go get gonum.org/v1/plot@latest
```

## Output

All plots are generated as PNG files in the Synchronization directory, ready for:
- Academic papers
- Presentations
- Technical documentation
- Performance reports

## Total Plots Generated

**18 plots** covering:
- 6 process scaling plots
- 6 time analysis plots  
- 6 probability analysis plots

All plots are publication-quality with:
- Clear labels and legends
- Grid lines for readability
- Color-coded data series
- Appropriate scales and units
