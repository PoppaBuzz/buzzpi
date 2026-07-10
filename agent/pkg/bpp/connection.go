// Package bpp implements the BuzzPi Protocol (BPP) wire format.
//
// Connection provides a WebSocket transport that sends and receives BPP
// binary frames (Packet). It handles connection lifecycle, heartbeat keep-
// alive, concurrent I/O, and automatic reconnection with exponential backoff.
package bpp

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Default connection parameters.
const (
	DefaultHeartbeatInterval = 30 * time.Second
	DefaultHeartbeatTimeout  = 90 * time.Second
	DefaultWriteTimeout      = 10 * time.Second
	DefaultReadLimit         = 1 << 24 // 16 MB
	DefaultReconnectBase     = 1 * time.Second
	DefaultReconnectMax      = 60 * time.Second
	DefaultReconnectJitter   = 0.1
)

// Handler processes incoming BPP packets from a Connection.
type Handler interface {
	// HandlePacket processes a received BPP packet.
	// If the packet is a TypeRequest, the handler should send a response
	// or error packet back through the connection.
	HandlePacket(ctx context.Context, pkt *Packet, conn *Connection)

	// OnConnect is called when the connection is established (or re-established).
	OnConnect(ctx context.Context, conn *Connection)

	// OnDisconnect is called when the connection is lost.
	OnDisconnect(ctx context.Context, conn *Connection, err error)
}

// ConnState represents the state of a BPP WebSocket connection.
type ConnState int

const (
	ConnDisconnected ConnState = iota
	ConnConnecting
	ConnAuthenticating
	ConnReady
)

func (s ConnState) String() string {
	switch s {
	case ConnDisconnected:
		return "disconnected"
	case ConnConnecting:
		return "connecting"
	case ConnAuthenticating:
		return "authenticating"
	case ConnReady:
		return "ready"
	default:
		return "unknown"
	}
}

// ConnStats exposes runtime connection statistics.
type ConnStats struct {
	State          ConnState
	ConnectedAt    time.Time
	LastHeartbeat  time.Time
	PacketsSent    uint64
	PacketsRecv    uint64
	BytesSent      uint64
	BytesRecv      uint64
	ReconnectCount uint64
	RTTMs          int64
}

