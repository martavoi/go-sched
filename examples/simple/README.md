# Simple Job Scheduler Example

This example demonstrates the core functionality of the go-sched library with a realistic scenario of concurrent job processing and graceful shutdown.

## What This Example Does

- **Creates 50 jobs** with random scheduling (1-30 seconds delay)
- **5 concurrent workers** process jobs with random durations (1-8 seconds)
- **30-second visibility timeout** for fault tolerance (jobs auto-recover if workers crash)
- **Demonstrates graceful shutdown** with immediate job visibility reset
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

### 3. **Visibility Timeout & Fault Tolerance**
- Jobs become "invisible" for 30 seconds when picked up by workers
- If a worker crashes, jobs automatically become visible again after timeout
- Prevents job loss and enables automatic recovery without manual intervention

### 4. **Graceful Shutdown**
- Scheduler stops accepting new jobs on interrupt signal
- Active workers complete their current jobs before terminating
- Remaining unprocessed jobs are immediately made visible for next restart

## Sample Execution Log

Here's what you'll see when running the example:

```
time=2025-07-18T14:12:07.319+03:00 level=INFO msg="scheduler started" workers=5 interval=2s visibility_timeout=30s jobs=50
time=2025-07-18T14:12:07.319+03:00 level=INFO msg="press Ctrl+C to stop gracefully"

# Jobs start being processed concurrently with visibility timeout protection
time=2025-07-18T14:12:09.319+03:00 level=INFO msg="processing job" job-id=job-46 duration=4s
time=2025-07-18T14:12:09.319+03:00 level=INFO msg="processing job" job-id=job-8 duration=4s
time=2025-07-18T14:12:11.319+03:00 level=INFO msg="processing job" job-id=job-36 duration=4s
time=2025-07-18T14:12:11.319+03:00 level=INFO msg="processing job" job-id=job-17 duration=8s
time=2025-07-18T14:12:11.319+03:00 level=INFO msg="processing job" job-id=job-30 duration=4s

# Jobs complete and new ones are picked up immediately
time=2025-07-18T14:12:13.319+03:00 level=INFO msg="job completed" job-id=job-46
time=2025-07-18T14:12:13.319+03:00 level=INFO msg="job completed" job-id=job-8
time=2025-07-18T14:12:13.319+03:00 level=INFO msg="processing job" job-id=job-2 duration=7s
time=2025-07-18T14:12:13.319+03:00 level=INFO msg="processing job" job-id=job-41 duration=1s
time=2025-07-18T14:12:14.319+03:00 level=INFO msg="job completed" job-id=job-41
time=2025-07-18T14:12:14.319+03:00 level=INFO msg="processing job" job-id=job-28 duration=7s

# User presses Ctrl+C - graceful shutdown with visibility timeout reset
time=2025-07-18T14:12:14.869+03:00 level=INFO msg="received shutdown signal" signal=interrupt
time=2025-07-18T14:12:15.319+03:00 level=INFO msg="job completed" job-id=job-36
time=2025-07-18T14:12:15.319+03:00 level=INFO msg="job completed" job-id=job-30
time=2025-07-18T14:12:15.319+03:00 level=INFO msg="processing job" job-id=job-5 duration=4s
time=2025-07-18T14:12:15.319+03:00 level=INFO msg="shutting down scheduler... making remaining jobs visible" remaining-jobs=2
time=2025-07-18T14:12:15.319+03:00 level=INFO msg="processing job" job-id=job-42 duration=7s

# Workers finish their current jobs before stopping
time=2025-07-18T14:12:19.319+03:00 level=INFO msg="job completed" job-id=job-17
time=2025-07-18T14:12:19.320+03:00 level=INFO msg="job completed" job-id=job-5
time=2025-07-18T14:12:20.319+03:00 level=INFO msg="job completed" job-id=job-2
time=2025-07-18T14:12:21.320+03:00 level=INFO msg="job completed" job-id=job-28
time=2025-07-18T14:12:22.319+03:00 level=INFO msg="job completed" job-id=job-42

# Graceful shutdown completes
time=2025-07-18T14:12:22.319+03:00 level=INFO msg="scheduler shutdown complete"
time=2025-07-18T14:12:22.319+03:00 level=INFO msg="scheduler stopped gracefully"
```

## Log Analysis

### **Startup Phase**
- Scheduler initializes with 5 workers, 2-second fetch interval, and 30-second visibility timeout
- 50 jobs are created with random scheduling delays

### **Processing Phase**
- Jobs are dispatched and automatically become "invisible" for 30 seconds (fault tolerance)
- Workers process jobs concurrently with different durations
- New jobs are automatically picked up as workers become available

### **Shutdown Phase**
1. **Signal Received**: `Ctrl+C` triggers graceful shutdown
2. **Visibility Reset**: 2 remaining jobs immediately made visible for next restart
3. **Worker Completion**: All active workers finished their current jobs
4. **Clean Exit**: Scheduler stopped gracefully with jobs ready for immediate pickup

## Configuration

| Setting | Value | Description |
|---------|-------|-------------|
| **Workers** | 5 | Number of concurrent goroutines |
| **Fetch Interval** | 2s | Pause when no jobs are available |
| **Visibility Timeout** | 30s | Time before failed jobs become visible again |
| **Job Count** | 50 | Total jobs in this example |
| **Job Delay** | 1-30s | Random scheduling delay |
| **Processing Time** | 1-8s | Random job duration |

## Key Takeaways

1. **Efficiency**: Workers are never idle when jobs are available
2. **Concurrency**: Multiple jobs process simultaneously without conflicts  
3. **Fault Tolerance**: 30-second visibility timeout automatically recovers crashed jobs
4. **Storage Resilience**: Exponential backoff retry handles temporary storage failures automatically
5. **Graceful Shutdown**: Remaining jobs immediately available on restart (no 30s delay)
6. **Zero Job Loss**: Jobs cannot be lost due to worker crashes, shutdowns, or storage issues
7. **Observability**: Comprehensive logging shows exactly what's happening

This example demonstrates production-ready job processing with automatic fault recovery, suitable for microservices, background tasks, API calls, data processing, and more. 