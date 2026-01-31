# Week 2: Go Advanced Patterns

## Goal
Master pointers, clojure, system design in go (uber problem), goroutines, channels, select, sync primitives, and common concurrency patterns used in production systems.

---

## Problem 1: Pipeline Pattern - Data Processing System

### Requirements
Build a multi-stage log processing pipeline that:
1. **Generator Stage**: Reads log lines from multiple sources concurrently
2. **Filter Stage**: Filters logs by level (INFO, ERROR, etc.)
3. **Enrichment Stage**: Enriches logs with metadata using multiple workers
4. **Output Stage**: Writes to multiple outputs (stdout, file, etc.)

### Specifications

**LogEntry Structure**:
- Timestamp
- Level (INFO, ERROR, WARN, DEBUG)
- Message
- Source identifier

**Pipeline Stages**:
1. Generate logs from N sources (fan-out pattern)
2. Filter by log levels
3. Enrich with worker pool (fan-out/fan-in pattern)
4. Output to M destinations (fan-out pattern)

### Concepts to Implement
- ✅ Goroutines for concurrent processing
- ✅ Buffered channels for stage communication
- ✅ `sync.WaitGroup` for synchronization
- ✅ `context.Context` for graceful shutdown
- ✅ Fan-out: multiple goroutines reading from same source
- ✅ Fan-in: multiple goroutines writing to same destination

### Challenges
1. **Graceful Shutdown**: Drain pipeline without data loss
2. **Error Handling**: Add error channel alongside data channel
3. **Backpressure**: Handle slow consumers
4. **Metrics**: Count processed, filtered, and enriched logs

### Test Scenarios
- Pipeline processes all logs correctly
- Graceful shutdown with context cancellation
- No goroutine leaks
- Race detector passes (`go test -race`)

---

## Problem 2: Worker Pool with Rate Limiting

### Requirements
Build a concurrent web scraper that:
1. **Worker Pool**: Fixed number of workers processing jobs
2. **Rate Limiting**: Max requests per second
3. **Retry Logic**: Exponential backoff for failures
4. **Shared State**: Thread-safe statistics tracking
5. **Job Tracking**: Know which jobs are in progress

### Specifications

**Job Structure**:
- ID
- URL to scrape
- Priority (optional extension)
- Retry count

**Result Structure**:
- Job reference
- Content (string)
- Error (if failed)

**Worker Pool Features**:
- Configurable worker count
- Configurable rate limit (requests/second)
- Retry up to N times with exponential backoff
- Track: processed count, failed count, in-progress jobs

### Concepts to Implement
- ✅ Worker pool pattern
- ✅ `sync.Mutex` for shared state
- ✅ `sync.RWMutex` for read-heavy operations
- ✅ `sync.Map` for concurrent tracking
- ✅ `time.Ticker` for rate limiting
- ✅ Buffered channels for job queue
- ✅ Semaphore pattern (optional)

### Challenges
1. **Dynamic Scaling**: Add/remove workers based on queue size
2. **Priority Queue**: Process high-priority jobs first
3. **Circuit Breaker**: Stop processing failing domains
4. **Per-Worker Stats**: Track statistics per worker

### Test Scenarios
- Rate limiter enforces limits correctly
- Retries work with exponential backoff
- Statistics are accurate under concurrent access
- No race conditions

---

## Problem 3: Pub/Sub System with Broadcast

### Requirements
Build a message broker that:
1. **Multiple Publishers**: Any goroutine can publish
2. **Multiple Subscribers**: Subscribe to topics/patterns
3. **Topic Routing**: Route messages based on topics
4. **Slow Consumer Handling**: Don't block fast publishers
5. **Dynamic Subscriptions**: Subscribe/unsubscribe at runtime

### Specifications

**Message Structure**:
- Topic (string)
- Payload (any data)
- Timestamp

**Subscriber**:
- ID
- Subscribed topics (can use wildcards)
- Buffered channel for receiving messages

**Broker Features**:
- Thread-safe subscriber management
- Non-blocking message delivery
- Wildcards in subscriptions (e.g., "orders.*" or "*")
- Auto-cleanup on context cancellation

