package connection

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	return NewEngine("dev_test", slog.Default())
}

func TestTransportTypeString(t *testing.T) {
	tests := []struct {
		t    TransportType
		want string
	}{
		{TransportNone, "none"},
		{TransportLAN, "lan"},
		{TransportP2P, "p2p"},
		{TransportRelay, "relay"},
		{TransportType(99), "none"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("TransportType(%d).String() = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}

func TestConnectionStateString(t *testing.T) {
	tests := []struct {
		s    ConnectionState
		want string
	}{
		{StateDisconnected, "disconnected"},
		{StateDiscovering, "discovering"},
		{StateConnecting, "connecting"},
		{StateAuthenticating, "authenticating"},
		{StateConnected, "connected"},
		{StateDegraded, "degraded"},
		{ConnectionState(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("ConnectionState(%d).String() = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestQualityLevelString(t *testing.T) {
	tests := []struct {
		q    QualityLevel
		want string
	}{
		{QualityUnknown, "unknown"},
		{QualityHigh, "high"},
		{QualityMedium, "medium"},
		{QualityLow, "low"},
		{QualityMinimum, "minimum"},
		{QualityLevel(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.q.String(); got != tt.want {
				t.Errorf("QualityLevel(%d).String() = %q, want %q", tt.q, got, tt.want)
			}
		})
	}
}

func TestNewEngine(t *testing.T) {
	e := newTestEngine(t)
	if e == nil {
		t.Fatal("NewEngine() returned nil")
	}
	if e.State() != StateDisconnected {
		t.Errorf("initial state = %v, want disconnected", e.State())
	}
	if e.Transport() != TransportNone {
		t.Errorf("initial transport = %v, want none", e.Transport())
	}
	if e.Quality() != QualityUnknown {
		t.Errorf("initial quality = %v, want unknown", e.Quality())
	}
}

func TestNewEngineNilLogger(t *testing.T) {
	e := NewEngine("dev_log", nil)
	if e == nil {
		t.Fatal("NewEngine(nil logger) returned nil")
	}
}

func TestName(t *testing.T) {
	e := newTestEngine(t)
	want := "connection-dev_test"
	if got := e.Name(); got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestHealth(t *testing.T) {
	e := newTestEngine(t)
	h := e.Health()
	m, ok := h.(map[string]interface{})
	if !ok {
		t.Fatalf("Health() returned %T, want map[string]interface{}", h)
	}
	if m["state"] != "disconnected" {
		t.Errorf("Health() state = %v, want disconnected", m["state"])
	}
	if m["transport"] != "none" {
		t.Errorf("Health() transport = %v, want none", m["transport"])
	}
	if m["quality"] != "unknown" {
		t.Errorf("Health() quality = %v, want unknown", m["quality"])
	}
}

func TestStartTransitions(t *testing.T) {
	e := newTestEngine(t)

	events := e.Events()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := e.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Should eventually transition to connected via simulated LAN
	select {
	case evt := <-events:
		if evt.State != StateDiscovering {
			t.Errorf("first event state = %v, want discovering", evt.State)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for first event")
	}

	select {
	case evt := <-events:
		if evt.State != StateConnected {
			t.Errorf("second event state = %v, want connected", evt.State)
		}
		if evt.Transport != TransportLAN {
			t.Errorf("second event transport = %v, want lan", evt.Transport)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for connected event")
	}
}

func TestStop(t *testing.T) {
	e := newTestEngine(t)

	if err := e.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
	if e.State() != StateDisconnected {
		t.Errorf("after Stop, state = %v, want disconnected", e.State())
	}
}

func TestEventsChannel(t *testing.T) {
	e := newTestEngine(t)
	ch := e.Events()
	if ch == nil {
		t.Fatal("Events() returned nil channel")
	}
}

func TestStartStopIdempotent(t *testing.T) {
	e := newTestEngine(t)

	// Start twice
	if err := e.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := e.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Stop again should be fine
	if err := e.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
