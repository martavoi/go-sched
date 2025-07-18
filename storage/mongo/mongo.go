package mongo

import (
	"context"
	"errors"
	"time"

	scheduler "go-sched"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore[T any] struct {
	db      *mongo.Database
	colName string
}

func NewMongoStore[T any](db *mongo.Database, colName string) *MongoStore[T] {
	return &MongoStore[T]{
		db:      db,
		colName: colName,
	}
}

func (s *MongoStore[T]) FetchPendingJobs(after time.Time, limit int, visibilityTimeout time.Duration) ([]*scheduler.Job[T], error) {
	collection := s.db.Collection(s.colName)

	filter := bson.M{
		"status":       "pending",
		"processAfter": bson.M{"$lt": after},
		"$or": []bson.M{
			{"visibleAfter": bson.M{"$exists": false}},
			{"visibleAfter": nil},
			{"visibleAfter": bson.M{"$lt": time.Now()}},
		},
	}

	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	jobs := make([]*scheduler.Job[T], 0)

	for cursor.Next(ctx) {
		var job Job[T]
		if err := cursor.Decode(&job); err != nil {
			return nil, err
		}

		jobs = append(jobs, &scheduler.Job[T]{
			Id:           job.Id,
			Status:       job.Status,
			ProcessAfter: job.ProcessAfter,
			VisibleAfter: job.VisibleAfter,
			ProcessedAt:  job.ProcessedAt,
			Payload:      job.Payload,
		})
	}

	return jobs, nil
}

func (s *MongoStore[T]) UpdateJob(job *scheduler.Job[T]) error {
	if job.Id == "" {
		return errors.New("job Id cannot be empty")
	}

	collection := s.db.Collection(s.colName)

	filter := bson.M{"_id": job.Id}

	update := bson.M{
		"$set": bson.M{
			"status":       job.Status,
			"visibleAfter": job.VisibleAfter,
			"processedAt":  job.ProcessedAt,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *MongoStore[T]) AddJob(job *scheduler.Job[T]) error {
	if job.Id == "" {
		return errors.New("job Id cannot be empty")
	}

	collection := s.db.Collection(s.colName)

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

	_, err := collection.InsertOne(ctx, jobDoc)
	if err != nil {
		return err
	}

	return nil
}
