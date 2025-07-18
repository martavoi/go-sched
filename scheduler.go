package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// JobHandler defines the function signature for processing jobs
// Jobs are passed by value to prevent accidental modifications
type JobHandler[T any] func(ctx context.Context, job Job[T]) error

// Scheduler manages the execution of jobs with a typed payload
type Scheduler[T any] struct {
	store             JobStore[T]
	workerCount       int
	interval          time.Duration
	visibilityTimeout time.Duration
	log               *slog.Logger
	jobHandler        JobHandler[T]
}

// NewScheduler creates a new scheduler instance with visibility timeout
func NewScheduler[T any](store JobStore[T], workerCount int, interval time.Duration, visibilityTimeout time.Duration, jobHandler JobHandler[T], log *slog.Logger) *Scheduler[T] {
	return &Scheduler[T]{
		store:             store,
		workerCount:       workerCount,
		interval:          interval,
		visibilityTimeout: visibilityTimeout,
		jobHandler:        jobHandler,
		log:               log,
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
				s.log.Info("shutting down scheduler... making remaining jobs visible", "remaining-jobs", len(jobs))
				// Graceful cleanup: make remaining jobs immediately visible
				for remainingJob := range jobs {
					remainingJob.MakeVisible()
					_, err := backoff.Retry(ctx, func() (any, error) {
						err := s.store.UpdateJob(remainingJob)
						return nil, err
					}, backoff.WithNotify(func(err error, d time.Duration) {
						s.log.Error("failed to make unprocessed job visible, retrying...", "job-id", remainingJob.Id, "error", err, "duration", d)
					}))
					if err != nil {
						s.log.Error("failed to make unprocessed job visible after retries", "job-id", remainingJob.Id, "error", err)
					}

					s.log.Debug("made unprocessed job visible", "job-id", remainingJob.Id)
				}
				wg.Wait()
				s.log.Info("scheduler shutdown complete")
				return

			default:
				// Calculate how many jobs we can fetch based on channel capacity
				availableSlots := cap(jobs) - len(jobs)
				if availableSlots > 0 {
					// Fetch jobs to fill available slots
					entries, err := backoff.Retry(ctx, func() ([]*Job[T], error) {
						return s.store.FetchPendingJobs(time.Now(), availableSlots, s.visibilityTimeout)
					}, backoff.WithNotify(func(err error, d time.Duration) {
						s.log.Error("failed to fetch pending entries, retrying...", "error", err, "duration", d)
					}))
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

					// Make jobs invisible and dispatch them
					for _, entry := range entries {
						s.log.Debug("making job invisible", "job-id", entry.Id)
						entry.MakeInvisible(s.visibilityTimeout)
						_, err := backoff.Retry(ctx, func() (any, error) {
							err := s.store.UpdateJob(entry)
							return nil, err
						}, backoff.WithNotify(func(err error, d time.Duration) {
							s.log.Error("failed to make job invisible, retrying...", "job-id", entry.Id, "error", err, "duration", d)
						}))
						if err != nil {
							s.log.Error("failed to make job invisible after retries", "job-id", entry.Id, "error", err)
						}

						s.log.Debug("dispatching job", "job-id", entry.Id)
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

		// Pass job by value to prevent modifications
		err := s.jobHandler(ctx, *job)
		duration := time.Since(startTime)

		// Update job status based on result
		if err != nil {
			s.log.Info("failed to process job", "job-id", job.Id, "worker-id", workerId, "duration", fmt.Sprintf("%.2fs", duration.Seconds()), "error", err)
			job.MakeFailed()
		} else {
			s.log.Info("job completed", "job-id", job.Id, "worker-id", workerId, "duration", fmt.Sprintf("%.2fs", duration.Seconds()))
			job.MakeCompleted()
		}

		// Update job with retry logic
		_, updateErr := backoff.Retry(ctx, func() (any, error) {
			err := s.store.UpdateJob(job)
			return nil, err
		}, backoff.WithNotify(func(err error, d time.Duration) {
			s.log.Error("failed to update job, retrying...", "job-id", job.Id, "error", err, "duration", d)
		}))

		if updateErr != nil {
			s.log.Error("failed to update job after retries", "job-id", job.Id, "error", updateErr)
		}
	}

	s.log.Debug("worker finished", "worker-id", workerId)
}
