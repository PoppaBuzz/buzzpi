package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"
)

type mockProcess struct{}

func (p *mockProcess) Kill() error { return nil }
func (p *mockProcess) Wait() error { return nil }

type mockPtyFile struct {
	data []byte
}

func (f *mockPtyFile) Read(b []byte) (int, error)  { return 0, nil }
func (f *mockPtyFile) Write(b []byte) (int, error) { return len(b), nil }
func (f *mockPtyFile) Close() error                { return nil }

func newTestSession(id string) *Session {
	return &Session{
		ID:        id,
		CreatedAt: time.Now(),
		cmd:       &mockProcess{},
		pty:       &mockPtyFile{},
	}
}

func injectSession(m *Manager, s *Session) {
	m.sessions.Store(s.ID, s)
}

func TestNewManager(t *testing.T) {
	m := NewManager(nil)
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestNewManagerWithLogger(t *testing.T) {
	m := NewManager(slog.Default())
	if m == nil {
		t.Fatal("NewManager(slog.Default()) returned nil")
	}
}

func TestManagerName(t *testing.T) {
	m := NewManager(nil)
	if m.Name() != "terminal" {
		t.Errorf("Name() = %q, want \"terminal\"", m.Name())
	}
}

func TestManagerStart(t *testing.T) {
	m := NewManager(nil)
	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
}

func TestManagerStop(t *testing.T) {
	m := NewManager(nil)
	if err := m.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestManagerCount(t *testing.T) {
	m := NewManager(nil)
	if m.Count() != 0 {
		t.Errorf("Count() = %d, want 0 (empty manager)", m.Count())
	}
}

func TestManagerHealth(t *testing.T) {
	m := NewManager(nil)
	h := m.Health()
	m2, ok := h.(map[string]interface{})
	if !ok {
		t.Fatalf("Health() returned %T, want map[string]interface{}", h)
	}
	if m2["status"] != "ok" {
		t.Errorf("Health() status = %v, want \"ok\"", m2["status"])
	}
	if m2["sessions"] != 0 {
		t.Errorf("Health() sessions = %v, want 0", m2["sessions"])
	}
}

func TestManagerOpen(t *testing.T) {
	m := NewManager(slog.Default())
	s, err := m.Open("")
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("Open() error = %v", err)
	}
	defer s.Close()

	if s.ID == "" {
		t.Error("session ID is empty")
	}
	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}

	// Verify it's tracked
	if m.Count() != 1 {
		t.Errorf("Count() = %d, want 1", m.Count())
	}
}

func TestManagerGet(t *testing.T) {
	m := NewManager(slog.Default())
	s, err := m.Open("")
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("Open() error = %v", err)
	}
	defer s.Close()

	got, ok := m.Get(s.ID)
	if !ok {
		t.Fatal("Get() returned false for existing session")
	}
	if got.ID != s.ID {
		t.Errorf("Get() ID = %q, want %q", got.ID, s.ID)
	}

	_, ok = m.Get("nonexistent")
	if ok {
		t.Error("Get() returned true for nonexistent session")
	}
}

func TestManagerClose(t *testing.T) {
	m := NewManager(slog.Default())
	s, err := m.Open("")
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("Open() error = %v", err)
	}

	if err := m.Close(s.ID); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if m.Count() != 0 {
		t.Errorf("Count() = %d after Close, want 0", m.Count())
	}

	if err := m.Close("nonexistent"); err != ErrNoSession {
		t.Errorf("Close(nonexistent) error = %v, want ErrNoSession", err)
	}
}

func TestManagerList(t *testing.T) {
	m := NewManager(slog.Default())

	if len(m.List()) != 0 {
		t.Errorf("List() = %v, want empty", m.List())
	}

	s1, err := m.Open("")
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("Open() error = %v", err)
	}
	defer s1.Close()

	s2, err := m.Open("")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer s2.Close()

	ids := m.List()
	if len(ids) != 2 {
		t.Errorf("List() returned %d sessions, want 2", len(ids))
	}
}

func TestManagerCloseAll(t *testing.T) {
	m := NewManager(slog.Default())
	m.Open("")
	m.Open("")
	m.Open("")

	m.CloseAll()

	if m.Count() != 0 {
		t.Errorf("Count() after CloseAll = %d, want 0", m.Count())
	}
}

func TestManagerWriteRead(t *testing.T) {
	m := NewManager(slog.Default())
	s, err := m.Open("")
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("Open() error = %v", err)
	}
	defer s.Close()

	// Send a simple command
	n, err := s.Write([]byte("echo hello\n"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n == 0 {
		t.Error("Write() returned 0 bytes")
	}
}

func TestSessionIDFormat(t *testing.T) {
	m := NewManager(slog.Default())
	s, err := m.Open("")
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("Open() error = %v", err)
	}
	defer s.Close()

	if len(s.ID) < 5 || s.ID[:5] != "term_" {
		t.Errorf("session ID = %q, want 'term_' prefix", s.ID)
	}
}

// --- BPP Handler Tests (work on all platforms) ---

