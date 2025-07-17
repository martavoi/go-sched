package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type JobEntry struct {
	Id        string
	Status    string
	ProcessAt time.Time
}

type Repository interface {
	FetchPendingEntries(after time.Time, limit int) ([]*JobEntry, error)
	MarkEntryAsProcessing(id string) error
	MarkEntryAsCompleted(id string) error
	MarkEntryAsPending(id string) error
	MarkEntryAsFailed(id string) error
	AddEntry(entry *JobEntry) error
}

type JobHandler func(ctx context.Context, job *JobEntry) error

type Scheduler struct {
	repo        Repository
	workerCount int
	interval    time.Duration
	log         *slog.Logger
	jobHandler  JobHandler
}

func NewScheduler(repo Repository, workerCount int, interval time.Duration, jobHandler JobHandler, log *slog.Logger) *Scheduler {
	return &Scheduler{
		repo:        repo,
		workerCount: workerCount,
		interval:    interval,
		jobHandler:  jobHandler,
		log:         log,
	}
}

func (s *Scheduler) Run(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		var wg sync.WaitGroup
		jobs := make(chan *JobEntry, s.workerCount)

		for i := 0; i < s.workerCount; i++ {
			wg.Add(1)
			go s.worker(ctx, jobs, &wg)
		}

		var ticker = time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				entries, err := s.repo.FetchPendingEntries(time.Now(), s.workerCount)
				if err != nil {
					s.log.Error("failed to fetch pending entries", "error", err)
					continue
				}

				for _, entry := range entries {
					s.log.Info("dispatching job", "job-id", entry.Id)
					// Mark as dispatched when putting in channel to prevent duplicate fetching
					err := s.repo.MarkEntryAsProcessing(entry.Id)
					if err != nil {
						s.log.Error("failed to mark job as dispatched", "job-id", entry.Id, "error", err)
						continue
					}
					jobs <- entry
				}

			case <-ctx.Done():
				s.log.Info("shutting down scheduler...")
				close(jobs)

				// Clean up any jobs remaining in channel (mark them back as pending)
				// Do this synchronously to ensure completion before exit
				for remainingJob := range jobs {
					s.repo.MarkEntryAsPending(remainingJob.Id)
					s.log.Info("marked unprocessed job as pending", "job-id", remainingJob.Id)
				}

				s.log.Info("waiting for workers to finish...")
				wg.Wait()
				s.log.Info("scheduler shutdown complete")
				return
			}
		}
	}()

	return done
}

func (s *Scheduler) worker(ctx context.Context, jobs chan *JobEntry, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		s.log.Info("processing job", "job-id", job.Id)

		err := s.jobHandler(ctx, job)
		if err != nil {
			s.repo.MarkEntryAsFailed(job.Id)
			s.log.Error("failed to process job", "job-id", job.Id, "error", err)
		} else {
			s.repo.MarkEntryAsCompleted(job.Id)
			s.log.Info("job completed", "job-id", job.Id)
		}
	}

	s.log.Info("worker finished")
}
