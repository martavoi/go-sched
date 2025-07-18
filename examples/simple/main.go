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
	"go-sched/storage"
)

func main() {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a simple logger
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create storage and add jobs
	store := storage.NewMemoryStore[any]()

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
		store.AddJob(job)
	}

	// Create a job handler with random processing time
	jobHandler := func(ctx context.Context, job *scheduler.Job[any]) error {
		// Random processing time between 1-8 seconds
		processingTime := time.Duration(rand.Intn(8)+1) * time.Second

		log.Info("processing job",
			"job-id", job.Id,
			"duration", fmt.Sprintf("%.2fs", processingTime.Seconds()))

		// Simulate work with random duration
		time.Sleep(processingTime)

		log.Info("job completed", "job-id", job.Id)
		return nil
	}

	// Create scheduler
	const workerCount = 5
	const interval = 2 * time.Second
	const visibilityTimeout = 30 * time.Second // Jobs become visible again after 30s if worker crashes

	scheduler := scheduler.NewScheduler(store, workerCount, interval, visibilityTimeout, jobHandler, log)
	done := scheduler.Run(ctx)

	log.Info("scheduler started",
		"workers", workerCount,
		"interval", interval,
		"visibility_timeout", visibilityTimeout,
		"jobs", 50)
	log.Info("press Ctrl+C to stop gracefully")

	// Wait for shutdown signal
	sig := <-sigChan
	log.Info("received shutdown signal", "signal", sig)
	cancel()

	// Wait for graceful shutdown
	<-done
	log.Info("scheduler stopped gracefully")
}
