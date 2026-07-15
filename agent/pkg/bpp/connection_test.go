package bpp

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type mockHandler struct {
	mu            sync.Mutex
	packets       []*Packet
	connectCalled bool
	disconnErr    error
}

func (h *mockHandler) HandlePacket(_ context.Context, pkt *Packet, _ *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.packets = append(h.packets, pkt)
}

func (h *mockHandler) OnConnect(_ context.Context, _ *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connectCalled = true
}

func (h *mockHandler) OnDisconnect(_ context.Context, _ *Connection, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.disconnErr = err
}

func TestConnStateString(t *testing.T) {
	tests := []struct {
		state ConnState
		want  string
	}{
		{ConnDisconnected, "disconnected"},
		{ConnConnecting, "connecting"},
		{ConnAuthenticating, "authenticating"},
		{ConnReady, "ready"},
		{ConnState(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("ConnState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestNewConnectionDefaults(t *testing.T) {
	c := NewConnection("ws://localhost:8080", nil)
	if c.url != "ws://localhost:8080" {
		t.Errorf("url = %q", c.url)
	}
	if c.heartbeatInterval != DefaultHeartbeatInterval {
		t.Errorf("heartbeatInterval = %v, want %v", c.heartbeatInterval, DefaultHeartbeatInterval)
	}
	if c.heartbeatTimeout != DefaultHeartbeatTimeout {
		t.Errorf("heartbeatTimeout = %v, want %v", c.heartbeatTimeout, DefaultHeartbeatTimeout)
	}
	if c.writeTimeout != DefaultWriteTimeout {
		t.Errorf("writeTimeout = %v, want %v", c.writeTimeout, DefaultWriteTimeout)
	}
	if c.reconnectBase != DefaultReconnectBase {
		t.Errorf("reconnectBase = %v, want %v", c.reconnectBase, DefaultReconnectBase)
	}
	if c.reconnectMax != DefaultReconnectMax {
		t.Errorf("reconnectMax = %v, want %v", c.reconnectMax, DefaultReconnectMax)
	}
	if c.autoReconnect {
		t.Error("autoReconnect should be false by default")
	}
}

func TestWithHeartbeatInterval(t *testing.T) {
	c := NewConnection("ws://x", nil, WithHeartbeatInterval(5*time.Second))
	if c.heartbeatInterval != 5*time.Second {
		t.Errorf("got %v, want 5s", c.heartbeatInterval)
	}
}

func TestWithHeartbeatTimeout(t *testing.T) {
	c := NewConnection("ws://x", nil, WithHeartbeatTimeout(20*time.Second))
	if c.heartbeatTimeout != 20*time.Second {
		t.Errorf("got %v, want 20s", c.heartbeatTimeout)
	}
}

func TestWithWriteTimeout(t *testing.T) {
	c := NewConnection("ws://x", nil, WithWriteTimeout(3*time.Second))
	if c.writeTimeout != 3*time.Second {
		t.Errorf("got %v, want 3s", c.writeTimeout)
	}
}

func TestWithReconnectDefaults(t *testing.T) {
	c := NewConnection("ws://x", nil, WithReconnect(2*time.Second, 30*time.Second, 0.2))
	if c.reconnectBase != 2*time.Second {
		t.Errorf("reconnectBase = %v, want 2s", c.reconnectBase)
	}
	if c.reconnectMax != 30*time.Second {
		t.Errorf("reconnectMax = %v, want 30s", c.reconnectMax)
	}
	if c.reconnectJitter != 0.2 {
		t.Errorf("reconnectJitter = %v, want 0.2", c.reconnectJitter)
	}
	if !c.autoReconnect {
		t.Error("autoReconnect should be true")
	}
}

func TestWithReconnectEdgeCases(t *testing.T) {
	c := NewConnection("ws://x", nil, WithReconnect(0, 0, -1))
	if c.reconnectBase != DefaultReconnectBase {
		t.Errorf("base clamped: got %v, want %v", c.reconnectBase, DefaultReconnectBase)
	}
	if c.reconnectMax != DefaultReconnectMax {
		t.Errorf("max clamped: got %v, want %v", c.reconnectMax, DefaultReconnectMax)
	}
	if c.reconnectJitter != DefaultReconnectJitter {
		t.Errorf("jitter clamped: got %v, want %v", c.reconnectJitter, DefaultReconnectJitter)
	}
}

func TestWithReconnectJitterHigh(t *testing.T) {
	c := NewConnection("ws://x", nil, WithReconnect(time.Second, time.Second, 1.0))
	if c.reconnectJitter != DefaultReconnectJitter {
		t.Errorf("jitter clamped: got %v, want %v", c.reconnectJitter, DefaultReconnectJitter)
	}
}

func TestWithDialer(t *testing.T) {
	d := &websocket.Dialer{}
	c := NewConnection("ws://x", nil, WithDialer(d))
	if c.dialer != d {
		t.Error("dialer not set")
	}
}

func TestWithHeader(t *testing.T) {
	h := http.Header{"X-Test": {"val"}}
	c := NewConnection("ws://x", nil, WithHeader(h))
	if c.header.Get("X-Test") != "val" {
		t.Error("header not set")
	}
}

func TestWithLogger(t *testing.T) {
	l := slog.Default()
	c := NewConnection("ws://x", nil, WithLogger(l))
	if c.log != l {
		t.Error("logger not set")
	}
}

func TestWithLoggerNil(t *testing.T) {
	c := NewConnection("ws://x", nil)
	if c.log == nil {
		t.Error("log should not be nil")
	}
}

func TestConnectionStats(t *testing.T) {
	c := NewConnection("ws://x", nil)
	s := c.Stats()
	if s.State != ConnDisconnected {
		t.Errorf("State = %v, want ConnDisconnected", s.State)
	}
	if !s.ConnectedAt.IsZero() {
		t.Error("ConnectedAt should be zero")
	}
}

func TestConnectionIsConnected(t *testing.T) {
	c := NewConnection("ws://x", nil)
	if c.IsConnected() {
		t.Error("should not be connected")
	}
}

func TestConnectionLocalAddr(t *testing.T) {
	c := NewConnection("ws://x", nil)
	if got := c.LocalAddr(); got != "" {
		t.Errorf("LocalAddr = %q, want empty", got)
	}
}

func TestConnectionRemoteAddr(t *testing.T) {
	c := NewConnection("ws://x", nil)
	if got := c.RemoteAddr(); got != "" {
		t.Errorf("RemoteAddr = %q, want empty", got)
	}
}

func TestConnectionSetState(t *testing.T) {
	c := NewConnection("ws://x", nil)
	c.setState(ConnReady)
	if c.Stats().State != ConnReady {
		t.Errorf("State = %v, want ConnReady", c.Stats().State)
	}
}

func TestConnectionCloseNoConn(t *testing.T) {
	c := NewConnection("ws://x", nil)
	if err := c.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestBackoff(t *testing.T) {
	c := NewConnection("ws://x", nil,
		WithReconnect(1*time.Second, 10*time.Second, 0))

	d1 := c.backoff(1)
	d2 := c.backoff(2)
	d3 := c.backoff(3)

	if d1 != 1*time.Second {
		t.Errorf("backoff(1) = %v, want 1s", d1)
	}
	if d2 != 2*time.Second {
		t.Errorf("backoff(2) = %v, want 2s", d2)
	}
	if d3 != 4*time.Second {
		t.Errorf("backoff(3) = %v, want 4s", d3)
	}
}

func TestBackoffMaxCap(t *testing.T) {
	c := NewConnection("ws://x", nil,
		WithReconnect(1*time.Second, 5*time.Second, 0))

	d := c.backoff(10)
	if d > 5*time.Second {
		t.Errorf("backoff(10) = %v, should be capped at 5s", d)
	}
}

func TestBackoffAttemptZero(t *testing.T) {
	c := NewConnection("ws://x", nil,
		WithReconnect(1*time.Second, 10*time.Second, 0))

	d0 := c.backoff(0)
	dNeg := c.backoff(-1)
	if d0 != 1*time.Second {
		t.Errorf("backoff(0) = %v, want 1s", d0)
	}
	if dNeg != 1*time.Second {
		t.Errorf("backoff(-1) = %v, want 1s", dNeg)
	}
}

func TestBackoffWithJitter(t *testing.T) {
	c := NewConnection("ws://x", nil,
		WithReconnect(1*time.Second, 60*time.Second, 0.5))

	for i := 0; i < 20; i++ {
		d := c.backoff(2)
		if d < 0 {
			t.Errorf("backoff(2) = %v, should be non-negative", d)
		}
	}
}

type wsTestServer struct {
	server *httptest.Server
	conn   *websocket.Conn
	mu     sync.Mutex
}

func newWSTestServer(t *testing.T) *wsTestServer {
	t.Helper()
	ts := &wsTestServer{}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	ts.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		ts.mu.Lock()
		ts.conn = c
		ts.mu.Unlock()
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				return
			}
		}
	}))
	return ts
}

