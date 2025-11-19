# Probability Analysis - Event Probability Impact

This program analyzes how different **event probability configurations** affect system behavior, keeping the number of processes (10) and simulation time (1 second) fixed.

## Running the Analysis

```bash
cd Synchronization
go run cmd/prob_analysis/main.go
```

## Generated Plots

### 1. **prob_total_events.png** - Total Events by Configuration
Shows how different probability combinations affect the total number of events generated.

**Expected behavior:** Higher probabilities = more events (roughly additive).

**Key insight:** Shows the relationship between configured probabilities and actual activity level.

### 2. **prob_event_types.png** - Event Type Breakdown
Compares local, send, and receive events across different configurations.

**Expected behavior:** 
- Higher local probability → more local events
- Higher send probability → more send events (and consequently receive events)

**Key insight:** Demonstrates how probability settings directly control the mix of event types.

### 3. **prob_concurrency.png** - Concurrency Rate by Configuration
Shows how probability settings affect the percentage of concurrent event pairs.

**Expected behavior:** 
- Higher local probability → more concurrency (local events are independent)
- Higher send probability → less concurrency (messages create ordering)

**Key insight:** Communication creates dependencies, reducing concurrency. Local-heavy configurations have more parallelism.

### 4. **prob_distribution.png** - Actual Event Distribution Percentages
Shows the actual percentage breakdown of event types within total events.

**Expected behavior:** Proportions should roughly match configured probabilities (with some variance).

**Key insight:** 
- Validates that the simulator respects configured probabilities
- Shows that receive events always match send events
- Demonstrates the trade-off between local work and communication

### 5. **prob_message_balance.png** - Receive/Send Ratio
Shows the ratio of receive events to send events (should be close to 1.0).

**Expected behavior:** Should hover around 1.0 for all configurations (every send creates a receive).

**Key insight:** 
- Values near 1.0 validate correct message delivery
- Slight variations due to timing and buffering in concurrent execution
- Demonstrates the fundamental property: sends = receives in a closed system

### 6. **prob_efficiency.png** - Communication vs Local Work
Compares total events, local events, and communication events (send + receive).

**Expected behavior:** 
- Communication events should be roughly 2× send probability (send + receive)
- Local events scale independently

**Key insight:** 
- Shows the overhead of communication (2 events per message)
- Helps understand the cost of different communication patterns
- Demonstrates that high-communication configurations generate more total events

## Configuration

The analysis tests these probability combinations:

```go
{0.1, 0.1, "Low Activity (10/10)"}
{0.2, 0.2, "Low-Med (20/20)"}
{0.3, 0.2, "Med Local/Low Send (30/20)"}
{0.2, 0.3, "Low Local/Med Send (20/30)"}
{0.3, 0.3, "Medium (30/30)"}
{0.4, 0.3, "High Local/Med Send (40/30)"}
{0.3, 0.4, "Med Local/High Send (30/40)"}
{0.4, 0.4, "High (40/40)"}
{0.5, 0.3, "Very High Local (50/30)"}
{0.3, 0.5, "Very High Send (30/50)"}
{0.5, 0.5, "Very High Activity (50/50)"}
```

Modify in `cmd/prob_analysis/main.go` to test other combinations.

## Key Questions Answered

### 1. **How do probabilities affect total activity?**
The total events plot shows nearly linear relationship with probability sums.

### 2. **Does local vs communication mix matter?**
Yes! The concurrency plot shows:
- Local-heavy: More concurrent events (processes work independently)
- Communication-heavy: Fewer concurrent events (messages create ordering)

### 3. **What's the cost of communication?**
Communication events plot shows that each message generates 2 events (send + receive), effectively doubling the cost compared to local events.

### 4. **Are probabilities accurately reflected?**
The distribution plot validates that actual event percentages match configured probabilities.

### 5. **Is message delivery reliable?**
Message balance plot confirms that every send results in a receive (ratio ≈ 1.0).

## Practical Implications

**For System Configuration:**
- **Low communication** (high local prob): More parallelism, less coordination
- **High communication** (high send prob): More overhead, more causal dependencies
- **Balanced**: Trade-off between independence and coordination

**For Performance Optimization:**
- Local events are "cheaper" (1 event each)
- Messages cost 2× (send + receive)
- High communication = more total system work

**For Understanding Distributed Systems:**
- Demonstrates the fundamental trade-off between local work and communication
- Shows how communication naturally creates ordering (reduces concurrency)
- Illustrates that messaging overhead is significant

## Design Patterns Revealed

### 1. **Compute-Heavy Pattern** (High Local, Low Send)
- Example: 50% local, 20% send
- More concurrent operations
- Less coordination overhead
- Good for: Independent parallel tasks

### 2. **Communication-Heavy Pattern** (Low Local, High Send)
- Example: 20% local, 50% send
- More causal dependencies
- More coordination, less concurrency
- Good for: Tightly coordinated systems

### 3. **Balanced Pattern** (Equal Local/Send)
- Example: 30% local, 30% send
- Moderate concurrency
- Moderate communication
- Good for: General distributed applications

## Comparison with Other Analyses

**Process Scaling** (`cmd/plots/`):
- Varies number of processes
- Shows spatial scalability
- Fixed probabilities

**Time Analysis** (`cmd/time_analysis/`):
- Varies simulation duration
- Shows temporal behavior
- Fixed probabilities

**Probability Analysis** (this):
- Varies event probabilities
- Shows configuration impact
- Fixed processes and time

Together, these three analyses provide complete understanding of:
- **Spatial** scaling (more nodes)
- **Temporal** behavior (over time)
- **Configurational** impact (different workload patterns)
