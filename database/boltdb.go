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
	jobsBucket      = "jobs"
	lockBucket     = "locks"
	lockExpiryTime = 5 * time.Minute
)

// BoltDBClient is a wrapper around bbolt.DB for job persistence.
type BoltDBClient struct {
	db *bbolt.DB
}

// NewDB opens a BoltDB database and initializes necessary buckets.
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

// parseLockInfo parses lock information from lock value
func parseLockInfo(lockData []byte) (instanceID string, lockedAt time.Time, err error) {
	parts := strings.Split(string(lockData), ":")
	if len(parts) != 2 {
		return "", time.Time{}, fmt.Errorf("malformed lock info: expected format instanceID:timestamp")
	}

	instanceID = parts[0]
	lockedAtNano, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("invalid timestamp in lock info: %w", err)
	}

	lockedAt = time.Unix(0, lockedAtNano)
	return instanceID, lockedAt, nil
}

// formatLockInfo formats lock information for storage
func formatLockInfo(instanceID string) string {
	return fmt.Sprintf("%s:%d", instanceID, time.Now().UnixNano())
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
			if err := b.Put(lockKey, []byte(formatLockInfo(instanceID))); err != nil {
				return errors.Wrap(err, "failed to put lock")
			}
			locked = true
			return nil
		}

		// Parse existing lock
		heldBy, lockedAt, err := parseLockInfo(currentLock)
		if err != nil {
			// Malformed lock - don't silently overwrite, log and return error
			// This prevents race conditions where a corrupted lock could be hijacked
			return errors.Wrap(err, "failed to parse existing lock")
		}

		// Check if lock is expired or held by this instance
		if heldBy == instanceID || time.Since(lockedAt) > lockExpiryTime {
			// Lock held by this instance or expired, re-acquire
			if err := b.Put(lockKey, []byte(formatLockInfo(instanceID))); err != nil {
				return errors.Wrap(err, "failed to re-acquire lock")
			}
			locked = true
			return nil
		}

		// Lock held by another active instance
		locked = false
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

		// Parse to verify ownership
		heldBy, _, err := parseLockInfo(currentLock)
		if err != nil {
			// Malformed lock, safe to delete
			return errors.Wrap(b.Delete(lockKey), "failed to delete malformed lock")
		}

		// Only release if held by this instance
		if heldBy == instanceID {
			return errors.Wrap(b.Delete(lockKey), "failed to delete lock")
		}

		// Lock held by different instance, don't release
		return nil
	})
}

// CleanupExpiredLocks removes locks that have expired
func (c *BoltDBClient) CleanupExpiredLocks() (int, error) {
	cleaned := 0
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(lockBucket))
		cursor := b.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			_, lockedAt, err := parseLockInfo(v)
			if err != nil {
				// Malformed lock, clean it up
				if err := b.Delete(k); err == nil {
					cleaned++
				}
				continue
			}

			// Check if expired
			if time.Since(lockedAt) > lockExpiryTime {
				if err := b.Delete(k); err == nil {
					cleaned++
				}
			}
		}
		return nil
	})

	return cleaned, err
}
