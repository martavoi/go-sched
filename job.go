package scheduler

import "time"

// Job represents a scheduled job with a typed payload
type Job[T any] struct {
	Id           string     `json:"id"`
	Status       string     `json:"status"`
	ProcessAfter time.Time  `json:"processAfter"`
	ProcessedAt  *time.Time `json:"processedAt,omitempty"`
	Payload      T          `json:"payload"`
}

// JobStore defines the interface for job persistence
type JobStore[T any] interface {
	// FetchPendingJobs retrieves pending jobs that are ready to be processed
	FetchPendingJobs(after time.Time, limit int) ([]*Job[T], error)

	// UpdateJob updates an existing job's status and processing timestamp
	UpdateJob(job *Job[T]) error

	// AddJob adds a new job to the store
	AddJob(job *Job[T]) error
}
