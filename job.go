package scheduler

import "time"

// Job represents a scheduled job with a typed payload
type Job[T any] struct {
	Id           string     `json:"id"`
	Status       string     `json:"status"`                 // "pending" or "completed"
	ProcessAfter time.Time  `json:"processAfter"`           // When job should be processed
	VisibleAfter *time.Time `json:"visibleAfter,omitempty"` // When job becomes visible again (visibility timeout)
	ProcessedAt  *time.Time `json:"processedAt,omitempty"`  // When job was completed
	Payload      T          `json:"payload"`
}

// IsVisible returns true if the job is currently visible (can be picked up by workers)
func (j *Job[T]) IsVisible() bool {
	if j.Status != "pending" {
		return false
	}
	if j.VisibleAfter == nil {
		return true
	}
	return time.Now().After(*j.VisibleAfter)
}

// MakeInvisible sets the visibility timeout for the job (marks as being processed)
func (j *Job[T]) MakeInvisible(visibilityTimeout time.Duration) {
	visibleAfter := time.Now().Add(visibilityTimeout)
	j.VisibleAfter = &visibleAfter
}

// MakeVisible clears the visibility timeout (makes job available again)
func (j *Job[T]) MakeVisible() {
	j.VisibleAfter = nil
}

// MakeFailed marks the job as failed and makes it visible again
func (j *Job[T]) MakeFailed() {
	j.Status = "failed"
	j.MakeVisible()
}

// MakeCompleted marks the job as completed and makes it visible again
func (j *Job[T]) MakeCompleted() {
	j.Status = "completed"
	now := time.Now()
	j.ProcessedAt = &now
	j.MakeVisible()
}

// JobStore defines the interface for job persistence
type JobStore[T any] interface {
	// FetchPendingJobs retrieves pending jobs that are ready to be processed
	// Jobs returned will have their visibility timeout set
	FetchPendingJobs(after time.Time, limit int, visibilityTimeout time.Duration) ([]*Job[T], error)

	// UpdateJob updates an existing job's status and processing timestamp
	UpdateJob(job *Job[T]) error

	// AddJob adds a new job to the store
	AddJob(job *Job[T]) error
}