### Concepts to Implement
- ✅ `sync.RWMutex` for subscriber map
- ✅ Buffered channels for each subscriber
- ✅ Select with default for non-blocking sends
- ✅ Dynamic channel creation/deletion
- ✅ Pattern matching for topics

### Challenges
1. **Wildcard Matching**: Implement "orders.*", "*.created", etc.
2. **Message Persistence**: Buffer missed messages for replay
3. **Acknowledgments**: Require subscribers to ack messages
4. **Health Monitoring**: Detect and remove dead subscribers

### Test Scenarios
- Messages route to correct subscribers
- Slow consumers don't block fast ones
- Wildcard subscriptions work correctly
- No goroutine leaks after unsubscribe

---

## Additional Patterns to Implement

### Pattern 4: Or-Done Channel
**Purpose**: Simplify context cancellation in range loops

**Requirements**:
- Wrap a channel with context awareness
- Return new channel that closes on context cancel or input close
- Useful for clean pipeline stages

### Pattern 5: Tee (Split Channel)
**Purpose**: Duplicate channel output to two destinations

**Requirements**:
- Read from one input channel
- Write to two output channels
- Handle slow consumers on either output
- Close outputs when input closes

### Pattern 6: Semaphore
**Purpose**: Limit concurrent access to resources

**Requirements**:
- Buffered channel with N capacity
- Acquire/Release methods
- Use for rate limiting or resource pooling

### Pattern 7: Error Group
**Purpose**: Run multiple goroutines and collect errors

**Requirements**:
- Start multiple goroutines
- Cancel all if one fails
- Return first error encountered
- Wait for all to complete

Hint: Look at `golang.org/x/sync/errgroup`

---

## Concurrency Patterns Summary

| Pattern | Use Case | Key Primitives |
|---------|----------|----------------|
| **Pipeline** | Multi-stage processing | Channels, goroutines |
| **Fan-Out** | Distribute work | Multiple readers, one channel |
| **Fan-In** | Merge results | Multiple writers, one channel |
| **Worker Pool** | Bounded parallelism | Jobs channel, fixed workers |
| **Pub/Sub** | Broadcasting | Maps, channels, RWMutex |
| **Or-Done** | Cancellation | Context, select |
| **Semaphore** | Resource limiting | Buffered channel |

---

## Testing Concurrency

### Race Detector
**Always** run tests with race detector:
```bash
go test -race ./...
```

### Benchmark Concurrent Code
Create benchmarks to measure performance:
```bash
go test -bench=. -benchmem ./...
```

### Common Testing Patterns

**Test Goroutine Leaks**:
- Use goroutine count before/after tests
- Ensure all goroutines exit

**Test Context Cancellation**:
- Cancel context mid-operation
- Verify graceful shutdown

**Test Race Conditions**:
- Run with `-race` flag
- Use high iteration counts

**Load Testing**:
- Send thousands of concurrent requests
- Verify system remains stable

---

## Project Structure

```
02-concurrency/
├── go.mod
├── pipeline/
│   ├── pipeline.go           # Pipeline implementation
│   ├── pipeline_test.go      # Tests
│   ├── benchmark_test.go     # Benchmarks
│   └── README.md             # Pattern explanation
├── workerpool/
│   ├── pool.go               # Worker pool
│   ├── pool_test.go
│   └── README.md
├── pubsub/
│   ├── broker.go             # Pub/sub broker
│   ├── broker_test.go
│   └── README.md
└── patterns/
    ├── or_done.go            # Or-Done pattern
    ├── tee.go                # Tee pattern
    ├── semaphore.go          # Semaphore
    └── patterns_test.go
```

---

## Implementation Guidelines

### Channel Guidelines
```go
// Buffered vs Unbuffered
unbuffered := make(chan int)      // Blocks until receiver ready
buffered := make(chan int, 100)   // Buffering for performance

// Sending
ch <- value                       // Send (blocks if full)

// Receiving
value := <-ch                     // Receive (blocks if empty)
value, ok := <-ch                 // Check if closed

// Closing
close(ch)                         // Close channel (sender's job)
```

### Select Statement
```go
select {
case msg := <-ch1:
    // Handle message from ch1
case ch2 <- value:
    // Send to ch2
case <-ctx.Done():
    // Context cancelled
case <-time.After(5 * time.Second):
    // Timeout
default:
    // Non-blocking (use carefully!)
}
```

