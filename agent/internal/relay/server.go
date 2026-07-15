package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/buzzpi/agent/pkg/bpp"
	"github.com/gorilla/websocket"
)

// Server is a lightweight WebSocket relay that bridges agent and client
// connections. Agents (runtimes) connect with ?device_id=X and are registered
// by device ID. Clients connect with ?target=X and their BPP messages are
// forwarded to the agent for that device. Responses are routed back via RID.
type Server struct {
	log      *slog.Logger
	upgrader websocket.Upgrader
	addr     string
	agents   sync.Map // deviceID -> *agentConn

	pendingMu sync.RWMutex
	pending   map[string]*clientConn // RID -> client awaiting response
}

type agentConn struct {
	conn      *websocket.Conn
	deviceID  string
	connected time.Time
	mu        sync.Mutex
}

type clientConn struct {
	conn     *websocket.Conn
	deviceID string
	mu       sync.Mutex
}

// NewServer creates a relay server on the given listen address.
func NewServer(addr string, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{
		log:  log.With("component", "relay-server"),
		addr: addr,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		pending: make(map[string]*clientConn),
	}
}

// ListenAndServe starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/agents", s.handleAgents)
	mux.HandleFunc("/version", s.handleVersion)

	httpServer := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpServer.Shutdown(shutdownCtx)
	}()

	s.log.Info("relay server listening", "addr", s.addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("relay server: %w", err)
	}
	return nil
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Warn("websocket upgrade failed", "error", err)
		return
	}

	deviceID := r.URL.Query().Get("device_id")
	targetID := r.URL.Query().Get("target")

	switch {
	case deviceID != "":
		s.handleAgent(conn, deviceID)
	case targetID != "":
		s.handleClient(conn, targetID)
	default:
		s.log.Warn("connection missing device_id or target query param")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation,
				"missing device_id (agent) or target (client)"))
		conn.Close()
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	count := 0
	s.agents.Range(func(_, _ any) bool {
		count++
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"agents":    count,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	type agentInfo struct {
		DeviceID  string `json:"device_id"`
		Connected string `json:"connected"`
	}

	var agents []agentInfo
	s.agents.Range(func(key, val any) bool {
		ac := val.(*agentConn)
		agents = append(agents, agentInfo{
			DeviceID:  ac.deviceID,
			Connected: ac.connected.Format(time.RFC3339),
		})
		return true
	})

	if agents == nil {
		agents = []agentInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"agents": agents,
		"count":  len(agents),
	})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": "0.2.0",
		"service": "buzzpi-relay",
	})
}

func (s *Server) handleAgent(conn *websocket.Conn, deviceID string) {
	log := s.log.With("device_id", deviceID)
	log.Info("agent connected")

	// Replace any existing agent connection for this device.
	if prev, loaded := s.agents.Load(deviceID); loaded {
		prev.(*agentConn).conn.Close()
	}

	ac := &agentConn{conn: conn, deviceID: deviceID, connected: time.Now()}
	s.agents.Store(deviceID, ac)

	defer func() {
		s.agents.Delete(deviceID)
		conn.Close()
		log.Info("agent disconnected")
	}()

	// Read loop: messages from the agent are responses or events.
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var env bpp.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			log.Warn("invalid envelope from agent", "error", err)
			continue
		}

		// Route response back to the waiting client.
		if env.RID != "" {
			s.pendingMu.RLock()
			cc, ok := s.pending[env.RID]
			s.pendingMu.RUnlock()

			if ok && cc != nil {
				cc.mu.Lock()
				err := cc.conn.WriteMessage(websocket.TextMessage, data)
				cc.mu.Unlock()
				if err != nil {
					log.Warn("failed to forward response to client", "rid", env.RID, "error", err)
				}

				s.pendingMu.Lock()
				delete(s.pending, env.RID)
				s.pendingMu.Unlock()
				continue
			}
		}

		// Unhandled agent message — no matching pending client.
		log.Warn("no pending client for agent message", "rid", env.RID, "type", env.Type)
	}
}

func (s *Server) handleClient(conn *websocket.Conn, targetID string) {
	log := s.log.With("target", targetID)
	log.Info("client connected")

	cc := &clientConn{conn: conn, deviceID: targetID}

	defer func() {
		conn.Close()
		log.Info("client disconnected")

		// Clean up any pending entries for this client.
		s.pendingMu.Lock()
		for k, v := range s.pending {
			if v == cc {
				delete(s.pending, k)
			}
		}
		s.pendingMu.Unlock()
	}()

	// Read loop: messages from the client are requests to forward to the agent.
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}

		val, ok := s.agents.Load(targetID)
		if !ok {
			log.Warn("target agent not connected")
			s.sendErrorResponse(conn, data, "agent_offline",
				fmt.Sprintf("device %q is not connected to relay", targetID))
			continue
		}
		ac := val.(*agentConn)

		var env bpp.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			log.Warn("invalid envelope from client", "error", err)
			continue
		}

		// Register the RID so the agent's response routes back to this client.
		if env.RID != "" {
			s.pendingMu.Lock()
			s.pending[env.RID] = cc
			s.pendingMu.Unlock()
		}

		ac.mu.Lock()
		err = ac.conn.WriteMessage(websocket.TextMessage, data)
		ac.mu.Unlock()
		if err != nil {
			log.Warn("failed to forward request to agent", "rid", env.RID, "error", err)
			s.pendingMu.Lock()
			delete(s.pending, env.RID)
			s.pendingMu.Unlock()
			s.sendErrorResponse(conn, data, "agent_error",
				"failed to forward message to agent")
		}
	}
}

func (s *Server) sendErrorResponse(conn *websocket.Conn, reqData []byte, code, message string) {
	var env bpp.Envelope
	if json.Unmarshal(reqData, &env) != nil {
		return
	}
	errResp := bpp.NewErrorResponse(&env, code, message)
	respData, err := errResp.Marshal()
	if err != nil {
		return
	}
	conn.WriteMessage(websocket.TextMessage, respData)
}