func TestHandleOpenInvalidJSON(t *testing.T) {
	m := NewManager(slog.Default())
	_, err := m.HandleOpen(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("HandleOpen with invalid JSON should return error")
	}
}

func TestHandleOpenValid(t *testing.T) {
	m := NewManager(slog.Default())
	resp, err := m.HandleOpen(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		if errors.Is(err, ErrUnsupported) || errors.Is(err, &json.UnmarshalTypeError{}) {
			t.Skip("terminal not supported or type error")
		}
		// On Windows, Open fails with ErrUnsupported
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("HandleOpen() error = %v", err)
	}
	result, ok := resp.(map[string]interface{})
	if !ok {
		t.Fatalf("HandleOpen() returned %T, want map[string]interface{}", resp)
	}
	if result["session_id"] == nil {
		t.Error("session_id should not be nil")
	}
	if result["created_at"] == nil {
		t.Error("created_at should not be nil")
	}
}

func TestHandleOpenWithShell(t *testing.T) {
	m := NewManager(slog.Default())
	resp, err := m.HandleOpen(context.Background(), json.RawMessage(`{"shell":"/bin/sh"}`))
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("HandleOpen() with shell error = %v", err)
	}
	if resp == nil {
		t.Fatal("HandleOpen() returned nil response")
	}
}

func TestHandleInputInvalidJSON(t *testing.T) {
	m := NewManager(slog.Default())
	_, err := m.HandleInput(context.Background(), json.RawMessage(`{bad`))
	if err == nil {
		t.Fatal("HandleInput with invalid JSON should return error")
	}
}

func TestHandleInputNoSession(t *testing.T) {
	m := NewManager(slog.Default())
	_, err := m.HandleInput(context.Background(), json.RawMessage(`{"session_id":"nonexistent","data":"hello"}`))
	if err == nil {
		t.Fatal("HandleInput with nonexistent session should return error")
	}
}

func TestHandleResizeInvalidJSON(t *testing.T) {
	m := NewManager(slog.Default())
	_, err := m.HandleResize(context.Background(), json.RawMessage(`{bad`))
	if err == nil {
		t.Fatal("HandleResize with invalid JSON should return error")
	}
}

func TestHandleResizeNoSession(t *testing.T) {
	m := NewManager(slog.Default())
	_, err := m.HandleResize(context.Background(), json.RawMessage(`{"session_id":"nonexistent","rows":24,"cols":80}`))
	if err == nil {
		t.Fatal("HandleResize with nonexistent session should return error")
	}
}

func TestHandleCloseInvalidJSON(t *testing.T) {
	m := NewManager(slog.Default())
	_, err := m.HandleClose(context.Background(), json.RawMessage(`{bad`))
	if err == nil {
		t.Fatal("HandleClose with invalid JSON should return error")
	}
}

func TestHandleCloseNoSession(t *testing.T) {
	m := NewManager(slog.Default())
	_, err := m.HandleClose(context.Background(), json.RawMessage(`{"session_id":"nonexistent"}`))
	if err == nil {
		t.Fatal("HandleClose with nonexistent session should return error")
	}
}

func TestErrorsConstants(t *testing.T) {
	if ErrNoSession == nil {
		t.Error("ErrNoSession should not be nil")
	}
	if ErrUnsupported == nil {
		t.Error("ErrUnsupported should not be nil")
	}
}

func TestManagerCloseAllEmpty(t *testing.T) {
	m := NewManager(nil)
	m.CloseAll()
	if m.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after CloseAll on empty", m.Count())
	}
}

func TestManagerListEmpty(t *testing.T) {
	m := NewManager(nil)
	ids := m.List()
	if len(ids) != 0 {
		t.Errorf("List() = %v, want empty", ids)
	}
}

func TestManagerCountAfterOperations(t *testing.T) {
	m := NewManager(nil)
	s1, err := m.Open("")
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("Open() error = %v", err)
	}

	if m.Count() != 1 {
		t.Errorf("Count() after 1 open = %d, want 1", m.Count())
	}

	s2, _ := m.Open("")
	if m.Count() != 2 {
		t.Errorf("Count() after 2 opens = %d, want 2", m.Count())
	}

	m.Close(s1.ID)
	if m.Count() != 1 {
		t.Errorf("Count() after 1 close = %d, want 1", m.Count())
	}

	m.Close(s2.ID)
	if m.Count() != 0 {
		t.Errorf("Count() after 2 closes = %d, want 0", m.Count())
	}
}

func TestManagerStopWithSessions(t *testing.T) {
	m := NewManager(slog.Default())
	m.Open("")
	m.Open("")
	m.Open("")

	if err := m.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	if m.Count() != 0 {
		t.Errorf("Count() after Stop = %d, want 0", m.Count())
	}
}

