package database

import (
	"encoding/json"
	"fmt"
	"github.com/bravo1goingdark/mailgrid/internal/types"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"strconv"
	"strings"
	"time"
)

const (
	jobsBucket = "jobs"
	lockBucket = "locks"
)

// BoltDBClient is a wrapper around bbolt.DB for job persistence.
type BoltDBClient struct {
	db *bbolt.DB
}

// NewDB opens a BoltDB database and initializes the necessary buckets.
func NewDB(path string) (*BoltDBClient, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open BoltDB at %s", path)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(jobsBucket))
		if err != nil {
			return errors.Wrapf(err, "create %s bucket", jobsBucket)
		}
		_, err = tx.CreateBucketIfNotExists([]byte(lockBucket))
		if err != nil {
			return errors.Wrapf(err, "create %s bucket", lockBucket)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize BoltDB buckets")
	}

	return &BoltDBClient{db: db},
		nil
}

// Close closes the BoltDB database.
func (c *BoltDBClient) Close() error {
	return c.db.Close()
}

// SaveJob saves a job to the BoltDB database.
func (c *BoltDBClient) SaveJob(job *types.Job) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(jobsBucket))
		encoded, err := json.Marshal(job)
		if err != nil {
			return errors.Wrap(err, "could not marshal job")
		}
		return errors.Wrap(b.Put([]byte(job.ID), encoded), "could not put job")
	})
}

// GetJob retrieves a job from the BoltDB database.
func (c *BoltDBClient) GetJob(id string) (*types.Job, error) {
	var job types.Job
	err := c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(jobsBucket))
		val := b.Get([]byte(id))
		if val == nil {
			return errors.New("job not found")
		}
		return errors.Wrap(json.Unmarshal(val, &job), "could not unmarshal job")
	})
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// LoadJobs loads all jobs from the BoltDB database.
func (c *BoltDBClient) LoadJobs() ([]types.Job, error) {
	var jobs []types.Job
	err := c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(jobsBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var job types.Job
			if err := json.Unmarshal(v, &job); err != nil {
				return errors.Wrap(err, "could not unmarshal job from bucket")
			}
			jobs = append(jobs, job)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// AcquireLock attempts to acquire a lock for a job.
func (c *BoltDBClient) AcquireLock(jobID, instanceID string) (bool, error) {
	var locked bool
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(lockBucket))
		lockKey := []byte(jobID)
		currentLock := b.Get(lockKey)

		if currentLock == nil {
			// No lock exists, acquire it
			lockInfo := fmt.Sprintf("%s:%d", instanceID, time.Now().UnixNano())
			if err := b.Put(lockKey, []byte(lockInfo)); err != nil {
				return errors.Wrap(err, "failed to put lock")
			}
			locked = true
			return nil
		}

		// Lock exists, check if it's expired or held by this instance
		parts := strings.Split(string(currentLock), ":")
		if len(parts) != 2 {
			// Malformed lock, overwrite it
			lockInfo := fmt.Sprintf("%s:%d", instanceID, time.Now().UnixNano())
			if err := b.Put(lockKey, []byte(lockInfo)); err != nil {
				return errors.Wrap(err, "failed to overwrite malformed lock")
			}
			locked = true
			return nil
		}

		heldBy := parts[0]
		lockedAtNano, _ := strconv.ParseInt(parts[1], 10, 64)
		lockedAt := time.Unix(0, lockedAtNano)

		if heldBy == instanceID || time.Since(lockedAt) > 5*time.Minute { // 5-minute expiry
			// Lock held by this instance or expired, re-acquire
			lockInfo := fmt.Sprintf("%s:%d", instanceID, time.Now().UnixNano())
			if err := b.Put(lockKey, []byte(lockInfo)); err != nil {
				return errors.Wrap(err, "failed to re-acquire lock")
			}
			locked = true
			return nil
		}

		locked = false // Lock held by another active instance
		return nil
	})
	if err != nil {
		return false, err
	}
	return locked, nil
}

// ReleaseLock releases a lock for a job.
func (c *BoltDBClient) ReleaseLock(jobID, instanceID string) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(lockBucket))
		lockKey := []byte(jobID)
		currentLock := b.Get(lockKey)

		if currentLock == nil {
			return nil // No lock to release
		}

		parts := strings.Split(string(currentLock), ":")
		if len(parts) == 2 && parts[0] == instanceID {
			// Only release if held by this instance
			return errors.Wrap(b.Delete(lockKey), "failed to delete lock")
		}
		return nil
	})
}
