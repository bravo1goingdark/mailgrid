package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Server provides HTTP endpoints for metrics and health checks
type Server struct {
	metrics *Metrics
	srv     *http.Server
}

// NewServer creates a new metrics server
func NewServer(metrics *Metrics, port int) *Server {
	mux := http.NewServeMux()

	s := &Server{
		metrics: metrics,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	// Register handlers
	mux.Handle("/metrics", metrics)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)

	return s
}

// Start starts the metrics server
func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Basic health check - just return 200 OK
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	// Check if system is ready to handle traffic
	m := s.metrics

	// Consider system ready if:
	// 1. Has been running for at least 1 minute
	// 2. Has active connections
	// 3. Recent error rate is acceptable
	uptime := time.Since(m.startTime)
	activeConns := m.ActiveConnections
	recentErrors := m.ConsecutiveErrs

	if uptime < time.Minute {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "System still starting up (uptime: %v)", uptime)
		return
	}

	if activeConns <= 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "No active SMTP connections")
		return
	}

	if recentErrors > 10 {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "High error rate detected (%d consecutive errors)", recentErrors)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Ready")
}