func TestManagerHealthWithSessions(t *testing.T) {
	m := NewManager(slog.Default())
	s1, err := m.Open("")
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			t.Skip("terminal not supported on this platform")
		}
		t.Fatalf("Open() error = %v", err)
	}
	s2, err := m.Open("")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	h := m.Health()
	m2, ok := h.(map[string]interface{})
	if !ok {
		t.Fatalf("Health() returned %T, want map[string]interface{}", h)
	}
	if m2["status"] != "ok" {
		t.Errorf("Health() status = %v, want \"ok\"", m2["status"])
	}
	if m2["sessions"] != 2 {
		t.Errorf("Health() sessions = %v, want 2", m2["sessions"])
	}
	s1.Close()
	s2.Close()
}

func TestGetInjectedSession(t *testing.T) {
	m := NewManager(nil)
	s := newTestSession("test_001")
	injectSession(m, s)

	got, ok := m.Get("test_001")
	if !ok {
		t.Fatal("Get() returned false for injected session")
	}
	if got.ID != "test_001" {
		t.Errorf("Get().ID = %q, want %q", got.ID, "test_001")
	}
}

func TestCloseInjectedSession(t *testing.T) {
	m := NewManager(nil)
	s := newTestSession("test_close")
	injectSession(m, s)

	if err := m.Close("test_close"); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if m.Count() != 0 {
		t.Errorf("Count() after Close = %d, want 0", m.Count())
	}
}

func TestCloseAllInjectedSessions(t *testing.T) {
	m := NewManager(nil)
	injectSession(m, newTestSession("a"))
	injectSession(m, newTestSession("b"))
	injectSession(m, newTestSession("c"))

	if m.Count() != 3 {
		t.Fatalf("Count() = %d, want 3", m.Count())
	}

	m.CloseAll()

	if m.Count() != 0 {
		t.Errorf("Count() after CloseAll = %d, want 0", m.Count())
	}
}

func TestListInjectedSessions(t *testing.T) {
	m := NewManager(nil)
	injectSession(m, newTestSession("x"))
	injectSession(m, newTestSession("y"))

	ids := m.List()
	if len(ids) != 2 {
		t.Errorf("List() returned %d sessions, want 2", len(ids))
	}
}

func TestCountInjectedSessions(t *testing.T) {
	m := NewManager(nil)
	injectSession(m, newTestSession("c1"))
	if m.Count() != 1 {
		t.Errorf("Count() = %d, want 1", m.Count())
	}
	injectSession(m, newTestSession("c2"))
	if m.Count() != 2 {
		t.Errorf("Count() = %d, want 2", m.Count())
	}
}

func TestHandleInputInjectedSession(t *testing.T) {
	m := NewManager(nil)
	injectSession(m, newTestSession("inp"))

	_, err := m.HandleInput(context.Background(), json.RawMessage(`{"session_id":"inp","data":"echo\n"}`))
	if err != nil && !errors.Is(err, ErrUnsupported) {
		t.Fatalf("HandleInput() unexpected error = %v", err)
	}
}

func TestHandleResizeInjectedSession(t *testing.T) {
	m := NewManager(nil)
	injectSession(m, newTestSession("rsz"))

	_, err := m.HandleResize(context.Background(), json.RawMessage(`{"session_id":"rsz","rows":40,"cols":120}`))
	if err != nil && !errors.Is(err, ErrUnsupported) {
		t.Fatalf("HandleResize() unexpected error = %v", err)
	}
}

func TestHandleCloseInjectedSession(t *testing.T) {
	m := NewManager(nil)
	injectSession(m, newTestSession("cls"))

	resp, err := m.HandleClose(context.Background(), json.RawMessage(`{"session_id":"cls"}`))
	if err != nil {
		t.Fatalf("HandleClose() error = %v", err)
	}
	result, ok := resp.(map[string]interface{})
	if !ok {
		t.Fatalf("HandleClose() returned %T, want map[string]interface{}", resp)
	}
	if result["closed"] != true {
		t.Errorf("closed = %v, want true", result["closed"])
	}
	if m.Count() != 0 {
		t.Errorf("Count() after HandleClose = %d, want 0", m.Count())
	}
}

func TestStopInjectedSessions(t *testing.T) {
	m := NewManager(nil)
	injectSession(m, newTestSession("s1"))
	injectSession(m, newTestSession("s2"))

	if err := m.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if m.Count() != 0 {
		t.Errorf("Count() after Stop = %d, want 0", m.Count())
	}
}

func TestHealthInjectedSessions(t *testing.T) {
	m := NewManager(nil)
	injectSession(m, newTestSession("h1"))
	injectSession(m, newTestSession("h2"))

	h := m.Health()
	m2 := h.(map[string]interface{})
	if m2["sessions"] != 2 {
		t.Errorf("Health() sessions = %v, want 2", m2["sessions"])
	}
}

func TestManagerOpenThenInjectedGetClose(t *testing.T) {
	m := NewManager(nil)
	injectSession(m, newTestSession("combo"))

	got, ok := m.Get("combo")
	if !ok || got == nil {
		t.Fatal("Get() failed for injected session")
	}

	if err := m.Close("combo"); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	_, ok = m.Get("combo")
	if ok {
		t.Error("Get() returned true after Close()")
	}
}
