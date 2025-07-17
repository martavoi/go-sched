# Simple Job Scheduler Example

This example demonstrates the core functionality of the go-sched library with a realistic scenario of concurrent job processing and graceful shutdown.

## What This Example Does

- **Creates 50 jobs** with random scheduling (1-30 seconds delay)
- **5 concurrent workers** process jobs with random durations (1-8 seconds)
- **Demonstrates graceful shutdown** when receiving interrupt signals
- **Shows demand-driven fetching** - jobs are fetched as workers become available

## Running the Example

```bash
cd examples/simple
go run main.go
```

Press **Ctrl+C** to trigger graceful shutdown and observe the cleanup behavior.

## Key Features Demonstrated

### 1. **Concurrent Job Processing**
- Multiple workers process jobs simultaneously
- Each job has a random processing time (1-8 seconds)
- Workers automatically pick up new jobs as they become available

### 2. **Demand-Driven Job Fetching**
- Jobs are fetched based on worker capacity, not fixed intervals
- No artificial delays when workers are ready for more work
- Efficient utilization of worker goroutines

### 3. **Graceful Shutdown**
- Scheduler stops accepting new jobs on interrupt signal
- Active workers complete their current jobs before terminating
- Remaining unprocessed jobs are cleaned up and marked as "pending"

## Sample Execution Log

Here's what you'll see when running the example:

```
time=2025-07-18T01:28:16.836+03:00 level=INFO msg="scheduler started" workers=5 interval=2s jobs=50
time=2025-07-18T01:28:16.836+03:00 level=INFO msg="press Ctrl+C to stop gracefully"

# Jobs start being processed concurrently
time=2025-07-18T01:28:18.837+03:00 level=INFO msg="processing job" job-id=job-28 duration=2s
time=2025-07-18T01:28:18.837+03:00 level=INFO msg="processing job" job-id=job-24 duration=4s
time=2025-07-18T01:28:20.837+03:00 level=INFO msg="job completed" job-id=job-28
time=2025-07-18T01:28:20.837+03:00 level=INFO msg="processing job" job-id=job-44 duration=5s
time=2025-07-18T01:28:20.837+03:00 level=INFO msg="processing job" job-id=job-36 duration=1s
time=2025-07-18T01:28:20.837+03:00 level=INFO msg="processing job" job-id=job-29 duration=2s
time=2025-07-18T01:28:20.838+03:00 level=INFO msg="processing job" job-id=job-50 duration=5s

# Jobs complete and new ones are picked up immediately
time=2025-07-18T01:28:21.838+03:00 level=INFO msg="job completed" job-id=job-36
time=2025-07-18T01:28:22.837+03:00 level=INFO msg="job completed" job-id=job-24
time=2025-07-18T01:28:22.838+03:00 level=INFO msg="job completed" job-id=job-29
time=2025-07-18T01:28:22.838+03:00 level=INFO msg="processing job" job-id=job-14 duration=8s
time=2025-07-18T01:28:22.838+03:00 level=INFO msg="processing job" job-id=job-7 duration=3s
time=2025-07-18T01:28:22.838+03:00 level=INFO msg="processing job" job-id=job-11 duration=7s

# User presses Ctrl+C - graceful shutdown begins
time=2025-07-18T01:28:25.423+03:00 level=INFO msg="received shutdown signal" signal=interrupt
time=2025-07-18T01:28:25.838+03:00 level=INFO msg="job completed" job-id=job-44
time=2025-07-18T01:28:25.838+03:00 level=INFO msg="processing job" job-id=job-39 duration=7s
time=2025-07-18T01:28:25.838+03:00 level=INFO msg="job completed" job-id=job-50
time=2025-07-18T01:28:25.838+03:00 level=INFO msg="processing job" job-id=job-17 duration=5s
time=2025-07-18T01:28:25.838+03:00 level=INFO msg="job completed" job-id=job-7
time=2025-07-18T01:28:25.838+03:00 level=INFO msg="processing job" job-id=job-34 duration=3s
time=2025-07-18T01:28:26.839+03:00 level=INFO msg="shutting down scheduler..." remaining-jobs=2

# Workers finish their current jobs before stopping
time=2025-07-18T01:28:28.839+03:00 level=INFO msg="job completed" job-id=job-34
time=2025-07-18T01:28:29.838+03:00 level=INFO msg="job completed" job-id=job-11
time=2025-07-18T01:28:30.838+03:00 level=INFO msg="job completed" job-id=job-14
time=2025-07-18T01:28:30.839+03:00 level=INFO msg="job completed" job-id=job-17
time=2025-07-18T01:28:32.838+03:00 level=INFO msg="job completed" job-id=job-39

# Graceful shutdown completes
time=2025-07-18T01:28:32.838+03:00 level=INFO msg="scheduler shutdown complete"
time=2025-07-18T01:28:32.838+03:00 level=INFO msg="scheduler stopped gracefully"
```

## Log Analysis

### **Startup Phase**
- Scheduler initializes with 5 workers and 2-second fetch interval
- 50 jobs are created with random scheduling delays

### **Processing Phase**
- Jobs are dispatched in batches as they become ready
- Workers process jobs concurrently with different durations
- New jobs are automatically dispatched when workers become available

### **Shutdown Phase**
1. **Signal Received**: `Ctrl+C` triggers graceful shutdown
2. **Active Jobs**: 3 jobs were still in the queue when shutdown began
3. **Worker Completion**: All 5 active workers finished their current jobs
4. **Clean Exit**: Scheduler stopped gracefully without killing any jobs

## Configuration

| Setting | Value | Description |
|---------|-------|-------------|
| **Workers** | 5 | Number of concurrent goroutines |
| **Fetch Interval** | 2s | Pause when no jobs are available |
| **Job Count** | 50 | Total jobs in this example |
| **Job Delay** | 1-30s | Random scheduling delay |
| **Processing Time** | 1-8s | Random job duration |

## Key Takeaways

1. **Efficiency**: Workers are never idle when jobs are available
2. **Concurrency**: Multiple jobs process simultaneously without conflicts
3. **Resilience**: Graceful shutdown ensures no jobs are lost or corrupted
4. **Observability**: Comprehensive logging shows exactly what's happening

This example demonstrates real-world job processing patterns suitable for background tasks, API calls, data processing, and more. 