package storage

import (
	"fmt"
	"time"

	"go-scheduler"
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
func (s *MemoryStore[T]) FetchPendingJobs(after time.Time, limit int) ([]*scheduler.Job[T], error) {
	entries := make([]*scheduler.Job[T], 0)

	for _, entry := range s.jobs {
		if entry.ProcessAfter.Before(after) && entry.Status == "pending" {
			entries = append(entries, entry)
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

	// Update the existing job with new values
	existingJob.Status = job.Status
	existingJob.ProcessedAt = job.ProcessedAt

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
