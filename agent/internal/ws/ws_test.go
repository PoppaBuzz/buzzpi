package ws

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockHandler implements Handler for testing.
type mockHandler struct {
	onConnect    func(ctx context.Context, conn *Connection)
	onDisconnect func(ctx context.Context, conn *Connection)
}

func (m *mockHandler) HandleMessage(ctx context.Context, data []byte) ([]byte, error) {
	return nil, nil
}
func (m *mockHandler) OnConnect(ctx context.Context, conn *Connection) {
	if m.onConnect != nil {
		m.onConnect(ctx, conn)
	}
}
func (m *mockHandler) OnDisconnect(ctx context.Context, conn *Connection) {
	if m.onDisconnect != nil {
		m.onDisconnect(ctx, conn)
	}
}

func TestNewServer(t *testing.T) {
	s := NewServer(&mockHandler{}, 0, slog.Default())
	if s == nil {
		t.Fatal("NewServer() returned nil")
	}
	if s.Name() != "ws-server" {
		t.Errorf("Name() = %q, want \"ws-server\"", s.Name())
	}
}

func TestNewServerNilLogger(t *testing.T) {
	s := NewServer(&mockHandler{}, 0, nil)
	if s == nil {
		t.Fatal("NewServer(nil logger) returned nil")
	}
}

func TestServerHealth(t *testing.T) {
	s := NewServer(&mockHandler{}, 0, nil)
	h := s.Health()
	m, ok := h.(map[string]interface{})
	if !ok {
		t.Fatalf("Health() returned %T, want map[string]interface{}", h)
	}
	if m["status"] != "ok" {
		t.Errorf("Health() status = %v, want \"ok\"", m["status"])
	}
	if m["connections"] != 0 {
		t.Errorf("Health() connections = %v, want 0", m["connections"])
	}
}