// Connection is a BPP-over-WebSocket transport.
//
// A Connection wraps a gorilla/websocket.Conn and provides:
//   - BPP binary packet send/receive
//   - Heartbeat keep-alive (ping/pong)
//   - Concurrent-safe send via internal mutex
//   - Read pump that dispatches to a Handler
//   - Optional automatic reconnection with exponential backoff
type Connection struct {
	// Dial configuration.
	url    string
	header http.Header
	dialer *websocket.Dialer

	// WebSocket connection (nil when disconnected).
	conn   *websocket.Conn
	connMu sync.Mutex

	// Handler for incoming packets.
	handler Handler

	// Heartbeat parameters.
	heartbeatInterval time.Duration
	heartbeatTimeout  time.Duration

	// Write timeout.
	writeTimeout time.Duration

	// Reconnection parameters.
	reconnectBase   time.Duration
	reconnectMax    time.Duration
	reconnectJitter float64
	autoReconnect   bool

	// Read pump goroutine control.
	cancelRead context.CancelFunc
	readWg     sync.WaitGroup

	// Heartbeat goroutine control.
	cancelHeartbeat context.CancelFunc
	heartbeatWg     sync.WaitGroup

	// Statistics.
	stats       ConnStats
	statsMu     sync.RWMutex
	connectedAt time.Time
	lastHb      time.Time

	// Logger.
	log *slog.Logger

	// Lifecycle.
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ConnectionOption configures a Connection.
type ConnectionOption func(*Connection)

// WithHeartbeatInterval sets the heartbeat ping interval.
func WithHeartbeatInterval(d time.Duration) ConnectionOption {
	return func(c *Connection) { c.heartbeatInterval = d }
}

// WithHeartbeatTimeout sets the heartbeat timeout (time before declaring dead).
func WithHeartbeatTimeout(d time.Duration) ConnectionOption {
	return func(c *Connection) { c.heartbeatTimeout = d }
}

// WithWriteTimeout sets the WebSocket write deadline.
func WithWriteTimeout(d time.Duration) ConnectionOption {
	return func(c *Connection) { c.writeTimeout = d }
}

// WithReconnect configures automatic reconnection with exponential backoff.
func WithReconnect(base, max time.Duration, jitter float64) ConnectionOption {
	if base <= 0 {
		base = DefaultReconnectBase
	}
	if max <= 0 {
		max = DefaultReconnectMax
	}
	if jitter < 0 || jitter > 0.5 {
		jitter = DefaultReconnectJitter
	}
	return func(c *Connection) {
		c.reconnectBase = base
		c.reconnectMax = max
		c.reconnectJitter = jitter
		c.autoReconnect = true
	}
}

// WithDialer sets a custom WebSocket dialer.
func WithDialer(d *websocket.Dialer) ConnectionOption {
	return func(c *Connection) { c.dialer = d }
}

// WithHeader sets HTTP headers for the WebSocket upgrade request.
func WithHeader(h http.Header) ConnectionOption {
	return func(c *Connection) { c.header = h }
}

// WithLogger sets a logger for the connection.
func WithLogger(log *slog.Logger) ConnectionOption {
	return func(c *Connection) {
		if log != nil {
			c.log = log
		}
	}
}

// NewConnection creates a new BPP WebSocket Connection.
//
// The connection does not connect until Dial is called.
func NewConnection(url string, handler Handler, opts ...ConnectionOption) *Connection {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Connection{
		url:               url,
		handler:           handler,
		dialer:            websocket.DefaultDialer,
		header:            make(http.Header),
		heartbeatInterval: DefaultHeartbeatInterval,
		heartbeatTimeout:  DefaultHeartbeatTimeout,
		writeTimeout:      DefaultWriteTimeout,
		reconnectBase:     DefaultReconnectBase,
		reconnectMax:      DefaultReconnectMax,
		reconnectJitter:   DefaultReconnectJitter,
		autoReconnect:     false,
		log:               slog.Default().With("component", "bpp-conn", "url", url),
		ctx:               ctx,
		cancel:            cancel,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Dial establishes the WebSocket connection and starts the read pump
// and heartbeat goroutines. Blocks until connected or an error occurs.
func (c *Connection) Dial(ctx context.Context) error {
	c.setState(ConnConnecting)

	conn, _, err := c.dialer.DialContext(ctx, c.url, c.header)
	if err != nil {
		c.setState(ConnDisconnected)
		return fmt.Errorf("bpp: dial %s: %w", c.url, err)
	}

	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()

	now := time.Now()
	c.statsMu.Lock()
	c.connectedAt = now
	c.lastHb = now
	c.stats.ReconnectCount = 0
	c.statsMu.Unlock()

	// Set up WebSocket read limit.
	conn.SetReadLimit(DefaultReadLimit)

	// Start the read pump.
	readCtx, readCancel := context.WithCancel(c.ctx)
	c.cancelRead = readCancel
	c.readWg.Add(1)
	go c.readPump(readCtx)

	// Start heartbeat.
	hbCtx, hbCancel := context.WithCancel(c.ctx)
	c.cancelHeartbeat = hbCancel
	c.heartbeatWg.Add(1)
	go c.heartbeatLoop(hbCtx)

	c.setState(ConnReady)

	if c.handler != nil {
		c.handler.OnConnect(ctx, c)
	}

	c.log.Info("connected", "url", c.url)
	return nil
}

// Close gracefully shuts down the connection.
func (c *Connection) Close() error {
	c.cancel()
	c.autoReconnect = false // prevent reconnect loop

	// Stop heartbeat.
	if c.cancelHeartbeat != nil {
		c.cancelHeartbeat()
	}
	c.heartbeatWg.Wait()

	// Stop read pump.
	if c.cancelRead != nil {
		c.cancelRead()
	}
	c.readWg.Wait()

	// Close WebSocket.
	c.connMu.Lock()
	if c.conn != nil {
		_ = c.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(c.writeTimeout),
		)
		c.conn.Close()
		c.conn = nil
	}
	c.connMu.Unlock()

	c.setState(ConnDisconnected)
	c.log.Info("disconnected")
	return nil
}

// SendPacket marshals and writes a BPP packet to the WebSocket connection.
// Thread-safe. Returns an error if the connection is not established.
func (c *Connection) SendPacket(pkt *Packet) error {
	data, err := pkt.Marshal()
	if err != nil {
		return fmt.Errorf("bpp: marshal packet: %w", err)
	}

	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("bpp: not connected")
	}

	if c.writeTimeout > 0 {
		_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}

	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		c.log.Warn("write error", "error", err)
		return fmt.Errorf("bpp: write: %w", err)
	}

	c.statsMu.Lock()
	c.stats.PacketsSent++
	c.stats.BytesSent += uint64(len(data))
	c.statsMu.Unlock()

	return nil
}