### Context Patterns
```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Check for cancellation
select {
case <-ctx.Done():
    return ctx.Err()  // context.Canceled or context.DeadlineExceeded
default:
    // Continue working
}
```

### WaitGroup Pattern
```go
var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // Do work
    }(i)
}

wg.Wait()  // Block until all Done() called
```

### Mutex Patterns
```go
// Mutex for exclusive access
var mu sync.Mutex
mu.Lock()
// Critical section
mu.Unlock()

// RWMutex for read-heavy workloads
var rwmu sync.RWMutex
rwmu.RLock()   // Multiple readers
// Read data
rwmu.RUnlock()

rwmu.Lock()    // Exclusive writer
// Write data
rwmu.Unlock()
```

---

## Common Pitfalls

### 1. Closing Channels
❌ **Wrong**: Receiver closes channel
```go
// Never close on receiver side!
value, ok := <-ch
if !ok {
    close(ch)  // WRONG!
}
```

✅ **Correct**: Sender closes channel
```go
// Close on sender side after all sends
close(ch)
```

### 2. Loop Variable Capture
❌ **Wrong**: Closing over loop variable
```go
for i := 0; i < 10; i++ {
    go func() {
        fmt.Println(i)  // All see same value!
    }()
}
```

✅ **Correct**: Pass as parameter
```go
for i := 0; i < 10; i++ {
    go func(id int) {
        fmt.Println(id)
    }(i)
}
```

### 3. Forgetting to Wait
❌ **Wrong**: Program exits before goroutines finish
```go
for i := 0; i < 10; i++ {
    go doWork(i)
}
// Program exits immediately!
```

✅ **Correct**: Use WaitGroup
```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        doWork(id)
    }(i)
}
wg.Wait()
```

### 4. Nil Channels
```go
var ch chan int  // nil channel
ch <- 1          // Blocks forever!
<-ch             // Blocks forever!
```

### 5. Deadlocks
❌ **Wrong**: Waiting for each other
```go
ch := make(chan int)
ch <- 1  // Blocks forever (no receiver)
```

✅ **Correct**: Use buffered or goroutine
```go
ch := make(chan int, 1)
ch <- 1  // Doesn't block

// OR
go func() {
    ch <- 1
}()
value := <-ch
```

---

## Debugging Commands

```bash
# Run with race detector (always!)
go test -race ./...

# Run specific test
go test -run TestPipeline ./pipeline/

# Benchmark
go test -bench=BenchmarkWorkerPool -benchtime=10s ./workerpool/

# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Check for goroutine leaks
go test -v -run TestLeaks
```

---

## Success Criteria

By end of Week 2, you should master:
- ✅ Goroutines and channels (buffered & unbuffered)
- ✅ Select statement for multiplexing
- ✅ Context for cancellation and timeouts
- ✅ sync.WaitGroup, sync.Mutex, sync.RWMutex
- ✅ sync.Map for concurrent map access
- ✅ Worker pool pattern
- ✅ Pipeline and fan-out/fan-in patterns
- ✅ Pub/Sub and broadcast patterns
- ✅ Rate limiting techniques
- ✅ Race detection and concurrency testing
- ✅ Common pitfalls and how to avoid them

---

## Resources

- [Go Concurrency Patterns](https://www.youtube.com/watch?v=f6kdp27TYZs) - Rob Pike (Video)
- [Advanced Go Concurrency Patterns](https://www.youtube.com/watch?v=QDDwwePbDtw) - Sameer Ajmani (Video)
- [Concurrency in Go](https://katherine.cox-buday.com/concurrency-in-go/) - Book
- [Go Blog: Concurrency Patterns](https://blog.golang.org/pipelines)
- [Go Blog: Context](https://blog.golang.org/context)

---

## Bonus Challenges

1. **Distributed Rate Limiter**: Implement token bucket algorithm
2. **Priority Queue**: Use heap for job priorities
3. **Circuit Breaker**: Stop calling failing services
4. **Bulkhead Pattern**: Isolate failures
5. **Request Coalescing**: Merge duplicate requests
6. **Caching Layer**: Thread-safe LRU cache
7. **Connection Pool**: Reusable connections with limits

**Target**: Complete all 3 main problems + 2 additional patterns by end of week!
