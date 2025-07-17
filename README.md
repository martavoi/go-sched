# go-scheduler

A lightweight, concurrent job scheduler for Go that efficiently processes background tasks using goroutines with graceful shutdown support.

## Features

- **Concurrent Processing**: Configurable number of worker goroutines
- **Rate-Limited Fetching**: Controlled job retrieval from storage
- **Graceful Shutdown**: Proper cleanup on termination signals
- **I/O Optimized**: Designed for HTTP requests, database operations, and other I/O-bound tasks

## Quick Start

```go
// Configure workers and fetch interval
const workerCount = 20
const interval = 3 * time.Second

// Create job handler
jobHandler := func(ctx context.Context, job *scheduler.JobEntry) error {
    // Your job logic here (HTTP requests, DB operations, etc.)
    return nil
}

// Start scheduler
storage := storage.NewInMemStorage()
scheduler := scheduler.NewScheduler(storage, workerCount, interval, jobHandler, logger)

done := scheduler.Run(ctx)
<-done // Wait for graceful shutdown
```

## Configuration

| Parameter | Description | Typical Value |
|-----------|-------------|---------------|
| `workerCount` | Number of concurrent goroutines | 20-80 |
| `interval` | Job fetch frequency | 2-5 seconds |

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

## Graceful Shutdown

The scheduler handles `SIGTERM` and `SIGINT` signals:

1. Stops fetching new jobs
2. Waits for active workers to complete
3. Cleans up remaining jobs in queue
4. Exits cleanly

Perfect for containerized environments (Docker, Kubernetes).
