package scheduler_test

import (
	"os"
	"testing"
	"time"

	"github.com/bravo1goingdark/mailgrid/database"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/bravo1goingdark/mailgrid/logger"
	"github.com/bravo1goingdark/mailgrid/scheduler"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *database.BoltDBClient {
	dbPath := "test_scheduler.db"
	os.Remove(dbPath) // Clean up any previous test DB
	db, err := database.NewDB(dbPath)
	assert.NoError(t, err)
	return db
}

func teardownTestDB(db *database.BoltDBClient) {
	db.Close()
	os.Remove("test_scheduler.db")
}

func TestNewScheduler(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	log := logrus.New()
	sched := scheduler.NewScheduler(db, log)
	assert.NotNil(t, sched)
}

func TestAddJob(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	log := logger.New("test-scheduler")
	sched := scheduler.NewScheduler(db, log)

	args := types.CLIArgs{
		Subject: "Test Job",
	}
	job := scheduler.NewJob(args, time.Time{}, "", "1s")

	err := sched.AddJob(job, func(j types.Job) error { return nil })
	assert.NoError(t, err)

	jobs, err := sched.ListJobs()
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, job.ID, jobs[0].ID)
}

func TestCancelJob(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	log := logger.New("test-scheduler")
	sched := scheduler.NewScheduler(db, log)

	args := types.CLIArgs{
		Subject: "Test Job to Cancel",
	}
	job := scheduler.NewJob(args, time.Time{}, "", "1s")

	err := sched.AddJob(job, func(j types.Job) error { return nil })
	assert.NoError(t, err)

	cancelled := sched.CancelJob(job.ID)
	assert.True(t, cancelled)

	jobs, err := sched.ListJobs()
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "cancelled", jobs[0].Status)
}

func TestListJobs(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	log := logger.New("test-scheduler")
	sched := scheduler.NewScheduler(db, log)

	args1 := types.CLIArgs{Subject: "Job 1"}
	job1 := scheduler.NewJob(args1, time.Time{}, "", "1s")
	err := sched.AddJob(job1, func(j types.Job) error { return nil })
	assert.NoError(t, err)

	args2 := types.CLIArgs{Subject: "Job 2"}
	job2 := scheduler.NewJob(args2, time.Time{}, "", "1s")
	err = sched.AddJob(job2, func(j types.Job) error { return nil })
	assert.NoError(t, err)

	jobs, err := sched.ListJobs()
	assert.NoError(t, err)
	assert.Len(t, jobs, 2)

	// Verify job IDs are present
	found1 := false
	found2 := false
	for _, j := range jobs {
		if j.ID == job1.ID {
			found1 = true
		}
		if j.ID == job2.ID {
			found2 = true
		}
	}
	assert.True(t, found1)
	assert.True(t, found2)
}

func TestJobExecution(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	log := logger.New("test-scheduler")
	sched := scheduler.NewScheduler(db, log)

	// Use a channel to signal job execution
	doneChan := make(chan bool, 1)
	handler := func(j types.Job) error {
		doneChan <- true
		return nil
	}

	args := types.CLIArgs{Subject: "Executable Job"}
	// Schedule a job to run in 100ms
	job := scheduler.NewJob(args, time.Now().Add(100*time.Millisecond), "", "")

	err := sched.AddJob(job, handler)
	assert.NoError(t, err)

	select {
	case <-doneChan:
		// Job executed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Job did not execute in time")
	}

	// Verify job status after execution with a short poll to avoid races
	deadline := time.Now().Add(2 * time.Second)
	for {
		jobs, err := sched.ListJobs()
		assert.NoError(t, err)
		if assert.Len(t, jobs, 1) && jobs[0].Status == "done" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("job status did not reach 'done' in time; last=%q", jobs[0].Status)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
