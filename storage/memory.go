package storage

import (
	"fmt"
	"time"

	scheduler "go-sched"
)

// MemoryStore is an in-memory implementation of JobStore for testing and development
type MemoryStore[T any] struct {
	jobs map[string]*scheduler.Job[T]
}

// NewMemoryStore creates a new in-memory job store
func NewMemoryStore[T any]() *MemoryStore[T] {
	return &MemoryStore[T]{
		jobs: make(map[string]*scheduler.Job[T]),
	}
}

// FetchPendingJobs retrieves pending jobs that are ready to be processed
// Sets visibility timeout on fetched jobs to mark them as being processed
func (s *MemoryStore[T]) FetchPendingJobs(after time.Time, limit int, visibilityTimeout time.Duration) ([]*scheduler.Job[T], error) {
	entries := make([]*scheduler.Job[T], 0)

	for _, job := range s.jobs {
		// Only fetch jobs that are pending, ready to run, and visible
		if job.Status == "pending" &&
			job.ProcessAfter.Before(after) &&
			job.IsVisible() {

			entries = append(entries, job)
		}

		if len(entries) >= limit {
			break
		}
	}

	return entries, nil
}

// UpdateJob updates an existing job's status and processing timestamp
func (s *MemoryStore[T]) UpdateJob(job *scheduler.Job[T]) error {
	existingJob, ok := s.jobs[job.Id]
	if !ok {
		return fmt.Errorf("job not found: %s", job.Id)
	}

	// Update fields
	existingJob.Status = job.Status
	existingJob.ProcessedAt = job.ProcessedAt
	existingJob.VisibleAfter = job.VisibleAfter

	return nil
}

// AddJob adds a new job to the store
func (s *MemoryStore[T]) AddJob(job *scheduler.Job[T]) error {
	if _, exists := s.jobs[job.Id]; exists {
		return fmt.Errorf("job already exists: %s", job.Id)
	}

	s.jobs[job.Id] = job
	return nil
}

// GetJobs returns all jobs (for debugging/testing)
func (s *MemoryStore[T]) GetJobs() map[string]*scheduler.Job[T] {
	result := make(map[string]*scheduler.Job[T])
	for k, v := range s.jobs {
		result[k] = v
	}
	return result
}
