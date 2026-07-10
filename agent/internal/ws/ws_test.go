package ws

import (
	"context"
	"log/slog"
	"testing"
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
	// The upgrader allows all origins during development
	if !upgrader.CheckOrigin(nil) {
		t.Error("CheckOrigin should return true (development mode)")
	}
}
