# go-sched

A lightweight, concurrent job scheduler library for Go that efficiently processes background tasks using goroutines with fault-tolerant storage operations and graceful shutdown support.

## Features

- **Type-Safe Generics**: Compile-time type safety for job payloads
- **Concurrent Processing**: Configurable number of worker goroutines  
- **Demand-Driven Fetching**: Fetches jobs based on worker capacity, not timers
- **Fault Tolerance**: Visibility timeout ensures zero job loss on worker crashes
- **Exponential Backoff Retry**: Automatic retry with exponential backoff for all storage operations
- **Graceful Shutdown**: Proper cleanup on termination signals
- **I/O Optimized**: Designed for HTTP requests, database operations, and other I/O-bound tasks

## Installation

```bash
go get github.com/martavoi/go-sched
```

## Quick Start

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/martavoi/go-sched"
    "github.com/martavoi/go-sched/storage"
)

func main() {
    // Setup signal handling for graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create storage and logger
    store := storage.NewMemoryStore[any]()
    log := slog.New(slog.NewTextHandler(os.Stdout, nil))
    
    // Add a job
    job := &scheduler.Job[any]{
        Id:           "job-1",
        Status:       "pending",
        ProcessAfter: time.Now(),
        Payload:      nil,
    }
    store.AddJob(job)
    
    // Create job handler
    jobHandler := func(ctx context.Context, job *scheduler.Job[any]) error {
        log.Info("processing job", "job-id", job.Id)
        time.Sleep(1 * time.Second) // Simulate work
        log.Info("job completed", "job-id", job.Id)
        return nil
    }
    
    // Configure and start scheduler
    const workerCount = 2
    const fetchInterval = 1 * time.Second
    const visibilityTimeout = 30 * time.Second // Jobs auto-recover after 30s if worker crashes
    
    scheduler := scheduler.NewScheduler(store, workerCount, fetchInterval, visibilityTimeout, jobHandler, log)
    done := scheduler.Run(ctx)
    
    log.Info("scheduler started - press Ctrl+C to stop gracefully")
    
    // Wait for shutdown signal
    sig := <-sigChan
    log.Info("received shutdown signal", "signal", sig)
    cancel()
    
    // Wait for graceful shutdown
    <-done
    log.Info("scheduler stopped gracefully")
}
```

## Library Structure

```
go-sched/
├── scheduler.go        # Main scheduler implementation
├── job.go             # Job types and interfaces
├── storage/           # Storage implementations
│   ├── memory.go     # In-memory store (for development/testing)
│   └── mongo/        # MongoDB store (for production)
└── examples/         # Usage examples
    ├── simple/       # In-memory example
    └── mongo/        # MongoDB example
```

## Storage Implementations

### Memory Store (Included)

Perfect for development, testing, and small-scale applications:

```go
store := storage.NewMemoryStore[YourPayloadType]()
```

### MongoDB Store (Included)

Production-ready persistent storage with MongoDB:

```go
import mongostore "go-sched/storage/mongo"

client, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
db := client.Database("scheduler")
store := mongostore.NewMongoStore[YourPayloadType](db, "jobs")
```

### Custom Storage

Implement the `JobStore` interface for your database:

```go
type JobStore[T any] interface {
    FetchPendingJobs(after time.Time, limit int, visibilityTimeout time.Duration) ([]*Job[T], error)
    UpdateJob(job *Job[T]) error  
    AddJob(job *Job[T]) error
}
```

Examples: PostgreSQL, Redis, etc.

## Running Examples

### Memory Store Example
```bash
cd examples/simple
go run main.go
```

### MongoDB Example
```bash
# Start MongoDB first
docker run -d -p 27017:27017 --name mongo-scheduler mongo:latest

# Run example
cd examples/mongo
go run main.go
```

## Configuration

| Parameter | Description | Recommended Value |
|-----------|-------------|-------------------|
| `workerCount` | Number of concurrent goroutines | CPU_cores × 10-20 for I/O-bound |
| `fetchInterval` | Pause when no jobs available | 1-5 seconds |
| `visibilityTimeout` | Time before failed jobs become visible again | 30 seconds - 5 minutes |

### **Retry Policy (Built-in)**
The scheduler automatically uses exponential backoff for all storage operations:
- **Initial Delay**: 100ms
- **Max Delay**: 5 seconds  
- **Max Retries**: Continues until context cancellation
- **Backoff Factor**: 2x (exponential growth)

No configuration needed - works out of the box for maximum reliability.

## Performance Tuning

### For I/O-Bound Jobs
I/O-bound jobs (HTTP requests, database queries) typically use **~5% CPU per worker**:

- **1 CPU core** → ~20 concurrent workers
- **4 CPU cores** → ~80 concurrent workers  
- **8 CPU cores** → ~160 concurrent workers

**Formula**: `workerCount = CPU_cores × 20`

### For CPU-Bound Jobs
CPU-intensive jobs should match core count:

**Formula**: `workerCount = CPU_cores`

## Fault Tolerance & Graceful Shutdown

The scheduler provides automatic fault recovery and graceful shutdown:

### **Visibility Timeout (Fault Tolerance)**
- Jobs become "invisible" when picked up by workers (default 30s)
- If worker crashes, jobs automatically become visible again after timeout
- Prevents job loss and enables automatic recovery

### **Exponential Backoff Retry Policy**
All storage operations are protected with exponential backoff retry:
- **Job Status Updates**: Automatic retry when marking jobs as completed/failed
- **Visibility Changes**: Retry when making jobs visible/invisible
- **Shutdown Operations**: Resilient cleanup of remaining jobs
- **Self-Recovery**: Temporary storage failures don't stop the scheduler

**Benefits:**
- **Network Resilience**: Handles temporary network issues with databases
- **Database Failover**: Survives brief database connection drops
- **Resource Contention**: Gracefully handles storage system overload
- **Zero Data Loss**: Jobs are never lost due to transient storage failures

### **Graceful Shutdown**
The scheduler handles `SIGTERM` and `SIGINT` signals:

1. Stops fetching new jobs
2. Waits for active workers to complete current jobs  
3. Makes remaining unprocessed jobs immediately visible (with retry)
4. Exits cleanly

Perfect for containerized environments (Docker, Kubernetes).

## Type Safety

The scheduler uses **Go generics** for compile-time type safety:

```go
// ✅ Simple payload (any type)
store := storage.NewMemoryStore[any]()

// ✅ Custom payload type
type EmailJob struct {
    UserID string
    Email  string
}
store := storage.NewMemoryStore[EmailJob]()

// ✅ Type-safe payload access (no type assertions!)
func handler(ctx context.Context, job *scheduler.Job[EmailJob]) error {
    email := job.Payload.Email    // Direct access, compile-time verified
    userID := job.Payload.UserID  // No runtime type checking needed
    return sendEmail(email, userID)
}
```

## License

Apache License 2.0