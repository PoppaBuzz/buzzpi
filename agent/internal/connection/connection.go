// Package connection implements the Connection Engine — the transport layer
// that manages client-to-device connections. It probes available transports
// in priority order (LAN > P2P > Relay) and transitions transparently.
package connection

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// TransportType represents a connection transport method.
type TransportType int

const (
	TransportNone  TransportType = iota
	TransportLAN                 // Direct WebSocket on local network
	TransportP2P                 // WebRTC peer-to-peer
	TransportRelay               // Cloud Relay WebSocket
)

func (t TransportType) String() string {
	switch t {
	case TransportLAN:
		return "lan"
	case TransportP2P:
		return "p2p"
	case TransportRelay:
		return "relay"
	default:
		return "none"
	}
}

// ConnectionState represents the state of a device connection.
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateDiscovering
	StateConnecting
	StateAuthenticating
	StateConnected
	StateDegraded
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateDiscovering:
		return "discovering"
	case StateConnecting:
		return "connecting"
	case StateAuthenticating:
		return "authenticating"
	case StateConnected:
		return "connected"
	case StateDegraded:
		return "degraded"
	default:
		return "unknown"
	}
}

// QualityLevel represents the connection quality.
type QualityLevel int

const (
	QualityUnknown QualityLevel = iota
	QualityHigh
	QualityMedium
	QualityLow
	QualityMinimum
)

func (q QualityLevel) String() string {
	switch q {
	case QualityHigh:
		return "high"
	case QualityMedium:
		return "medium"
	case QualityLow:
		return "low"
	case QualityMinimum:
		return "minimum"
	default:
		return "unknown"
	}
}

// QualityMetrics contains network quality measurements.
type QualityMetrics struct {
	RTTMs         int     // Round-trip time in milliseconds
	JitterMs      int     // Jitter in milliseconds
	PacketLoss    float64 // Packet loss percentage (0.0 - 1.0)
	BandwidthKbps int     // Estimated available bandwidth in Kbps
	Transport     TransportType
}

// ConnectionEvent represents a change in connection state.
type ConnectionEvent struct {
	DeviceID      string
	State         ConnectionState
	Transport     TransportType
	Quality       QualityLevel
	PreviousState ConnectionState
	Error         error
}

// Engine manages the connection lifecycle for a device.
type Engine struct {
	deviceID  string
	state     ConnectionState
	transport TransportType
	quality   QualityLevel
	metrics   QualityMetrics
	eventCh   chan ConnectionEvent
	mu        sync.RWMutex
	log       *slog.Logger
	stopCh    chan struct{}
	stopOnce  sync.Once
}

// NewEngine creates a new Connection Engine for a device.
func NewEngine(deviceID string, log *slog.Logger) *Engine {
	if log == nil {
		log = slog.Default()
	}
	return &Engine{
		deviceID:  deviceID,
		state:     StateDisconnected,
		transport: TransportNone,
		quality:   QualityUnknown,
		eventCh:   make(chan ConnectionEvent, 32),
		log:       log.With("component", "connection-engine", "device", deviceID),
		stopCh:    make(chan struct{}),
	}
}

// Events returns a channel of connection state changes.
func (e *Engine) Events() <-chan ConnectionEvent {
	return e.eventCh
}

// Start begins the connection process.
func (e *Engine) Start(ctx context.Context) error {
	e.log.Info("connection engine starting")

	// TODO: Probe available transports in priority order:
	// 1. LAN (mDNS + direct WebSocket)
	// 2. P2P (WebRTC ICE)
	// 3. Relay (cloud WebSocket)

	e.setState(StateDiscovering)

	go e.probeTransports(ctx)

	return nil
}

// Stop shuts down the connection engine.
func (e *Engine) Stop(ctx context.Context) error {
	e.log.Info("connection engine stopping")
	e.stopOnce.Do(func() {
		close(e.stopCh)
	})
	e.setState(StateDisconnected)
	return nil
}

func (e *Engine) probeTransports(ctx context.Context) {
	e.log.Info("probing available transports")

	// TODO: Implement transport probing:
	// 1. Try mDNS discovery for LAN
	// 2. If paired, try ICE for P2P
	// 3. Fall back to Cloud Relay

	// For now, just simulate a successful LAN connection after a brief delay
	select {
	case <-time.After(100 * time.Millisecond):
		e.setTransport(TransportLAN)
		e.setQuality(QualityHigh)
		e.setState(StateConnected)
		e.log.Info("connected via LAN")
	case <-ctx.Done():
		return
	case <-e.stopCh:
		return
	}
}

func (e *Engine) setState(s ConnectionState) {
	e.mu.Lock()
	prev := e.state
	e.state = s
	e.mu.Unlock()

	// Non-blocking send to event channel
	select {
	case e.eventCh <- ConnectionEvent{
		DeviceID:      e.deviceID,
		State:         s,
		Transport:     e.transport,
		Quality:       e.quality,
		PreviousState: prev,
	}:
	default:
	}
}

func (e *Engine) setTransport(t TransportType) {
	e.mu.Lock()
	e.transport = t
	e.mu.Unlock()
}

func (e *Engine) setQuality(q QualityLevel) {
	e.mu.Lock()
	e.quality = q
	e.mu.Unlock()
}

// State returns the current connection state.
func (e *Engine) State() ConnectionState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

// Transport returns the current transport type.
func (e *Engine) Transport() TransportType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.transport
}

// Quality returns the current quality level.
func (e *Engine) Quality() QualityLevel {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.quality
}

// Name returns the component name for the Supervisor.
func (e *Engine) Name() string { return fmt.Sprintf("connection-%s", e.deviceID) }

// Health returns the engine health status.
func (e *Engine) Health() interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return map[string]interface{}{
		"state":     e.state.String(),
		"transport": e.transport.String(),
		"quality":   e.quality.String(),
	}
}

// StateMachine diagram (for documentation):
//
//	Disconnected
//	    │
//	    ▼
//	Discovering ←──┐
//	    │          │
//	    ▼          │
//	Connecting     │ (retry)
//	    │          │
//	    ▼          │
//	Authenticating │
//	    │          │
//	    ├─── OK ───┤
//	    ▼          │
//	Connected ─────┘
//	    │
//	    ▼
//	Degraded ──→ Connected (improved)
