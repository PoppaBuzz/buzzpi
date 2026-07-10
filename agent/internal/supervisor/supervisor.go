// Package supervisor manages the Runtime process lifecycle.
// It is the first component started and the last stopped.
// It owns startup ordering, health monitoring, and graceful shutdown.
package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// HealthStatus represents the health of a component.
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthStarting
	HealthOK
	HealthDegraded
	HealthFailed
)

func (s HealthStatus) String() string {
	switch s {
	case HealthUnknown:
		return "unknown"
	case HealthStarting:
		return "starting"
	case HealthOK:
		return "ok"
	case HealthDegraded:
		return "degraded"
	case HealthFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Component is the interface every Runtime sub-component must implement.
type Component interface {
	// Name returns the component's human-readable name.
	Name() string

	// Start initializes the component. It should block until the component is
	// fully running, or return an error if initialization fails.
	Start(ctx context.Context) error

	// Stop shuts down the component gracefully. It should return within the
	// context's deadline or cancellation.
	Stop(ctx context.Context) error

	// Health returns the component's current health status.
	Health() interface{}
}

// ComponentState tracks a component's lifecycle state.
type ComponentState struct {
	Component
	health    HealthStatus
	restarts  int
	lastError error
	mu        sync.RWMutex
}

func (cs *ComponentState) setHealth(h HealthStatus) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.health = h
}

// Config defines the Supervisor's behavior.
type Config struct {
	// ShutdownTimeout is the maximum time to wait for all components to stop.
	ShutdownTimeout time.Duration

	// MaxRestarts is the maximum number of times a component is restarted
	// within RestartWindow before giving up.
	MaxRestarts int

	// RestartWindow is the time window for restart counting.
	RestartWindow time.Duration

	// HealthCheckInterval is how often to poll component health.
	HealthCheckInterval time.Duration
}

// DefaultConfig returns a default Supervisor configuration.
func DefaultConfig() Config {
	return Config{
		ShutdownTimeout:     30 * time.Second,
		MaxRestarts:         3,
		RestartWindow:       5 * time.Minute,
		HealthCheckInterval: 15 * time.Second,
	}
}

// Supervisor manages component lifecycle and process signals.
type Supervisor struct {
	cfg   Config
	comps []*ComponentState
	log   *slog.Logger
	mu    sync.RWMutex
}

// New creates a new Supervisor.
func New(cfg Config, log *slog.Logger) *Supervisor {
	if log == nil {
		log = slog.Default()
	}
	return &Supervisor{
		cfg: cfg,
		log: log.With("component", "supervisor"),
	}
}

// Register adds a component to the supervisor. Components are started in the
// order they are registered and stopped in reverse order.
func (s *Supervisor) Register(c Component) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.comps = append(s.comps, &ComponentState{Component: c})
}

// Run starts all components and blocks until a shutdown signal is received
// or a component fails fatally. It handles SIGTERM and SIGINT.
func (s *Supervisor) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start all components in registration order
	if err := s.startAll(ctx); err != nil {
		return fmt.Errorf("startup failed: %w", err)
	}

	s.log.Info("all components started", "count", len(s.comps))

	// Health monitoring goroutine
	go s.healthLoop(ctx)

	// Wait for shutdown signal or context cancellation
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sigCh:
		s.log.Info("received signal", "signal", sig)
	case <-ctx.Done():
		s.log.Info("context cancelled")
	}

	return s.shutdown(ctx)
}

func (s *Supervisor) startAll(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, cs := range s.comps {
		compCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		cs.setHealth(HealthStarting)

		s.log.Info("starting component", "name", cs.Name())
		if err := cs.Start(compCtx); err != nil {
			cancel()
			cs.setHealth(HealthFailed)
			cs.lastError = err
			return fmt.Errorf("%s: %w", cs.Name(), err)
		}
		cancel()
		cs.setHealth(HealthOK)
		s.log.Info("component started", "name", cs.Name())
	}
	return nil
}

func (s *Supervisor) shutdown(ctx context.Context) error {
	s.log.Info("shutting down all components")
	shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()

	s.mu.RLock()
	// Stop in reverse order
	for i := len(s.comps) - 1; i >= 0; i-- {
		cs := s.comps[i]
		s.log.Info("stopping component", "name", cs.Name())
		if err := cs.Stop(shutdownCtx); err != nil {
			s.log.Error("component stop error",
				"name", cs.Name(),
				"error", err,
			)
		}
	}
	s.mu.RUnlock()

	s.log.Info("shutdown complete")
	return nil
}

func (s *Supervisor) healthLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkHealth()
		}
	}
}

func (s *Supervisor) checkHealth() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, cs := range s.comps {
		_ = cs.Health()
		// Component health is self-reported via the Health() method.
		// Lifecycle state (starting/failed/ok) is tracked by the supervisor
		// via Start/Stop — not via Health() polling.
	}
}

// Components returns the list of registered components (for diagnostics).
func (s *Supervisor) Components() []Component {
	s.mu.RLock()
	defer s.mu.RUnlock()

	comps := make([]Component, len(s.comps))
	for i, cs := range s.comps {
		comps[i] = cs.Component
	}
	return comps
}
