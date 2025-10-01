package email

import (
	"context"
	"testing"
	"time"

	"github.com/bravo1goingdark/mailgrid/config"
)

func TestSMTPPool_Creation(t *testing.T) {
	cfg := config.SMTPConfig{
		Host:     "localhost",
		Port:     1025, // MailHog port for testing
		Username: "test",
		Password: "test",
		From:     "test@example.com",
	}

	poolCfg := PoolConfig{
		InitialSize:         2,
		MaxSize:            5,
		MaxIdleTime:        1 * time.Minute,
		MaxWaitTime:        5 * time.Second,
		HealthCheckInterval: 10 * time.Second,
	}

	// This will fail without a running SMTP server, but tests the structure
	_, err := NewSMTPPool(cfg, poolCfg)
	if err == nil {
		t.Error("Expected error when no SMTP server is running")
	}
}

func TestPoolConfig_Defaults(t *testing.T) {
	cfg := config.SMTPConfig{
		Host: "localhost",
		Port: 1025,
	}

	poolCfg := PoolConfig{} // All zeros, should get defaults

	pool, err := NewSMTPPool(cfg, poolCfg)
	if err == nil {
		defer pool.Close()
		
		if pool.config.InitialSize != 5 {
			t.Errorf("Expected InitialSize=5, got %d", pool.config.InitialSize)
		}
		if pool.config.MaxSize != 20 {
			t.Errorf("Expected MaxSize=20, got %d", pool.config.MaxSize)
		}
	}
}

func TestPool_GetPut_WithoutServer(t *testing.T) {
	cfg := config.SMTPConfig{
		Host:     "localhost",
		Port:     1025,
		Username: "test",
		Password: "test",
		From:     "test@example.com",
	}

	poolCfg := PoolConfig{
		InitialSize: 0, // Start with 0 connections
		MaxSize:     2,
		MaxWaitTime: 100 * time.Millisecond,
	}

	pool, err := NewSMTPPool(cfg, poolCfg)
	if err != nil {
		t.Skip("Skipping pool test - no SMTP server available")
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should fail quickly since no SMTP server is running
	_, err = pool.Get(ctx)
	if err == nil {
		t.Error("Expected error when getting connection without SMTP server")
	}
}

func TestPool_Close(t *testing.T) {
	cfg := config.SMTPConfig{
		Host: "localhost",
		Port: 1025,
	}

	poolCfg := PoolConfig{
		InitialSize: 0,
		MaxSize:     1,
	}

	pool, err := NewSMTPPool(cfg, poolCfg)
	if err != nil {
		t.Skip("Skipping test - no SMTP server available")
	}

	// Close should work even if connections fail
	err = pool.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Second close should return error
	err = pool.Close()
	if err != ErrPoolClosed {
		t.Errorf("Expected ErrPoolClosed, got %v", err)
	}
}