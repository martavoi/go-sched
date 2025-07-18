package couchbase

import "time"

type Job[T any] struct {
	Id           string     `json:"id"`
	Status       string     `json:"status"`
	ProcessAfter time.Time  `json:"processAfter"`
	VisibleAfter *time.Time `json:"visibleAfter,omitempty"`
	ProcessedAt  *time.Time `json:"processedAt,omitempty"`
	Payload      T          `json:"payload"`
}
