package supervisor

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"
)

type mockComponent struct {
	name    string
	startFn func(ctx context.Context) error
	stopFn  func(ctx context.Context) error
	health  interface{}
}

func (m *mockComponent) Name() string { return m.name }
func (m *mockComponent) Start(ctx context.Context) error {
	if m.startFn != nil {
		return m.startFn(ctx)
	}
	return nil
}
func (m *mockComponent) Stop(ctx context.Context) error {
	if m.stopFn != nil {
		return m.stopFn(ctx)
	}
	return nil
}
func (m *mockComponent) Health() interface{} {
	if m.health != nil {
		return m.health
	}
	return "ok"
}

func newTestSupervisor(t *testing.T) *Supervisor {
	t.Helper()
	return New(DefaultConfig(), slog.Default())
}

func TestNew(t *testing.T) {
	s := newTestSupervisor(t)
	if s == nil {
		t.Fatal("New() returned nil")
	}
	comps := s.Components()
	if len(comps) != 0 {
		t.Errorf("New() has %d components, want 0", len(comps))
	}

	t.Run("nil logger uses default", func(t *testing.T) {
		s2 := New(DefaultConfig(), nil)
		if s2 == nil {
			t.Fatal("New(nil logger) returned nil")
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 30s", cfg.ShutdownTimeout)
	}
	if cfg.MaxRestarts != 3 {
		t.Errorf("MaxRestarts = %d, want 3", cfg.MaxRestarts)
	}
	if cfg.RestartWindow != 5*time.Minute {
		t.Errorf("RestartWindow = %v, want 5m", cfg.RestartWindow)
	}
	if cfg.HealthCheckInterval != 15*time.Second {
		t.Errorf("HealthCheckInterval = %v, want 15s", cfg.HealthCheckInterval)
	}
}

func TestRegister(t *testing.T) {
	s := newTestSupervisor(t)

	c1 := &mockComponent{name: "comp1"}
	c2 := &mockComponent{name: "comp2"}

	s.Register(c1)
	s.Register(c2)

	comps := s.Components()
	if len(comps) != 2 {
		t.Fatalf("Components() = %d, want 2", len(comps))
	}
	if comps[0].Name() != "comp1" {
		t.Errorf("Components()[0] = %q, want \"comp1\"", comps[0].Name())
	}
	if comps[1].Name() != "comp2" {
		t.Errorf("Components()[1] = %q, want \"comp2\"", comps[1].Name())
	}
}

func TestHealthStatusString(t *testing.T) {
	tests := []struct {
		status HealthStatus
		want   string
	}{
		{HealthUnknown, "unknown"},
		{HealthStarting, "starting"},
		{HealthOK, "ok"},
		{HealthDegraded, "degraded"},
		{HealthFailed, "failed"},
		{HealthStatus(99), "unknown"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("HealthStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestStartAll(t *testing.T) {
	t.Run("starts all components in order", func(t *testing.T) {
		s := newTestSupervisor(t)
		var mu sync.Mutex
		var order []string

		s.Register(&mockComponent{
			name: "first",
			startFn: func(ctx context.Context) error {
				mu.Lock()
				order = append(order, "first")
				mu.Unlock()
				return nil
			},
		})
		s.Register(&mockComponent{
			name: "second",
			startFn: func(ctx context.Context) error {
				mu.Lock()
				order = append(order, "second")
				mu.Unlock()
				return nil
			},
		})

		err := s.startAll(context.Background())
		if err != nil {
			t.Fatalf("startAll() failed: %v", err)
		}

		if len(order) != 2 || order[0] != "first" || order[1] != "second" {
			t.Errorf("start order = %v, want [first second]", order)
		}
	})

	t.Run("returns error on component failure", func(t *testing.T) {
		s := newTestSupervisor(t)
		wantErr := errors.New("start failed")

		s.Register(&mockComponent{
			name: "good",
			startFn: func(ctx context.Context) error {
				return nil
			},
		})
		s.Register(&mockComponent{
			name: "bad",
			startFn: func(ctx context.Context) error {
				return wantErr
			},
		})

		err := s.startAll(context.Background())
		if err == nil {
			t.Fatal("startAll() expected error, got nil")
		}
	})
}

func TestShutdown(t *testing.T) {
	t.Run("stops components in reverse order", func(t *testing.T) {
		s := newTestSupervisor(t)
		var mu sync.Mutex
		var order []string

		s.Register(&mockComponent{
			name: "first",
			stopFn: func(ctx context.Context) error {
				mu.Lock()
				order = append(order, "first")
				mu.Unlock()
				return nil
			},
		})
		s.Register(&mockComponent{
			name: "second",
			stopFn: func(ctx context.Context) error {
				mu.Lock()
				order = append(order, "second")
				mu.Unlock()
				return nil
			},
		})

		err := s.shutdown(context.Background())
		if err != nil {
			t.Fatalf("shutdown() failed: %v", err)
		}

		if len(order) != 2 || order[0] != "second" || order[1] != "first" {
			t.Errorf("stop order = %v, want [second first]", order)
		}
	})

	t.Run("handles stop errors gracefully", func(t *testing.T) {
		s := newTestSupervisor(t)

		s.Register(&mockComponent{
			name: "a",
			stopFn: func(ctx context.Context) error {
				return errors.New("stop error")
			},
		})
		s.Register(&mockComponent{
			name: "b",
			stopFn: func(ctx context.Context) error {
				return nil
			},
		})

		// Should not return error (stop errors are logged, not returned)
		err := s.shutdown(context.Background())
		if err != nil {
			t.Fatalf("shutdown() should not propagate stop errors: %v", err)
		}
	})
}
