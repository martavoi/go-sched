package storage

import (
	"fmt"
	"go-scheduler/internal/scheduler"
	"time"
)

type InMemStorage struct {
	jobs map[string]*scheduler.JobEntry
}

func NewInMemStorage() *InMemStorage {
	return &InMemStorage{jobs: make(map[string]*scheduler.JobEntry)}
}

func (s *InMemStorage) FetchPendingEntries(after time.Time, limit int) ([]*scheduler.JobEntry, error) {
	entries := make([]*scheduler.JobEntry, 0)

	for _, entry := range s.jobs {
		if entry.ProcessAt.After(after) && entry.Status == "pending" {
			entries = append(entries, entry)
		}

		if len(entries) >= limit {
			break
		}
	}

	return entries, nil
}

func (s *InMemStorage) MarkEntryAsProcessing(id string) error {
	entry, ok := s.jobs[id]
	if !ok {
		return fmt.Errorf("entry not found")
	}

	entry.Status = "processing"
	return nil
}

func (s *InMemStorage) MarkEntryAsCompleted(id string) error {
	entry, ok := s.jobs[id]
	if !ok {
		return fmt.Errorf("entry not found")
	}

	entry.Status = "completed"
	return nil
}

func (s *InMemStorage) MarkEntryAsPending(id string) error {
	entry, ok := s.jobs[id]
	if !ok {
		return fmt.Errorf("entry not found")
	}

	entry.Status = "pending"
	return nil
}

func (s *InMemStorage) MarkEntryAsFailed(id string) error {
	entry, ok := s.jobs[id]
	if !ok {
		return fmt.Errorf("entry not found")
	}

	entry.Status = "failed"
	return nil
}

func (s *InMemStorage) AddEntry(entry *scheduler.JobEntry) error {
	s.jobs[entry.Id] = entry
	return nil
}
