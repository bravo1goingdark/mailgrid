package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func createTempDB(t *testing.T) (*BoltDBClient, string, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	client, err := NewDB(dbPath)
	require.NoError(t, err)

	cleanup := func() {
		client.Close()
		os.Remove(dbPath)
	}

	return client, dbPath, cleanup
}

func TestNewDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	client, err := NewDB(dbPath)
	require.NoError(t, err)
	defer client.Close()

	assert.NotNil(t, client)
	assert.NotNil(t, client.db)

	// Verify file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}

func TestNewDB_InvalidPath(t *testing.T) {
	// Try to create database in non-existent directory without permission
	client, err := NewDB("/invalid/path/test.db")
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestSaveAndGetJob(t *testing.T) {
	client, _, cleanup := createTempDB(t)
	defer cleanup()

	job := &types.Job{
		ID:          "test-job-1",
		Status:      "pending",
		RunAt:       time.Now(),
		NextRunAt:   time.Now().Add(time.Hour),
		Args:        []byte(`{"test": "data"}`),
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save job
	err := client.SaveJob(job)
	require.NoError(t, err)

	// Get job
	retrieved, err := client.GetJob("test-job-1")
	require.NoError(t, err)
	assert.Equal(t, job.ID, retrieved.ID)
	assert.Equal(t, job.Status, retrieved.Status)
	assert.JSONEq(t, string(job.Args), string(retrieved.Args)) // Compare JSON content, not raw bytes
	assert.Equal(t, job.Attempts, retrieved.Attempts)
	assert.Equal(t, job.MaxAttempts, retrieved.MaxAttempts)
}

func TestGetJob_NotFound(t *testing.T) {
	client, _, cleanup := createTempDB(t)
	defer cleanup()

	job, err := client.GetJob("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "job not found")
}

func TestLoadJobs(t *testing.T) {
	client, _, cleanup := createTempDB(t)
	defer cleanup()

	// Test empty database
	jobs, err := client.LoadJobs()
	require.NoError(t, err)
	assert.Empty(t, jobs)

	// Add multiple jobs
	job1 := &types.Job{
		ID:          "job-1",
		Status:      "pending",
		RunAt:       time.Now(),
		Args:        []byte(`{"test": "data1"}`),
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	job2 := &types.Job{
		ID:          "job-2",
		Status:      "completed",
		RunAt:       time.Now(),
		Args:        []byte(`{"test": "data2"}`),
		Attempts:    1,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = client.SaveJob(job1)
	require.NoError(t, err)
	err = client.SaveJob(job2)
	require.NoError(t, err)

	// Load all jobs
	jobs, err = client.LoadJobs()
	require.NoError(t, err)
	assert.Len(t, jobs, 2)

	// Verify jobs
	jobMap := make(map[string]types.Job)
	for _, job := range jobs {
		jobMap[job.ID] = job
	}

	assert.Contains(t, jobMap, "job-1")
	assert.Contains(t, jobMap, "job-2")
	assert.Equal(t, "pending", jobMap["job-1"].Status)
	assert.Equal(t, "completed", jobMap["job-2"].Status)
}

func TestAcquireLock(t *testing.T) {
	client, _, cleanup := createTempDB(t)
	defer cleanup()

	jobID := "test-job"
	instanceID1 := "instance-1"
	instanceID2 := "instance-2"

	// First acquisition should succeed
	locked, err := client.AcquireLock(jobID, instanceID1)
	require.NoError(t, err)
	assert.True(t, locked)

	// Same instance should be able to re-acquire
	locked, err = client.AcquireLock(jobID, instanceID1)
	require.NoError(t, err)
	assert.True(t, locked)

	// Different instance should fail to acquire
	locked, err = client.AcquireLock(jobID, instanceID2)
	require.NoError(t, err)
	assert.False(t, locked)
}

func TestAcquireLock_ExpiredLock(t *testing.T) {
	client, _, cleanup := createTempDB(t)
	defer cleanup()

	jobID := "test-job"
	instanceID1 := "instance-1"
	instanceID2 := "instance-2"

	// Acquire lock
	locked, err := client.AcquireLock(jobID, instanceID1)
	require.NoError(t, err)
	assert.True(t, locked)

	// Manually expire the lock by modifying the timestamp
	// This is a bit hacky but tests the expiration logic
	err = client.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(lockBucket))
		oldTime := time.Now().Add(-6 * time.Minute).UnixNano()
		lockInfo := instanceID1 + ":" + fmt.Sprintf("%d", oldTime)
		return b.Put([]byte(jobID), []byte(lockInfo))
	})
	require.NoError(t, err)

	// Different instance should now be able to acquire the expired lock
	locked, err = client.AcquireLock(jobID, instanceID2)
	require.NoError(t, err)
	assert.True(t, locked)
}

func TestReleaseLock(t *testing.T) {
	client, _, cleanup := createTempDB(t)
	defer cleanup()

	jobID := "test-job"
	instanceID1 := "instance-1"
	instanceID2 := "instance-2"

	// Acquire lock
	locked, err := client.AcquireLock(jobID, instanceID1)
	require.NoError(t, err)
	assert.True(t, locked)

	// Release lock
	err = client.ReleaseLock(jobID, instanceID1)
	require.NoError(t, err)

	// Different instance should now be able to acquire
	locked, err = client.AcquireLock(jobID, instanceID2)
	require.NoError(t, err)
	assert.True(t, locked)
}

func TestReleaseLock_WrongInstance(t *testing.T) {
	client, _, cleanup := createTempDB(t)
	defer cleanup()

	jobID := "test-job"
	instanceID1 := "instance-1"
	instanceID2 := "instance-2"

	// Acquire lock
	locked, err := client.AcquireLock(jobID, instanceID1)
	require.NoError(t, err)
	assert.True(t, locked)

	// Try to release with wrong instance - should not release
	err = client.ReleaseLock(jobID, instanceID2)
	require.NoError(t, err) // No error, but lock not released

	// Original instance should still hold the lock
	locked, err = client.AcquireLock(jobID, instanceID2)
	require.NoError(t, err)
	assert.False(t, locked) // Lock still held by instance1
}

func TestReleaseLock_NoLock(t *testing.T) {
	client, _, cleanup := createTempDB(t)
	defer cleanup()

	// Try to release non-existent lock - should not error
	err := client.ReleaseLock("nonexistent", "instance")
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	client, _, cleanup := createTempDB(t)
	defer cleanup()

	err := client.Close()
	assert.NoError(t, err)

	// Operations after close should fail
	err = client.SaveJob(&types.Job{ID: "test"})
	assert.Error(t, err)
}