package main

import (
	"context"
	"fmt"
	"go-scheduler/internal/scheduler"
	"go-scheduler/internal/storage"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var storage = storage.NewInMemStorage()
	for i := 0; i < 50; i++ {
		randomTime := time.Now().Add(time.Duration(rand.Intn(120)) * time.Second)
		entry := &scheduler.JobEntry{
			Id:        fmt.Sprintf("job-%d", i),
			Status:    "pending",
			ProcessAt: randomTime,
		}
		storage.AddEntry(entry)
	}

	var log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	const workerCount = 3
	const interval = 3 * time.Second

	var jobHandler = func(ctx context.Context, job *scheduler.JobEntry) error {
		randomTime := time.Duration(rand.Intn(8)) * time.Second
		log.Info("job started and will take some time", "job-id", job.Id, "duration", randomTime)
		time.Sleep(randomTime)
		log.Info("job completed", "job-id", job.Id)
		return nil
	}

	var s = scheduler.NewScheduler(storage, workerCount, interval, jobHandler, log)

	done := s.Run(ctx)

	var sig = <-sigChan
	log.Info("received shutdown signal", "signal", sig)
	cancel()

	<-done
	log.Info("scheduler finished")
}
