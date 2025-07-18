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
	couchbasestore "go-sched/storage/couchbase"

	"github.com/couchbase/gocb/v2"
)

// EmailJob represents an email sending task
type EmailJob struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Type    string `json:"type"`
}

// Sample email types and recipients
var emailTypes = []string{"welcome", "newsletter", "reminder", "notification"}
var recipients = []string{
	"alice@example.com", "bob@company.org", "charlie@startup.io", "diana@tech.com",
}

func main() {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a simple logger
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Connect to Couchbase
	couchbaseURI := os.Getenv("COUCHBASE_URI")
	if couchbaseURI == "" {
		couchbaseURI = "couchbase://localhost"
	}

	username := os.Getenv("COUCHBASE_USERNAME")
	if username == "" {
		username = "Administrator"
	}

	password := os.Getenv("COUCHBASE_PASSWORD")
	if password == "" {
		password = "password"
	}

	cluster, err := gocb.Connect(couchbaseURI, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		log.Error("failed to connect to Couchbase", "error", err)
		os.Exit(1)
	}
	defer cluster.Close(nil)

	// Wait until the cluster is ready for use
	err = cluster.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Error("failed to wait for Couchbase cluster", "error", err)
		os.Exit(1)
	}

	log.Info("connected to Couchbase", "uri", couchbaseURI)

	// Get bucket and collection
	bucket := cluster.Bucket("scheduler")

	// Wait for the bucket to be ready
	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Error("failed to wait for Couchbase bucket", "error", err)
		os.Exit(1)
	}

	// Create Couchbase storage - using jobs scope with email-jobs collection
	store := couchbasestore.NewCouchbaseStore[EmailJob](bucket, "jobs", "email-jobs")

	// Add sample email jobs with random scheduling
	for i := 1; i <= 20; i++ {
		// Random delay between 1-30 seconds from now
		randomDelay := time.Duration(rand.Intn(30)+1) * time.Second

		// Create simple email job
		emailType := emailTypes[rand.Intn(len(emailTypes))]
		recipient := recipients[rand.Intn(len(recipients))]

		emailJob := EmailJob{
			To:      recipient,
			Subject: fmt.Sprintf("%s email #%d", emailType, i),
			Type:    emailType,
		}

		job := scheduler.NewJob(time.Now().Add(randomDelay), emailJob)

		if err := store.AddJob(job); err != nil {
			log.Error("failed to add email job", "job-id", job.Id, "error", err)
		}
	}

	// Create email sending job handler
	jobHandler := func(ctx context.Context, job scheduler.Job[EmailJob]) error {
		email := job.Payload

		// Random processing time between 1-5 seconds (email sending simulation)
		processingTime := time.Duration(rand.Intn(5)+1) * time.Second

		log.Info("sending email",
			"job-id", job.Id,
			"to", email.To,
			"type", email.Type,
			"subject", email.Subject,
			"duration", fmt.Sprintf("%.2fs", processingTime.Seconds()))

		// Simulate email sending work
		time.Sleep(processingTime)

		// Simulate occasional failures (5% failure rate)
		if rand.Intn(100) < 5 {
			log.Error("failed to send email", "job-id", job.Id, "to", email.To, "error", "SMTP server temporarily unavailable")
			return fmt.Errorf("SMTP server temporarily unavailable")
		}

		log.Info("email sent successfully",
			"job-id", job.Id,
			"to", email.To,
			"type", email.Type)
		return nil
	}

	// Create scheduler
	const workerCount = 5
	const interval = 2 * time.Second
	const visibilityTimeout = 30 * time.Second // Jobs auto-recover after 30s if worker crashes

	scheduler := scheduler.NewScheduler(store, workerCount, interval, visibilityTimeout, jobHandler, log)
	done := scheduler.Run(ctx)

	log.Info("email scheduler started",
		"workers", workerCount,
		"interval", interval,
		"visibility_timeout", visibilityTimeout,
		"jobs", 20,
		"storage", "couchbase",
		"bucket", "scheduler",
		"scope", "jobs",
		"collection", "email-jobs")
	log.Info("press Ctrl+C to stop gracefully")

	// Wait for shutdown signal
	sig := <-sigChan
	log.Info("received shutdown signal", "signal", sig)
	cancel()

	// Wait for graceful shutdown
	<-done
	log.Info("scheduler stopped gracefully")
}
