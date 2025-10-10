// Package email provides high-performance email sending capabilities.
package email

import (
	"context"
	"errors"
	"github.com/bravo1goingdark/mailgrid/config"
	"net/smtp"
	"sync"
	"time"
)

var (
	ErrPoolClosed      = errors.New("connection pool is closed")
	ErrPoolExhausted   = errors.New("connection pool exhausted")
	ErrConnectionStale = errors.New("connection is stale")
)

type PoolConfig struct {
	// Initial number of connections in pool
	InitialSize int
	// Maximum number of connections pool can hold
	MaxSize int
	// Maximum time a connection can be idle before being closed
	MaxIdleTime time.Duration
	// Maximum time to wait for a connection from pool
	MaxWaitTime time.Duration
	// How often to check connection health
	HealthCheckInterval time.Duration
}

// SMTPPool manages a pool of SMTP connections with health checking and circuit breaking
type SMTPPool struct {
	mu sync.RWMutex

	// Connection management
	conns    chan *poolConn
	numConns int
	config   PoolConfig
	smtpCfg  config.SMTPConfig

	// Health checking
	healthCheck     chan struct{}
	healthCheckStop chan struct{}

	// Circuit breaker
	failures       int
	lastFailure    time.Time
	breaker        bool
	breakerTimeout time.Duration

	// Lifecycle
	closed bool
}

type poolConn struct {
	client    *smtp.Client
	createdAt time.Time
	lastUsed  time.Time
}

// NewSMTPPool creates a new connection pool with the given configuration
func NewSMTPPool(smtpCfg config.SMTPConfig, poolCfg PoolConfig) (*SMTPPool, error) {
	if poolCfg.InitialSize <= 0 {
		poolCfg.InitialSize = 5
	}
	if poolCfg.MaxSize <= 0 {
		poolCfg.MaxSize = 20
	}
	if poolCfg.MaxIdleTime <= 0 {
		poolCfg.MaxIdleTime = 5 * time.Minute
	}
	if poolCfg.MaxWaitTime <= 0 {
		poolCfg.MaxWaitTime = 30 * time.Second
	}
	if poolCfg.HealthCheckInterval <= 0 {
		poolCfg.HealthCheckInterval = 30 * time.Second
	}

	p := &SMTPPool{
		conns:           make(chan *poolConn, poolCfg.MaxSize),
		config:          poolCfg,
		smtpCfg:         smtpCfg,
		healthCheck:     make(chan struct{}),
		healthCheckStop: make(chan struct{}),
		breakerTimeout:  1 * time.Minute,
	}

	// Initialize pool with connections
	for i := 0; i < poolCfg.InitialSize; i++ {
		conn, err := p.createConn()
		if err != nil {
			p.Close()
			return nil, err
		}
		p.conns <- conn
		p.numConns++
	}

	// Start health checker
	go p.healthChecker()

	return p, nil
}

// Get gets a connection from the pool, creating a new one if needed
func (p *SMTPPool) Get(ctx context.Context) (*smtp.Client, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, ErrPoolClosed
	}
	if p.breaker {
		if time.Since(p.lastFailure) > p.breakerTimeout {
			p.mu.RUnlock()
			p.resetBreaker()
		} else {
			p.mu.RUnlock()
			return nil, errors.New("circuit breaker is open")
		}
	}
	p.mu.RUnlock()

	select {
	case conn := <-p.conns:
		if time.Since(conn.lastUsed) > p.config.MaxIdleTime {
			_ = conn.client.Close()
			conn, err := p.createConn()
			if err != nil {
				return nil, err
			}
			return conn.client, nil
		}
		conn.lastUsed = time.Now()
		return conn.client, nil

	case <-time.After(p.config.MaxWaitTime):
		p.mu.Lock()
		if p.numConns < p.config.MaxSize {
			conn, err := p.createConn()
			if err != nil {
				p.mu.Unlock()
				return nil, err
			}
			p.numConns++
			p.mu.Unlock()
			return conn.client, nil
		}
		p.mu.Unlock()
		return nil, ErrPoolExhausted

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Put returns a connection to the pool
func (p *SMTPPool) Put(client *smtp.Client) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return ErrPoolClosed
	}
	p.mu.RUnlock()

	conn := &poolConn{
		client:   client,
		lastUsed: time.Now(),
	}

	select {
	case p.conns <- conn:
		return nil
	default:
		// Pool is full, close the connection
		return client.Close()
	}
}

// Close closes the connection pool and all its connections
func (p *SMTPPool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return ErrPoolClosed
	}
	p.closed = true
	p.mu.Unlock()

	// Stop health checker
	close(p.healthCheckStop)

	// Close all connections
	close(p.conns)
	for conn := range p.conns {
		_ = conn.client.Close()
	}

	return nil
}

func (p *SMTPPool) createConn() (*poolConn, error) {
	client, err := ConnectSMTP(p.smtpCfg)
	if err != nil {
		p.recordFailure()
		return nil, err
	}
	return &poolConn{
		client:    client,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
	}, nil
}

func (p *SMTPPool) healthChecker() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.checkConnections()
		case <-p.healthCheckStop:
			return
		}
	}
}

func (p *SMTPPool) checkConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	// Only check a subset of connections to reduce overhead
	checkCount := len(p.conns)
	if checkCount > 5 {
		checkCount = 5 // Limit to checking 5 connections per cycle
	}

	var checkedConns []*poolConn
	unhealthyCount := 0

	// Check only a limited number of connections
	for i := 0; i < checkCount && len(p.conns) > 0; i++ {
		conn := <-p.conns
		if p.isConnHealthy(conn) {
			checkedConns = append(checkedConns, conn)
		} else {
			_ = conn.client.Close()
			unhealthyCount++
			p.numConns--
		}
	}

	// Replace unhealthy connections
	for i := 0; i < unhealthyCount && p.numConns < p.config.MaxSize; i++ {
		if conn, err := p.createConn(); err == nil {
			checkedConns = append(checkedConns, conn)
			p.numConns++
		}
	}

	// Restore checked connections
	for _, conn := range checkedConns {
		p.conns <- conn
	}
}

func (p *SMTPPool) isConnHealthy(conn *poolConn) bool {
	if time.Since(conn.lastUsed) > p.config.MaxIdleTime {
		return false
	}

	// Test connection with NOOP
	if err := conn.client.Noop(); err != nil {
		return false
	}

	return true
}

func (p *SMTPPool) recordFailure() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.failures++
	p.lastFailure = time.Now()

	// Trip circuit breaker after consecutive failures
	if p.failures >= 3 {
		p.breaker = true
	}
}

func (p *SMTPPool) resetBreaker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.breaker = false
	p.failures = 0
}