func TestActiveConnections(t *testing.T) {
	s := NewServer(&mockHandler{}, 0, nil)
	if n := s.ActiveConnections(); n != 0 {
		t.Errorf("ActiveConnections() = %d, want 0", n)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("ws://localhost:8080/ws", slog.Default())
	if c == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestNewClientNilLogger(t *testing.T) {
	c := NewClient("ws://localhost:8080/ws", nil)
	if c == nil {
		t.Fatal("NewClient(nil logger) returned nil")
	}
}

func TestSendBeforeConnect(t *testing.T) {
	c := NewClient("ws://localhost:8080/ws", nil)
	err := c.Send([]byte("hello"))
	if err == nil {
		t.Error("Send() before Connect should return error")
	}
}

func TestCloseIdempotent(t *testing.T) {
	c := NewClient("ws://localhost:8080/ws", nil)
	// Close on unconnected client should not panic
	if err := c.Close(); err != nil {
		t.Fatalf("Close() on unconnected client failed: %v", err)
	}
	// Second close should also succeed
	if err := c.Close(); err != nil {
		t.Fatalf("Close() again failed: %v", err)
	}
}

func TestConnectionStruct(t *testing.T) {
	conn := &Connection{
		ID:         "conn_test",
		RemoteAddr: "192.168.1.5:54321",
	}
	if conn.ID != "conn_test" {
		t.Errorf("Connection.ID = %q", conn.ID)
	}
	if conn.RemoteAddr != "192.168.1.5:54321" {
		t.Errorf("Connection.RemoteAddr = %q", conn.RemoteAddr)
	}
	if conn.Conn != nil {
		t.Error("Connection.Conn should be nil (not connected)")
	}
}

func TestUpgraderCheckOrigin(t *testing.T) {
	if !upgrader.CheckOrigin(nil) {
		t.Error("CheckOrigin should return true (development mode)")
	}
}

func TestContextKeys(t *testing.T) {
	if ContextKeyConnID == "" {
		t.Error("ContextKeyConnID should not be empty")
	}
	if ContextKeyConn == "" {
		t.Error("ContextKeyConn should not be empty")
	}
	if ContextKeyConnID == ContextKeyConn {
		t.Error("ContextKeyConnID and ContextKeyConn should be different")
	}
}

func TestServerStartAndStop(t *testing.T) {
	s := NewServer(&mockHandler{}, 0, slog.Default())
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start(ctx)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Logf("Start returned: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestServerStopWithoutStart(t *testing.T) {
	s := NewServer(&mockHandler{}, 0, slog.Default())
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() without Start should not error: %v", err)
	}
}

func TestServerHandleHealthEndpoint(t *testing.T) {
	s := NewServer(&mockHandler{}, 0, slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Start(ctx)
	time.Sleep(200 * time.Millisecond)

	// Find the actual listener addr
}

func TestServerWebSocketConnection(t *testing.T) {
	handler := &mockHandler{}
	s := NewServer(handler, 0, slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer s.Stop(context.Background())

	time.Sleep(200 * time.Millisecond)

	if s.httpServer == nil {
		t.Fatal("httpServer not initialized")
	}
}

func TestConnectionID(t *testing.T) {
	conn := &Connection{
		ID:         "test_conn_123",
		RemoteAddr: "10.0.0.1:12345",
	}
	if conn.ID != "test_conn_123" {
		t.Errorf("ID = %q", conn.ID)
	}
	if conn.RemoteAddr != "10.0.0.1:12345" {
		t.Errorf("RemoteAddr = %q", conn.RemoteAddr)
	}
	if conn.ConnectedAt.IsZero() {
		t.Log("ConnectedAt is zero (expected for manually created)")
	}
}

func TestServerMultipleHealthChecks(t *testing.T) {
	s := NewServer(&mockHandler{}, 0, slog.Default())
	for i := 0; i < 10; i++ {
		h := s.Health()
		if h == nil {
			t.Fatalf("Health() returned nil on iteration %d", i)
		}
	}
}

func TestClientConnectAndClose(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader.Upgrade(w, r, nil)
	}))
	defer ts.Close()

	wsURL := "ws" + ts.URL[len("http"):] + "/ws"
	c := NewClient(wsURL, slog.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	if err := c.Send([]byte("hello")); err != nil {
		t.Fatalf("Send() failed: %v", err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
}

func TestClientConnectInvalidURL(t *testing.T) {
	c := NewClient("ws://127.0.0.1:1/ws", slog.Default())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Connect(ctx)
	if err == nil {
		t.Error("Connect() to invalid port should fail")
		c.Close()
	}
}

func TestServerWithNilHandler(t *testing.T) {
	s := NewServer(nil, 0, slog.Default())
	ts := httptest.NewServer(http.HandlerFunc(s.handleWebSocket))
	defer ts.Close()

	wsURL := "ws" + ts.URL[len("http"):] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	conn.WriteMessage(websocket.TextMessage, []byte("test"))
	time.Sleep(100 * time.Millisecond)
}

func TestServerMessageHandlerError(t *testing.T) {
	handler := &errorHandler{}
	s := NewServer(handler, 0, slog.Default())
	ts := httptest.NewServer(http.HandlerFunc(s.handleWebSocket))
	defer ts.Close()

	wsURL := "ws" + ts.URL[len("http"):] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	time.Sleep(100 * time.Millisecond)

	err = conn.WriteMessage(websocket.TextMessage, []byte("test message"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestServerHandlerReturnsResponse(t *testing.T) {
	handler := &responseHandler{}
	s := NewServer(handler, 0, slog.Default())
	ts := httptest.NewServer(http.HandlerFunc(s.handleWebSocket))
	defer ts.Close()

	wsURL := "ws" + ts.URL[len("http"):] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	time.Sleep(100 * time.Millisecond)

	err = conn.WriteMessage(websocket.TextMessage, []byte("test"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if string(msg) != "response" {
		t.Errorf("response = %q, want %q", string(msg), "response")
	}
}

type errorHandler struct{}

func (e *errorHandler) HandleMessage(ctx context.Context, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("handler error")
}
func (e *errorHandler) OnConnect(ctx context.Context, conn *Connection)    {}
func (e *errorHandler) OnDisconnect(ctx context.Context, conn *Connection) {}

type responseHandler struct{}

func (r *responseHandler) HandleMessage(ctx context.Context, data []byte) ([]byte, error) {
	return []byte("response"), nil
}
func (r *responseHandler) OnConnect(ctx context.Context, conn *Connection)    {}
func (r *responseHandler) OnDisconnect(ctx context.Context, conn *Connection) {}

func TestServerOnConnectOnDisconnect(t *testing.T) {
	handler := &lifecycleHandler{}
	s := NewServer(handler, 0, slog.Default())
	ts := httptest.NewServer(http.HandlerFunc(s.handleWebSocket))
	defer ts.Close()

	wsURL := "ws" + ts.URL[len("http"):] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !handler.connected {
		t.Error("OnConnect not called")
	}

	conn.Close()
	time.Sleep(200 * time.Millisecond)

	if !handler.disconnected {
		t.Error("OnDisconnect not called")
	}

	if handler.connCount != 1 {
		t.Errorf("connection count = %d, want 1", handler.connCount)
	}
}

type lifecycleHandler struct {
	connected    bool
	disconnected bool
	connCount    int
}

func (l *lifecycleHandler) HandleMessage(ctx context.Context, data []byte) ([]byte, error) {
	return nil, nil
}
func (l *lifecycleHandler) OnConnect(ctx context.Context, conn *Connection) {
	l.connected = true
	l.connCount++
}
func (l *lifecycleHandler) OnDisconnect(ctx context.Context, conn *Connection) {
	l.disconnected = true
}
