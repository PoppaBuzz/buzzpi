package bpp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// Client is a BPP client that communicates with a device over WebSocket.
type Client struct {
	deviceID string
	addr     string
	port     int
	conn     *websocket.Conn
	log      *slog.Logger
}

// NewClient creates a new BPP client for a device.
func NewClient(deviceID, addr string, port int) *Client {
	return &Client{
		deviceID: deviceID,
		addr:     addr,
		port:     port,
		log:      slog.Default().With("component", "bpp-client", "device", deviceID),
	}
}

// Connect establishes a WebSocket connection to the device.
func (c *Client) Connect(ctx context.Context) error {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", c.addr, c.port), Path: "/ws"}
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial %s: %w", u.String(), err)
	}
	c.conn = conn
	c.log.Info("connected", "url", u.String())
	return nil
}

// Handshake performs the BPP handshake over the established connection.
// If sessionToken is non-empty, it will be presented for session resumption.
// Returns the CapabilityAccept on success, or prompts re-pairing via AuthChallenge error.
func (c *Client) Handshake(ctx context.Context, sessionToken string) (*CapabilityAccept, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	offer := &CapabilityOffer{
		Version:      CurrentVersion,
		Capabilities: DefaultCapabilities,
	}
	if sessionToken != "" {
		offer.SessionToken = sessionToken
	}

	req, err := NewRequest(MethodHandshake, offer)
	if err != nil {
		return nil, fmt.Errorf("create handshake: %w", err)
	}

	data, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	_, respData, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var env Envelope
	if err := json.Unmarshal(respData, &env); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if env.Error != nil {
		return nil, fmt.Errorf("handshake error: %s: %s", env.Error.Code, env.Error.Message)
	}

	var accept CapabilityAccept
	if err := json.Unmarshal(env.Result, &accept); err != nil {
		var chal AuthChallenge
		if uerr := json.Unmarshal(env.Result, &chal); uerr == nil && chal.ChallengeType != "" {
			return nil, &HandshakeAuthError{PIN: chal.PIN, DeviceID: chal.DeviceID}
		}
		return nil, fmt.Errorf("unmarshal accept: %w", err)
	}

	if accept.SessionToken == "" {
		return nil, fmt.Errorf("handshake: empty session token in response")
	}

	c.log.Info("handshake complete", "device", accept.DeviceID, "caps", accept.Capabilities)
	return &accept, nil
}

// HandshakeAuthError is returned when the handshake requires user
// authentication (e.g., PIN entry) rather than accepting a session token.
type HandshakeAuthError struct {
	PIN      string
	DeviceID string
}

func (e *HandshakeAuthError) Error() string {
	return fmt.Sprintf("authentication required: pair with device %s using PIN %s", e.DeviceID, e.PIN)
}

// Send writes a BPP request to the connection. Used for concurrent I/O
// where Call's synchronous write+read pattern is insufficient.
func (c *Client) Send(method string, params interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	req, err := NewRequest(method, params)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	data, err := req.Marshal()
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Recv reads the next response from the connection.
func (c *Client) Recv(ctx context.Context) (*Envelope, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &env, nil
}

// Close disconnects from the device.
func (c *Client) Close() error {
	if c.conn != nil {
		err := c.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// Call sends a BPP request and waits for the response.
func (c *Client) Call(ctx context.Context, method string, params interface{}) (*Envelope, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	req, err := NewRequest(method, params)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	data, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	// Read response with timeout
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	_, respData, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	var env Envelope
	if err := json.Unmarshal(respData, &env); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if env.Error != nil {
		return &env, fmt.Errorf("bpp error: %s: %s", env.Error.Code, env.Error.Message)
	}

	return &env, nil
}