// SendJSON is a helper that wraps an Envelope in a BPP packet and sends it
// over the reliable RPC channel.
func (c *Connection) SendJSON(env *Envelope) error {
	payload, err := env.Marshal()
	if err != nil {
		return err
	}
	pkt := NewPacket(PacketTypeRequest, ChannelRPC, payload)
	// Text payload if the outer type is event
	if env.Type == TypeEvent {
		pkt.Header.PType = PacketTypeEvent
		pkt.Header.Channel = ChannelEvent
	}
	return c.SendPacket(pkt)
}

// readPump reads binary messages from the WebSocket, deserializes them into
// BPP packets, and dispatches them to the Handler.
func (c *Connection) readPump(ctx context.Context) {
	defer c.readWg.Done()
	defer c.notifyDisconnect(fmt.Errorf("read pump exited"))

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		c.connMu.Lock()
		conn := c.conn
		c.connMu.Unlock()

		if conn == nil {
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			// Normal closure — not an error.
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.log.Debug("connection closed normally")
				return
			}
			// Unexpected error — trigger reconnect if configured.
			c.log.Warn("read error", "error", err)
			c.handleDisconnect(err)
			return
		}

		pkt, err := UnmarshalPacket(data)
		if err != nil {
			c.log.Warn("invalid packet", "error", err)
			continue
		}

		c.statsMu.Lock()
		c.stats.PacketsRecv++
		c.stats.BytesRecv += uint64(len(data))
		c.statsMu.Unlock()

		// Handle heartbeats at the transport layer.
		if pkt.Header.PType == PacketTypeHeartbeat {
			c.handleHeartbeat(pkt)
			continue
		}
		if pkt.Header.PType == PacketTypeHeartbeatAck {
			c.handleHeartbeatAck()
			continue
		}

		// Dispatch to the application handler.
		if c.handler != nil {
			c.handler.HandlePacket(ctx, pkt, c)
		}
	}
}

// heartbeatLoop sends periodic ping frames and checks for pong responses.
func (c *Connection) heartbeatLoop(ctx context.Context) {
	defer c.heartbeatWg.Done()

	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.sendHeartbeat()
		}
	}
}

func (c *Connection) sendHeartbeat() {
	pkt := NewControlPacket(PacketTypeHeartbeat)
	if err := c.SendPacket(pkt); err != nil {
		c.log.Warn("heartbeat send failed", "error", err)
	}

	c.statsMu.Lock()
	c.lastHb = time.Now()
	c.statsMu.Unlock()
}

func (c *Connection) handleHeartbeat(pkt *Packet) {
	// Respond with a heartbeat ack.
	ack := NewControlPacket(PacketTypeHeartbeatAck)
	_ = c.SendPacket(ack)
}

func (c *Connection) handleHeartbeatAck() {
	c.statsMu.Lock()
	c.lastHb = time.Now()
	c.statsMu.Unlock()
}

func (c *Connection) handleDisconnect(err error) {
	c.connMu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connMu.Unlock()

	c.notifyDisconnect(err)

	if c.autoReconnect && c.ctx.Err() == nil {
		go c.reconnectLoop()
	}
}

func (c *Connection) notifyDisconnect(err error) {
	c.setState(ConnDisconnected)
	if c.handler != nil {
		c.handler.OnDisconnect(c.ctx, c, err)
	}
}