func (ts *wsTestServer) closeConn() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.conn != nil {
		ts.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(time.Second),
		)
		ts.conn.Close()
		ts.conn = nil
	}
}

func (ts *wsTestServer) close() {
	ts.closeConn()
	ts.server.Close()
}

func TestDialAndClose(t *testing.T) {
	ts := newWSTestServer(t)
	defer ts.close()

	wsURL := "ws" + ts.server.URL[4:]
	h := &mockHandler{}
	c := NewConnection(wsURL, h)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Dial(ctx); err != nil {
		t.Fatalf("Dial() error = %v", err)
	}

	if !c.IsConnected() {
		t.Error("should be connected after Dial()")
	}
	if c.LocalAddr() == "" {
		t.Error("LocalAddr should be non-empty after Dial()")
	}
	if c.RemoteAddr() == "" {
		t.Error("RemoteAddr should be non-empty after Dial()")
	}
	if c.Stats().ConnectedAt.IsZero() {
		t.Error("ConnectedAt should be set after Dial()")
	}

	ts.closeConn()
	if err := c.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
	if c.IsConnected() {
		t.Error("should not be connected after Close()")
	}
}

func TestDialFailure(t *testing.T) {
	c := NewConnection("ws://127.0.0.1:1", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Dial(ctx)
	if err == nil {
		t.Fatal("Dial() to invalid address should fail")
	}
}

func TestSendPacketAndSendJSON(t *testing.T) {
	ts := newWSTestServer(t)
	defer ts.close()

	wsURL := "ws" + ts.server.URL[4:]
	c := NewConnection(wsURL, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Dial(ctx); err != nil {
		t.Fatalf("Dial() error = %v", err)
	}

	pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("hello"))
	if err := c.SendPacket(pkt); err != nil {
		t.Errorf("SendPacket() error = %v", err)
	}

	env, _ := NewRequest("test.method", map[string]string{"key": "val"})
	if err := c.SendJSON(env); err != nil {
		t.Errorf("SendJSON() error = %v", err)
	}
}

func TestSendPacketNotConnected(t *testing.T) {
	c := NewConnection("ws://x", nil)
	pkt := NewPacket(PacketTypeRequest, ChannelRPC, []byte("test"))
	err := c.SendPacket(pkt)
	if err == nil {
		t.Error("SendPacket() should fail when not connected")
	}
}

func TestSendJSONNotConnected(t *testing.T) {
	c := NewConnection("ws://x", nil)
	env, _ := NewRequest("test", nil)
	err := c.SendJSON(env)
	if err == nil {
		t.Error("SendJSON() should fail when not connected")
	}
}

func TestServeHTTP(t *testing.T) {
	h := &mockHandler{}
	c := NewConnection("ws://x", h)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.ServeHTTP(w, r)
	}))
	defer server.Close()

	dialer := websocket.Dialer{}
	wsURL := "ws" + server.URL[4:]
	clientConn, _, err := dialer.DialContext(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatalf("client dial error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !c.IsConnected() {
		t.Error("should be connected after ServeHTTP upgrade")
	}

	clientConn.Close()
	c.Close()
}

func TestCloseWithActiveConn(t *testing.T) {
	ts := newWSTestServer(t)

	wsURL := "ws" + ts.server.URL[4:]
	c := NewConnection(wsURL, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Dial(ctx); err != nil {
		t.Fatalf("Dial() error = %v", err)
	}

	ts.closeConn()
	if err := c.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestConnStateUnknown(t *testing.T) {
	s := ConnState(42)
	if got := s.String(); got != "unknown" {
		t.Errorf("ConnState(42).String() = %q, want \"unknown\"", got)
	}
}

func TestBackoffWithJitterRange(t *testing.T) {
	c := NewConnection("ws://x", nil,
		WithReconnect(1*time.Second, 60*time.Second, 0.1))

	d := c.backoff(5)
	if d < 0 {
		t.Errorf("backoff(5) = %v, should be non-negative", d)
	}
}
