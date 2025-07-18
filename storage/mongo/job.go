package mongo

import "time"

type Job[T any] struct {
	Id           string     `bson:"_id"`
	Status       string     `bson:"status"`
	ProcessAfter time.Time  `bson:"processAfter"`
	VisibleAfter *time.Time `bson:"visibleAfter,omitempty"`
	ProcessedAt  *time.Time `bson:"processedAt,omitempty"`
	Payload      T          `bson:"payload"`
}
