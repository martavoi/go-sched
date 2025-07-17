package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// JobHandler defines the function signature for processing jobs
type JobHandler[T any] func(ctx context.Context, job *Job[T]) error

// Scheduler manages the execution of jobs with a typed payload
type Scheduler[T any] struct {
	store       JobStore[T]
	workerCount int
	interval    time.Duration
	log         *slog.Logger
	jobHandler  JobHandler[T]
}

// NewScheduler creates a new scheduler instance
func NewScheduler[T any](store JobStore[T], workerCount int, interval time.Duration, jobHandler JobHandler[T], log *slog.Logger) *Scheduler[T] {
	return &Scheduler[T]{
		store:       store,
		workerCount: workerCount,
		interval:    interval,
		jobHandler:  jobHandler,
		log:         log,
	}
}

// Run starts the scheduler and returns a channel that closes when shutdown is complete
func (s *Scheduler[T]) Run(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		var wg sync.WaitGroup
		jobs := make(chan *Job[T], s.workerCount)

		for i := 0; i < s.workerCount; i++ {
			wg.Add(1)
			go s.worker(ctx, i, jobs, &wg)
		}

		// Demand-driven fetching loop
		for {
			select {
			case <-ctx.Done():
				close(jobs)
				s.log.Info("shutting down scheduler...", "remaining-jobs", len(jobs))

				// Clean up any jobs remaining in channel (mark them back as pending)
				// Do this synchronously to ensure completion before exit
				for remainingJob := range jobs {
					remainingJob.Status = "pending"
					remainingJob.ProcessedAt = nil
					s.store.UpdateJob(remainingJob)
					s.log.Debug("marked unprocessed job as pending", "job-id", remainingJob.Id)
				}

				s.log.Debug("waiting for workers to finish...")
				wg.Wait()
				s.log.Info("scheduler shutdown complete")
				return

			default:
				// Calculate how many jobs we can fetch based on channel capacity
				availableSlots := cap(jobs) - len(jobs)
				if availableSlots > 0 {
					// Fetch jobs to fill available slots
					entries, err := s.store.FetchPendingJobs(time.Now(), availableSlots)
					if err != nil {
						s.log.Error("failed to fetch pending entries", "error", err)
						// Brief pause on error to prevent tight error loop
						time.Sleep(s.interval)
						continue
					}

					if len(entries) == 0 {
						// No jobs available, brief pause to prevent busy waiting
						time.Sleep(s.interval)
						continue
					}

					// Dispatch all fetched jobs
					for _, entry := range entries {
						s.log.Debug("dispatching job", "job-id", entry.Id)
						// Mark as dispatched when putting in channel to prevent duplicate fetching
						entry.Status = "processing"
						err := s.store.UpdateJob(entry)
						if err != nil {
							s.log.Error("failed to mark job as dispatched", "job-id", entry.Id, "error", err)
							continue
						}
						jobs <- entry
					}
				} else {
					time.Sleep(s.interval)
				}
			}
		}
	}()

	return done
}

func (s *Scheduler[T]) worker(ctx context.Context, workerId int, jobs chan *Job[T], wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		startTime := time.Now()
		s.log.Debug("processing job", "job-id", job.Id, "worker-id", workerId)

		err := s.jobHandler(ctx, job)
		duration := time.Since(startTime)

		if err != nil {
			job.Status = "failed"
			now := time.Now()
			job.ProcessedAt = &now
			s.store.UpdateJob(job)
			s.log.Error("failed to process job", "job-id", job.Id, "worker-id", workerId, "duration", duration, "error", err)
		} else {
			job.Status = "completed"
			now := time.Now()
			job.ProcessedAt = &now
			s.store.UpdateJob(job)
			s.log.Debug("job completed", "job-id", job.Id, "worker-id", workerId, "duration", duration)
		}
	}

	s.log.Debug("worker finished", "worker-id", workerId)
}
