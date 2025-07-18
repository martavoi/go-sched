package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	scheduler "go-sched"
	mongostore "go-sched/storage/mongo"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a simple logger
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Connect to MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer client.Disconnect(ctx)

	// Test MongoDB connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Error("failed to ping MongoDB", "error", err)
		os.Exit(1)
	}

	log.Info("connected to MongoDB", "uri", mongoURI)

	// Create MongoDB storage
	db := client.Database("scheduler")
	store := mongostore.NewMongoStore[any](db, "jobs")

	// Add sample jobs with random scheduling
	for i := 1; i <= 50; i++ {
		// Random delay between 1-30 seconds from now
		randomDelay := time.Duration(rand.Intn(30)+1) * time.Second

		job := &scheduler.Job[any]{
			Id:           fmt.Sprintf("job-%d", i),
			Status:       "pending",
			ProcessAfter: time.Now().Add(randomDelay),
			Payload:      nil,
		}

		if err := store.AddJob(job); err != nil {
			log.Error("failed to add job", "job-id", job.Id, "error", err)
		}
	}

	// Create a job handler with random processing time
	jobHandler := func(ctx context.Context, job *scheduler.Job[any]) error {
		// Random processing time between 1-8 seconds
		processingTime := time.Duration(rand.Intn(8)+1) * time.Second

		log.Info("processing job",
			"job-id", job.Id,
			"duration", processingTime)

		// Simulate work with random duration
		time.Sleep(processingTime)

		log.Info("job completed", "job-id", job.Id)
		return nil
	}

	// Create scheduler
	const workerCount = 5
	const interval = 2 * time.Second
	const visibilityTimeout = 30 * time.Second // Jobs auto-recover after 30s if worker crashes

	scheduler := scheduler.NewScheduler(store, workerCount, interval, visibilityTimeout, jobHandler, log)
	done := scheduler.Run(ctx)

	log.Info("scheduler started",
		"workers", workerCount,
		"interval", interval,
		"visibility_timeout", visibilityTimeout,
		"jobs", 50,
		"storage", "mongodb")
	log.Info("press Ctrl+C to stop gracefully")

	// Wait for shutdown signal
	sig := <-sigChan
	log.Info("received shutdown signal", "signal", sig)
	cancel()

	// Wait for graceful shutdown
	<-done
	log.Info("scheduler stopped gracefully")
}
