package couchbase

import (
	"context"
	"errors"
	"fmt"
	"time"

	scheduler "go-sched"

	"github.com/couchbase/gocb/v2"
)

type CouchbaseStore[T any] struct {
	bucket         *gocb.Bucket
	scopeName      string
	collectionName string
}

// NewCouchbaseStore creates a store with custom scope and collection (Couchbase 7.0+)
func NewCouchbaseStore[T any](bucket *gocb.Bucket, scopeName, collectionName string) *CouchbaseStore[T] {
	return &CouchbaseStore[T]{
		bucket:         bucket,
		scopeName:      scopeName,
		collectionName: collectionName,
	}
}

func (s *CouchbaseStore[T]) FetchPendingJobs(after time.Time, limit int, visibilityTimeout time.Duration) ([]*scheduler.Job[T], error) {
	// N1QL query to find pending and visible jobs
	query := fmt.Sprintf(`
		SELECT id, status, processAfter, visibleAfter, processedAt, payload
		FROM %s
		WHERE status = $status 
		AND processAfter < $after
		AND (visibleAfter IS MISSING OR visibleAfter IS NULL OR visibleAfter < $now)
		ORDER BY processAfter ASC
		LIMIT $limit`, "`"+s.collectionName+"`")

	options := &gocb.QueryOptions{
		NamedParameters: map[string]interface{}{
			"status": "pending",
			"after":  after,
			"now":    time.Now(),
			"limit":  limit,
		},
	}

	result, err := s.bucket.Scope(s.scopeName).Query(query, options)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	var jobs []*scheduler.Job[T]
	for result.Next() {
		var job Job[T]
		if err := result.Row(&job); err != nil {
			return nil, err
		}

		// Convert to scheduler.Job
		jobs = append(jobs, &scheduler.Job[T]{
			Id:           job.Id,
			Status:       job.Status,
			ProcessAfter: job.ProcessAfter,
			VisibleAfter: job.VisibleAfter,
			ProcessedAt:  job.ProcessedAt,
			Payload:      job.Payload,
		})
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (s *CouchbaseStore[T]) UpdateJob(job *scheduler.Job[T]) error {
	if job.Id == "" {
		return errors.New("job Id cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	jobDoc := Job[T]{
		Id:           job.Id,
		Status:       job.Status,
		ProcessAfter: job.ProcessAfter,
		VisibleAfter: job.VisibleAfter,
		ProcessedAt:  job.ProcessedAt,
		Payload:      job.Payload,
	}

	collection := s.bucket.Scope(s.scopeName).Collection(s.collectionName)
	_, err := collection.Replace(job.Id, jobDoc, &gocb.ReplaceOptions{
		Context: ctx,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *CouchbaseStore[T]) AddJob(job *scheduler.Job[T]) error {
	if job.Id == "" {
		return errors.New("job Id cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	jobDoc := Job[T]{
		Id:           job.Id,
		Status:       job.Status,
		ProcessAfter: job.ProcessAfter,
		VisibleAfter: job.VisibleAfter,
		ProcessedAt:  job.ProcessedAt,
		Payload:      job.Payload,
	}

	collection := s.bucket.Scope(s.scopeName).Collection(s.collectionName)
	_, err := collection.Insert(job.Id, jobDoc, &gocb.InsertOptions{
		Context: ctx,
	})
	if err != nil {
		return err
	}

	return nil
}
