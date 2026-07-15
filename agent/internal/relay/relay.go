package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/buzzpi/agent/pkg/bpp"
)

// Config configures the Cloud Relay client.
type Config struct {
	ServerURL string `json:"server_url"` // e.g. wss://relay.buzzpi.io/ws
	DeviceID  string `json:"device_id"`
	Token     string `json:"token"` // authentication token for the relay
	Reconnect bool   `json:"reconnect"`
}

// DefaultConfig returns a default relay configuration.
func DefaultConfig() Config {
	return Config{
		ServerURL: "",
		Reconnect: true,
	}
}

// Client maintains a persistent connection to the Cloud Relay server.
type Client struct {
	cfg    Config
	log    *slog.Logger
	conn   *websocket.Conn
	mu     sync.Mutex
	done   chan struct{}
	closed bool

	// OnMessage is called when a BPP message arrives from the relay.
	// The handler should process the message and optionally return a response.
	OnMessage func(ctx context.Context, env *bpp.Envelope) (*bpp.Envelope, error)
}

// NewClient creates a new Cloud Relay client.
func NewClient(cfg Config, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	return &Client{
		cfg:  cfg,
		log:  log.With("component", "relay", "device", cfg.DeviceID),
		done: make(chan struct{}),
	}
}

// Connect establishes the WebSocket connection to the relay server.
func (c *Client) Connect(ctx context.Context) error {
	if c.cfg.ServerURL == "" {
		return fmt.Errorf("relay server URL not configured")
	}

	u, err := url.Parse(c.cfg.ServerURL)
	if err != nil {
		return fmt.Errorf("parse relay URL: %w", err)
	}

	// Add device identification as query parameters
	q := u.Query()
	q.Set("device_id", c.cfg.DeviceID)
	if c.cfg.Token != "" {
		q.Set("token", c.cfg.Token)
	}
	u.RawQuery = q.Encode()

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial relay: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.closed = false
	c.mu.Unlock()

	c.log.Info("connected to relay", "url", c.cfg.ServerURL)
	return nil
}

// ReadLoop processes messages from the relay connection.
// Messages are dispatched to OnMessage. Blocks until disconnected.
func (c *Client) ReadLoop(ctx context.Context) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			return nil
		default:
		}

		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()

		if conn == nil {
			return fmt.Errorf("connection lost")
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			c.log.Warn("relay read error", "error", err)
			return fmt.Errorf("relay read: %w", err)
		}

		var env bpp.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			c.log.Warn("invalid message from relay", "error", err)
			continue
		}

		if c.OnMessage != nil {
			resp, err := c.OnMessage(ctx, &env)
			if err != nil {
				c.log.Error("message handler error", "error", err)
				errResp := bpp.NewErrorResponse(&env, "handler_error", err.Error())
				respData, _ := errResp.Marshal()
				c.sendRaw(respData)
				continue
			}
			if resp != nil {
				respData, err := resp.Marshal()
				if err != nil {
					c.log.Error("marshal response", "error", err)
					continue
				}
				c.sendRaw(respData)
			}
		}
	}
}

// Send sends a BPP message through the relay.
func (c *Client) Send(env *bpp.Envelope) error {
	data, err := env.Marshal()
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.sendRaw(data)
}

// sendRaw writes raw data to the WebSocket connection.
func (c *Client) sendRaw(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Close disconnects from the relay server.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	close(c.done)
	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		c.conn = nil
	}
	c.log.Info("disconnected from relay")
	return nil
}

// IsConnected returns whether the client is connected to the relay.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil && !c.closed
}

// ReconnectLoop attempts to reconnect to the relay on disconnect.
func (c *Client) ReconnectLoop(ctx context.Context) {
	if !c.cfg.Reconnect {
		return
	}

	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		default:
		}

		// Wait before reconnecting
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		case <-c.done:
			return
		}

		c.log.Info("reconnecting to relay", "backoff", backoff)
		if err := c.Connect(ctx); err != nil {
			c.log.Warn("reconnect failed", "error", err, "next_backoff", backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		backoff = 1 * time.Second
		c.log.Info("reconnected to relay")

		if err := c.ReadLoop(ctx); err != nil {
			c.log.Warn("relay read loop ended", "error", err)
		}
	}
}

// Start establishes the relay connection and begins the read loop.
// Satisfies the Supervisor Component interface.
func (c *Client) Start(ctx context.Context) error {
	if c.cfg.ServerURL == "" {
		c.log.Info("relay not configured, skipping")
		return nil
	}
	if err := c.Connect(ctx); err != nil {
		return fmt.Errorf("relay connect: %w", err)
	}
	go func() {
		if err := c.ReadLoop(ctx); err != nil {
			c.log.Warn("relay read loop exited", "error", err)
		}
	}()
	go c.ReconnectLoop(ctx)
	return nil
}

// Stop disconnects from the relay. Satisfies the Supervisor Component interface.
func (c *Client) Stop(ctx context.Context) error {
	return c.Close()
}

// Name returns the component name for the Supervisor.
func (c *Client) Name() string { return "relay" }

// Health returns the relay client's health status.
func (c *Client) Health() interface{} {
	return map[string]interface{}{
		"status":     "ok",
		"connected":  c.IsConnected(),
		"server_url": c.cfg.ServerURL,
	}
}
