package database

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/bravo1goingdark/mailgrid/internal/types"
)

func TestNewDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create BoltDB: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("BoltDB instance is nil")
	}
}

func TestNewDBInvalidPath(t *testing.T) {
	// Try to create DB in non-existent directory without creating it
	invalidPath := "/non/existent/path/test.db"

	_, err := NewDB(invalidPath)
	if err == nil {
		t.Error("Expected error when creating BoltDB with invalid path")
	}
}

func TestBoltDB_SaveAndGetJob(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create BoltDB: %v", err)
	}
	defer db.Close()

	// Create test job
	testJob := &types.Job{
		ID:        "test-job-123",
		Status:    "pending",
		RunAt:     time.Now(),
		CreatedAt: time.Now(),
		Args:      json.RawMessage(`{"env":"test"}`),
	}

	// Save job
	err = db.SaveJob(testJob)
	if err != nil {
		t.Fatalf("Failed to save job: %v", err)
	}

	// Get job
	retrievedJob, err := db.GetJob("test-job-123")
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.ID != testJob.ID {
		t.Errorf("Retrieved job ID doesn't match. Got: %s, Expected: %s",
			retrievedJob.ID, testJob.ID)
	}

	if retrievedJob.Status != testJob.Status {
		t.Errorf("Retrieved job status doesn't match. Got: %s, Expected: %s",
			retrievedJob.Status, testJob.Status)
	}
}

func TestBoltDB_GetNonExistentJob(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create BoltDB: %v", err)
	}
	defer db.Close()

	_, err = db.GetJob("non-existent-job")
	if err == nil {
		t.Error("Expected error when getting non-existent job")
	}
}

func TestBoltDB_LoadJobs(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create BoltDB: %v", err)
	}
	defer db.Close()

	// Save multiple jobs
	testJobs := []*types.Job{
		{ID: "job-1", Status: "pending", RunAt: time.Now(), CreatedAt: time.Now(), Args: json.RawMessage(`{"env":"test1"}`)},
		{ID: "job-2", Status: "pending", RunAt: time.Now(), CreatedAt: time.Now(), Args: json.RawMessage(`{"env":"test2"}`)},
		{ID: "job-3", Status: "pending", RunAt: time.Now(), CreatedAt: time.Now(), Args: json.RawMessage(`{"env":"test3"}`)},
	}

	for _, job := range testJobs {
		err = db.SaveJob(job)
		if err != nil {
			t.Fatalf("Failed to save job %s: %v", job.ID, err)
		}
	}

	// Load all jobs
	allJobs, err := db.LoadJobs()
	if err != nil {
		t.Fatalf("Failed to load jobs: %v", err)
	}

	if len(allJobs) != len(testJobs) {
		t.Errorf("Expected %d jobs, got %d", len(testJobs), len(allJobs))
	}

	// Check that all job IDs are present
	jobIDs := make(map[string]bool)
	for _, job := range allJobs {
		jobIDs[job.ID] = true
	}

	for _, expectedJob := range testJobs {
		if !jobIDs[expectedJob.ID] {
			t.Errorf("Expected job ID %s not found in loaded jobs", expectedJob.ID)
		}
	}
}
