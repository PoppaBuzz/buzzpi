// Package ws provides WebSocket server and client implementations
// for the BuzzPi Runtime. The WebSocket server accepts BPP connections
// from clients, and the client connects to the Cloud Relay.
package ws

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// Allow all origins during development; tighten for production
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler processes incoming BPP messages from WebSocket connections.
type Handler interface {
	// HandleMessage processes a BPP message and optionally returns a response.
	HandleMessage(ctx context.Context, data []byte) ([]byte, error)

	// OnConnect is called when a new client connects.
	OnConnect(ctx context.Context, conn *Connection)

	// OnDisconnect is called when a client disconnects.
	OnDisconnect(ctx context.Context, conn *Connection)
}

// Connection represents an active WebSocket connection.
type Connection struct {
	ID          string
	RemoteAddr  string
	Conn        *websocket.Conn
	ConnectedAt time.Time
	mu          sync.Mutex
}

// Server accepts BPP WebSocket connections from clients.
type Server struct {
	httpServer *http.Server
	handler    Handler
	log        *slog.Logger
	conns      map[string]*Connection
	mu         sync.RWMutex
	port       int
}

// NewServer creates a new WebSocket server.
func NewServer(handler Handler, port int, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{
		handler: handler,
		log:     log.With("component", "ws-server"),
		conns:   make(map[string]*Connection),
		port:    port,
	}
}

// Start begins listening for WebSocket connections.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	s.httpServer = &http.Server{
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.log.Info("WebSocket server listening", "addr", listener.Addr())

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.log.Error("http server error", "error", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the WebSocket server.
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("shutting down WebSocket server")

	// Close all active connections
	s.mu.RLock()
	for _, conn := range s.conns {
		conn.Conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseServiceRestart, "server shutdown"))
		conn.Conn.Close()
	}
	s.mu.RUnlock()

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error("websocket upgrade failed", "error", err)
		return
	}

	c := &Connection{
		ID:          fmt.Sprintf("conn_%d", time.Now().UnixNano()),
		RemoteAddr:  r.RemoteAddr,
		Conn:        conn,
		ConnectedAt: time.Now(),
	}

	s.mu.Lock()
	s.conns[c.ID] = c
	count := len(s.conns)
	s.mu.Unlock()

	s.log.Info("client connected",
		"id", c.ID,
		"remote", c.RemoteAddr,
		"active_connections", count,
	)

	if s.handler != nil {
		s.handler.OnConnect(r.Context(), c)
	}

	defer func() {
		s.mu.Lock()
		delete(s.conns, c.ID)
		count := len(s.conns)
		s.mu.Unlock()

		if s.handler != nil {
			s.handler.OnDisconnect(r.Context(), c)
		}

		conn.Close()
		s.log.Info("client disconnected",
			"id", c.ID,
			"active_connections", count,
		)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				s.log.Warn("websocket read error", "error", err)
			}
			break
		}

		if s.handler != nil {
			resp, err := s.handler.HandleMessage(r.Context(), message)
			if err != nil {
				s.log.Error("message handler error", "error", err)
				continue
			}
			if resp != nil {
				c.mu.Lock()
				if err := conn.WriteMessage(websocket.TextMessage, resp); err != nil {
					c.mu.Unlock()
					s.log.Error("write error", "error", err)
					break
				}
				c.mu.Unlock()
			}
		}
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	count := len(s.conns)
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","connections":%d}`, count)
}

// Name returns the component name for the Supervisor.
func (s *Server) Name() string { return "ws-server" }

// Health returns the server health status.
func (s *Server) Health() interface{} {
	s.mu.RLock()
	count := len(s.conns)
	s.mu.RUnlock()
	return map[string]interface{}{
		"status":      "ok",
		"connections": count,
	}
}

// ActiveConnections returns the current number of active connections.
func (s *Server) ActiveConnections() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.conns)
}

// Client connects to a remote WebSocket server (e.g., the Cloud Relay).
type Client struct {
	conn *websocket.Conn
	url  string
	log  *slog.Logger
	mu   sync.Mutex
}

// NewClient creates a new WebSocket client.
func NewClient(url string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	return &Client{
		url: url,
		log: log.With("component", "ws-client"),
	}
}

// Connect establishes a WebSocket connection to the specified URL.
func (c *Client) Connect(ctx context.Context) error {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("dial %s: %w", c.url, err)
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	c.log.Info("connected to relay", "url", c.url)
	return nil
}

// Send sends a message over the WebSocket connection.
func (c *Client) Send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Close shuts down the WebSocket client.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}
