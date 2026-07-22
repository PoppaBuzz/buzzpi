package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

const DefaultInactivityTimeout = 5 * time.Minute

// Manager manages multiple PTY terminal sessions.
type Manager struct {
	logger             *slog.Logger
	sessions           sync.Map
	inactivityTimeout  time.Duration
}

type senderKeyType struct{}

var SenderKey = senderKeyType{}

// NewManager creates a new terminal session manager.
func NewManager(logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{
		logger:            logger.With("component", "terminal"),
		inactivityTimeout: DefaultInactivityTimeout,
	}
}

// Open creates a new terminal session with the given shell.
// If shell is empty, the default shell is used.
// Returns the session and the generated session ID.
func (m *Manager) Open(shell string) (*Session, error) {
	id := "term_" + uuid.New().String()[:12]
	s, err := Create(id, shell, m.inactivityTimeout)
	if err != nil {
		return nil, fmt.Errorf("create terminal: %w", err)
	}

	m.sessions.Store(id, s)
	m.logger.Info("terminal session opened", "id", id)
	return s, nil
}

// Get returns a session by ID.
func (m *Manager) Get(id string) (*Session, bool) {
	v, ok := m.sessions.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Session), true
}

// Close terminates a terminal session.
func (m *Manager) Close(id string) error {
	v, ok := m.sessions.Load(id)
	if !ok {
		return ErrNoSession
	}
	s := v.(*Session)
	m.sessions.Delete(id)
	m.logger.Info("terminal session closed", "id", id)
	return s.Close()
}

// List returns all active session IDs.
func (m *Manager) List() []string {
	var ids []string
	m.sessions.Range(func(key, value interface{}) bool {
		ids = append(ids, key.(string))
		return true
	})
	return ids
}

// CloseAll terminates all sessions.
func (m *Manager) CloseAll() {
	m.sessions.Range(func(key, value interface{}) bool {
		id := key.(string)
		s := value.(*Session)
		s.Close()
		m.sessions.Delete(id)
		return true
	})
}

// Count returns the number of active sessions.
func (m *Manager) Count() int {
	count := 0
	m.sessions.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// Name returns the component name for the Supervisor.
func (m *Manager) Name() string { return "terminal" }

// Start is a no-op for the Supervisor — terminal sessions are created on demand.
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("terminal manager ready")
	return nil
}

// Stop terminates all active sessions for the Supervisor.
func (m *Manager) Stop(ctx context.Context) error {
	m.CloseAll()
	m.logger.Info("terminal manager stopped")
	return nil
}

// Health returns the terminal manager's health status.
func (m *Manager) Health() interface{} {
	return map[string]interface{}{
		"status":   "ok",
		"sessions": m.Count(),
	}
}

// SetInactivityTimeout configures the idle timeout for new sessions.
func (m *Manager) SetInactivityTimeout(d time.Duration) {
	m.inactivityTimeout = d
}

// CloseIdle closes sessions that have exceeded the inactivity timeout.
// Returns the number of sessions closed.
func (m *Manager) CloseIdle() int {
	var closed int
	now := time.Now()
	m.sessions.Range(func(key, value interface{}) bool {
		id := key.(string)
		s := value.(*Session)
		s.mu.Lock()
		idle := now.Sub(s.lastActivity)
		timeout := s.inactivityTimeout
		s.mu.Unlock()

		if timeout > 0 && idle >= timeout {
			m.Close(id)
			closed++
			m.logger.Info("closed idle terminal session", "id", id, "idle", idle)
		}
		return true
	})
	return closed
}

// HandleOpen is the BPP handler for terminal.open.
func (m *Manager) HandleOpen(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Shell string `json:"shell,omitempty"`
		Cols  int    `json:"cols,omitempty"`
		Rows  int    `json:"rows,omitempty"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	s, err := m.Open(req.Shell)
	if err != nil {
		return nil, fmt.Errorf("open terminal: %w", err)
	}

	if req.Cols > 0 && req.Rows > 0 {
		if err := s.Resize(uint16(req.Rows), uint16(req.Cols)); err != nil {
			m.logger.Warn("failed to apply initial terminal size", "error", err)
		}
	}

	if sender, ok := ctx.Value(SenderKey).(func([]byte) error); ok && sender != nil {
		s.StartOutputLoop(sender)
	}

	return map[string]interface{}{
		"session_id": s.ID,
		"created_at": s.CreatedAt,
	}, nil
}

// HandleInput is the BPP handler for terminal.input.
func (m *Manager) HandleInput(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
		Data      string `json:"data"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	s, ok := m.Get(req.SessionID)
	if !ok {
		return nil, fmt.Errorf("session not found: %s", req.SessionID)
	}

	if _, err := s.Write([]byte(req.Data)); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	return map[string]interface{}{
		"ok": true,
	}, nil
}

// HandleResize is the BPP handler for terminal.resize.
func (m *Manager) HandleResize(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
		Rows      uint16 `json:"rows"`
		Cols      uint16 `json:"cols"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	s, ok := m.Get(req.SessionID)
	if !ok {
		return nil, fmt.Errorf("session not found: %s", req.SessionID)
	}

	if err := s.Resize(req.Rows, req.Cols); err != nil {
		return nil, fmt.Errorf("resize: %w", err)
	}

	return map[string]interface{}{
		"rows": req.Rows,
		"cols": req.Cols,
	}, nil
}

// HandleClose is the BPP handler for terminal.close.
func (m *Manager) HandleClose(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	if err := m.Close(req.SessionID); err != nil {
		return nil, fmt.Errorf("close: %w", err)
	}

	return map[string]interface{}{
		"closed": true,
	}, nil
}