// reconnectLoop attempts to reconnect with exponential backoff.
func (c *Connection) reconnectLoop() {
	attempt := 0

	for {
		// Check if we should still be reconnecting.
		if !c.autoReconnect || c.ctx.Err() != nil {
			return
		}

		attempt++
		delay := c.backoff(attempt)

		c.log.Info("reconnecting",
			"attempt", attempt,
			"delay", delay,
		)

		select {
		case <-c.ctx.Done():
			return
		case <-time.After(delay):
		}

		// Try to reconnect.
		conn, _, err := c.dialer.DialContext(c.ctx, c.url, c.header)
		if err != nil {
			c.log.Warn("reconnect failed", "attempt", attempt, "error", err)
			continue
		}

		// Reconnect succeeded.
		c.connMu.Lock()
		c.conn = conn
		c.connMu.Unlock()

		now := time.Now()
		c.statsMu.Lock()
		c.stats.ReconnectCount++
		c.connectedAt = now
		c.lastHb = now
		c.statsMu.Unlock()

		conn.SetReadLimit(DefaultReadLimit)

		// Restart read pump.
		readCtx, readCancel := context.WithCancel(c.ctx)
		c.cancelRead = readCancel
		c.readWg.Add(1)
		go c.readPump(readCtx)

		c.setState(ConnReady)

		if c.handler != nil {
			c.handler.OnConnect(c.ctx, c)
		}

		c.log.Info("reconnected", "url", c.url, "attempts", attempt)
		return
	}
}

// backoff computes the next reconnect delay with jitter.
func (c *Connection) backoff(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	// Full jitter: sleep = random(0, min(max, base * 2^attempt))
	exp := float64(c.reconnectBase) * math.Pow(2, float64(attempt-1))
	max := float64(c.reconnectMax)
	if exp > max {
		exp = max
	}
	jitter := exp * c.reconnectJitter
	delay := exp - jitter + rand.Float64()*jitter*2
	return time.Duration(delay)
}

// setState updates the connection state.
func (c *Connection) setState(s ConnState) {
	c.statsMu.Lock()
	c.stats.State = s
	c.statsMu.Unlock()
}

// Stats returns a snapshot of connection statistics.
func (c *Connection) Stats() ConnStats {
	c.statsMu.RLock()
	defer c.statsMu.RUnlock()
	s := c.stats
	s.ConnectedAt = c.connectedAt
	s.LastHeartbeat = c.lastHb
	return s
}

// IsConnected returns true if the connection is in the Ready state.
func (c *Connection) IsConnected() bool {
	c.statsMu.RLock()
	defer c.statsMu.RUnlock()
	return c.stats.State == ConnReady
}

// LocalAddr returns the local side of the connection, if connected.
func (c *Connection) LocalAddr() string {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn != nil {
		return c.conn.LocalAddr().String()
	}
	return ""
}

// RemoteAddr returns the remote side of the connection, if connected.
func (c *Connection) RemoteAddr() string {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return ""
}

// ServeHTTP implements http.Handler for server-side WebSocket upgrades.
// Use this to accept incoming BPP connections from an HTTP server.
func (c *Connection) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.log.Error("upgrade failed", "error", err)
		return
	}

	c.connMu.Lock()
	// Close any existing connection.
	if c.conn != nil {
		c.conn.Close()
	}
	c.conn = conn
	c.connMu.Unlock()

	now := time.Now()
	c.statsMu.Lock()
	c.connectedAt = now
	c.lastHb = now
	c.stats.ReconnectCount = 0
	c.statsMu.Unlock()

	conn.SetReadLimit(DefaultReadLimit)

	// Start read pump.
	readCtx, readCancel := context.WithCancel(c.ctx)
	c.cancelRead = readCancel
	c.readWg.Add(1)
	go c.readPump(readCtx)

	// Start heartbeat.
	hbCtx, hbCancel := context.WithCancel(c.ctx)
	c.cancelHeartbeat = hbCancel
	c.heartbeatWg.Add(1)
	go c.heartbeatLoop(hbCtx)

	c.setState(ConnReady)

	if c.handler != nil {
		c.handler.OnConnect(r.Context(), c)
	}

	c.log.Info("accepted server-side connection", "remote", r.RemoteAddr)
}

// Ensure Connection implements io.Closer.
var _ io.Closer = (*Connection)(nil)
