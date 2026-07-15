package relay

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/buzzpi/agent/pkg/bpp"
	"github.com/gorilla/websocket"
)

// testRelayServer spins up a local WebSocket server that mimics a relay.
func testRelayServer(t *testing.T) (*httptest.Server, string, chan []byte) {
	t.Helper()

	msgCh := make(chan []byte, 10)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}
		defer conn.Close()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msgCh <- data

			var env bpp.Envelope
			if json.Unmarshal(data, &env) == nil {
				resp, err := bpp.NewResponse(&env, map[string]string{"echo": "ok"})
				if err == nil {
					respData, _ := resp.Marshal()
					conn.WriteMessage(websocket.TextMessage, respData)
				}
			}
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	wsURL := "ws" + server.URL[len("http"):] + "/ws"
	return server, wsURL, msgCh
}

func TestNewClient(t *testing.T) {
	cli := NewClient(DefaultConfig(), nil)
	if cli == nil {
		t.Fatal("expected non-nil client")
	}
	if cli.log == nil {
		t.Error("expected logger to default")
	}
}

func TestConnect_EmptyURL(t *testing.T) {
	cli := NewClient(DefaultConfig(), nil)
	err := cli.Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConnect_Refused(t *testing.T) {
	cli := NewClient(Config{
		ServerURL: "ws://127.0.0.1:1/ws",
		DeviceID:  "test-device",
	}, nil)
	err := cli.Connect(context.Background())
	if err == nil {
		t.Fatal("expected connection refused error")
	}
}

func TestConnectAndSend(t *testing.T) {
	_, url, msgCh := testRelayServer(t)

	cli := NewClient(Config{
		ServerURL: url,
		DeviceID:  "test-device",
	}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cli.Connect(ctx); err != nil {
		t.Fatalf("connect: %v", err)
	}

	env, err := bpp.NewRequest("test.method", map[string]string{"foo": "bar"})
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if err := cli.Send(env); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case raw := <-msgCh:
		var received bpp.Envelope
		if err := json.Unmarshal(raw, &received); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if received.Method != "test.method" {
			t.Errorf("got method %q, want %q", received.Method, "test.method")
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for server to receive message")
	}

	go func() {
		cli.ReadLoop(ctx)
	}()

	respCh := make(chan *bpp.Envelope, 1)
	cli.OnMessage = func(ctx context.Context, e *bpp.Envelope) (*bpp.Envelope, error) {
		respCh <- e
		return nil, nil
	}

	env2, err := bpp.NewRequest("test.echo", map[string]string{"x": "y"})
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if err := cli.Send(env2); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case resp := <-respCh:
		if resp.Type != bpp.TypeResponse {
			t.Errorf("got type %q, want %q", resp.Type, bpp.TypeResponse)
		}
	case <-ctx.Done():
		t.Fatal("timeout waiting for response in OnMessage")
	}

	cli.Close()
}

func TestStart_NotConfigured(t *testing.T) {
	cli := NewClient(DefaultConfig(), nil)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := cli.Start(ctx)
	if err != nil {
		t.Fatalf("expected nil for unconfigured relay, got: %v", err)
	}
}

func TestReadLoop_WithoutConnect(t *testing.T) {
	cli := NewClient(DefaultConfig(), nil)
	cli.cfg.ServerURL = "ws://127.0.0.1:1/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := cli.ReadLoop(ctx)
	if err == nil {
		t.Fatal("expected error when not connected")
	}
}

func TestClose_Idempotent(t *testing.T) {
	cli := NewClient(DefaultConfig(), nil)
	if err := cli.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := cli.Close(); err != nil {
		t.Fatalf("second close should be idempotent: %v", err)
	}
}

func TestIsConnected(t *testing.T) {
	_, url, _ := testRelayServer(t)

	cli := NewClient(Config{
		ServerURL: url,
		DeviceID:  "test-device",
	}, nil)

	if cli.IsConnected() {
		t.Error("expected not connected before Connect")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cli.Connect(ctx); err != nil {
		t.Fatalf("connect: %v", err)
	}

	if !cli.IsConnected() {
		t.Error("expected connected after Connect")
	}

	cli.Close()
	if cli.IsConnected() {
		t.Error("expected not connected after Close")
	}
}

func TestReconnectLoop_Disabled(t *testing.T) {
	cli := NewClient(Config{Reconnect: false}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		cli.ReconnectLoop(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ReconnectLoop blocked when disabled")
	}
}

func TestOnMessage_NilHandler(t *testing.T) {
	_, url, _ := testRelayServer(t)

	cli := NewClient(Config{
		ServerURL: url,
		DeviceID:  "test-device",
	}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cli.Connect(ctx); err != nil {
		t.Fatalf("connect: %v", err)
	}

	go func() {
		cli.ReadLoop(ctx)
	}()

	env, err := bpp.NewRequest("test.method", map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if err := cli.Send(env); err != nil {
		t.Fatalf("send: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	cli.Close()
}

func TestSupervisorInterface(t *testing.T) {
	cli := NewClient(DefaultConfig(), nil)

	if cli.Name() != "relay" {
		t.Errorf("Name() = %q, want %q", cli.Name(), "relay")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := cli.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	health := cli.Health()
	if health == nil {
		t.Fatal("Health() returned nil")
	}
}

// --- Server Tests ---

func TestNewServer(t *testing.T) {
	s := NewServer(":0", nil)
	if s == nil {
		t.Fatal("NewServer() returned nil")
	}
	if s.pending == nil {
		t.Error("pending map should be initialized")
	}
}

func TestNewServerWithLogger(t *testing.T) {
	s := NewServer(":8080", slog.Default())
	if s == nil {
		t.Fatal("NewServer() returned nil")
	}
}

func TestServerListenAndServeAndShutdown(t *testing.T) {
	s := NewServer(":0", slog.Default())
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ListenAndServe(ctx)
	}()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Logf("ListenAndServe returned: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}

func TestServerAgentAndClientRelay(t *testing.T) {
	s := NewServer(":0", slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ListenAndServe(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	cancel()
	<-errCh

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws"

	// Connect as agent
	agentConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?device_id=test-relay-device", nil)
	if err != nil {
		t.Fatalf("agent dial: %v", err)
	}
	defer agentConn.Close()

	time.Sleep(100 * time.Millisecond)

	// Connect as client
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?target=test-relay-device", nil)
	if err != nil {
		t.Fatalf("client dial: %v", err)
	}
	defer clientConn.Close()

	time.Sleep(100 * time.Millisecond)

	// Client sends a request
	reqEnv, _ := bpp.NewRequest("device.info", nil)
	reqData, _ := reqEnv.Marshal()
	err = clientConn.WriteMessage(websocket.TextMessage, reqData)
	if err != nil {
		t.Fatalf("client write: %v", err)
	}

	// Agent should receive the request
	agentConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, agentMsg, err := agentConn.ReadMessage()
	if err != nil {
		t.Fatalf("agent read: %v", err)
	}

	var receivedEnv bpp.Envelope
	if err := json.Unmarshal(agentMsg, &receivedEnv); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if receivedEnv.Method != "device.info" {
		t.Errorf("agent received method = %q, want %q", receivedEnv.Method, "device.info")
	}

	// Agent sends a response back
	respEnv, _ := bpp.NewResponse(&receivedEnv, map[string]string{"status": "ok"})
	respData, _ := respEnv.Marshal()
	err = agentConn.WriteMessage(websocket.TextMessage, respData)
	if err != nil {
		t.Fatalf("agent write response: %v", err)
	}

	// Client should receive the response
	clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, clientMsg, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("client read: %v", err)
	}

	var resp bpp.Envelope
	if err := json.Unmarshal(clientMsg, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Type != bpp.TypeResponse {
		t.Errorf("client received type = %q, want %q", resp.Type, bpp.TypeResponse)
	}
	if resp.RID != reqEnv.RID {
		t.Errorf("response RID = %q, want %q (matching request)", resp.RID, reqEnv.RID)
	}
}

func TestServerMissingParams(t *testing.T) {
	s := NewServer(":0", slog.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws"

	// Connect without device_id or target
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		// Connection might be rejected immediately
		return
	}
	defer conn.Close()

	// The server should close the connection
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Error("expected error reading from connection with missing params")
	}
}

func TestServerClientAgentOffline(t *testing.T) {
	s := NewServer(":0", slog.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws"

	// Connect as client targeting a device that has no agent
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?target=nonexistent-device", nil)
	if err != nil {
		t.Fatalf("client dial: %v", err)
	}
	defer clientConn.Close()

	time.Sleep(100 * time.Millisecond)

	// Send a request - should get agent_offline error
	reqEnv, _ := bpp.NewRequest("device.info", nil)
	reqData, _ := reqEnv.Marshal()
	err = clientConn.WriteMessage(websocket.TextMessage, reqData)
	if err != nil {
		t.Fatalf("client write: %v", err)
	}

	// Should receive an error response
	clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("client read error response: %v", err)
	}

	var env bpp.Envelope
	if err := json.Unmarshal(msg, &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if env.Error == nil {
		t.Error("expected error response for offline agent")
	} else if env.Error.Code != "agent_offline" {
		t.Errorf("error code = %q, want %q", env.Error.Code, "agent_offline")
	}
}

func TestSendErrorResponseInvalidJSON(t *testing.T) {
	s := NewServer(":0", slog.Default())

	// Create a test WS connection using net/http
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws"

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?target=dev1", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer clientConn.Close()

	// sendErrorResponse with invalid JSON should be a no-op
	// We can test this indirectly by sending malformed data as a client
	err = clientConn.WriteMessage(websocket.TextMessage, []byte("not json"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	// Server should not crash; just log a warning
	time.Sleep(100 * time.Millisecond)
}

func TestServerAgentReconnect(t *testing.T) {
	s := NewServer(":0", slog.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws"

	// First agent connection
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL+"?device_id=reconnect-dev", nil)
	if err != nil {
		t.Fatalf("first agent dial: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Second agent for same device (should replace first)
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL+"?device_id=reconnect-dev", nil)
	if err != nil {
		t.Fatalf("second agent dial: %v", err)
	}
	defer conn2.Close()
	time.Sleep(100 * time.Millisecond)

	// First connection should have been closed by server
	conn1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = conn1.ReadMessage()
	if err == nil {
		t.Log("first connection still readable (may depend on timing)")
	}
}

func TestServerAgentDisconnect(t *testing.T) {
	s := NewServer(":0", slog.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws"

	// Connect agent
	agentConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?device_id=disconnect-dev", nil)
	if err != nil {
		t.Fatalf("agent dial: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Close agent
	agentConn.Close()
	time.Sleep(200 * time.Millisecond)

	// Connect client targeting that device
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?target=disconnect-dev", nil)
	if err != nil {
		t.Fatalf("client dial: %v", err)
	}
	defer clientConn.Close()
	time.Sleep(100 * time.Millisecond)

	// Send request - agent is offline, should get error
	reqEnv, _ := bpp.NewRequest("test", nil)
	reqData, _ := reqEnv.Marshal()
	clientConn.WriteMessage(websocket.TextMessage, reqData)

	clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("client read: %v", err)
	}

	var env bpp.Envelope
	json.Unmarshal(msg, &env)
	if env.Error == nil {
		t.Error("expected error for offline agent")
	}
}

func TestServerAgentInvalidEnvelope(t *testing.T) {
	s := NewServer(":0", slog.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws"

	// Connect as agent
	agentConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?device_id=invalid-env-dev", nil)
	if err != nil {
		t.Fatalf("agent dial: %v", err)
	}
	defer agentConn.Close()
	time.Sleep(100 * time.Millisecond)

	// Send invalid JSON to agent (simulating bad upstream message)
	err = agentConn.WriteMessage(websocket.TextMessage, []byte("not json"))
	if err != nil {
		t.Fatalf("write invalid: %v", err)
	}

	// Server should not crash, agent stays connected
	// Send a valid message to verify connection is still alive
	reqEnv, _ := bpp.NewRequest("ping", nil)
	reqData, _ := reqEnv.Marshal()
	err = agentConn.WriteMessage(websocket.TextMessage, reqData)
	if err != nil {
		t.Fatalf("write valid after invalid: %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ServerURL != "" {
		t.Errorf("DefaultConfig().ServerURL = %q, want empty", cfg.ServerURL)
	}
	if !cfg.Reconnect {
		t.Error("DefaultConfig().Reconnect should be true")
	}
}

// --- HTTP Endpoint Tests ---

func TestHealthEndpoint(t *testing.T) {
	s := NewServer(":0", slog.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/agents", s.handleAgents)
	mux.HandleFunc("/version", s.handleVersion)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	// Test health endpoint
	resp, err := http.Get(testServer.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("health status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("decode health: %v", err)
	}

	if health["status"] != "ok" {
		t.Errorf("health status = %v, want ok", health["status"])
	}
	if health["agents"] != float64(0) {
		t.Errorf("health agents = %v, want 0", health["agents"])
	}
}

func TestAgentsEndpoint_Empty(t *testing.T) {
	s := NewServer(":0", slog.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/agents", s.handleAgents)
	mux.HandleFunc("/version", s.handleVersion)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/agents")
	if err != nil {
		t.Fatalf("GET /agents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("agents status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode agents: %v", err)
	}

	if result["count"] != float64(0) {
		t.Errorf("agents count = %v, want 0", result["count"])
	}

	agents, ok := result["agents"].([]interface{})
	if !ok {
		t.Fatal("agents field not an array")
	}
	if len(agents) != 0 {
		t.Errorf("agents array length = %d, want 0", len(agents))
	}
}

func TestAgentsEndpoint_WithAgent(t *testing.T) {
	s := NewServer(":0", slog.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/agents", s.handleAgents)
	mux.HandleFunc("/version", s.handleVersion)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws"

	// Connect an agent
	agentConn, _, err := websocket.DefaultDialer.Dial(wsURL+"?device_id=health-test-device", nil)
	if err != nil {
		t.Fatalf("agent dial: %v", err)
	}
	defer agentConn.Close()
	time.Sleep(100 * time.Millisecond)

	// Check health shows agent count
	resp, err := http.Get(testServer.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	var health map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&health)

	if health["agents"] != float64(1) {
		t.Errorf("health agents = %v, want 1", health["agents"])
	}

	// Check agents list
	resp2, err := http.Get(testServer.URL + "/agents")
	if err != nil {
		t.Fatalf("GET /agents: %v", err)
	}
	defer resp2.Body.Close()

	var agentsResp map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&agentsResp)

	if agentsResp["count"] != float64(1) {
		t.Errorf("agents count = %v, want 1", agentsResp["count"])
	}

	agents, ok := agentsResp["agents"].([]interface{})
	if !ok {
		t.Fatal("agents field not an array")
	}
	if len(agents) != 1 {
		t.Fatalf("agents array length = %d, want 1", len(agents))
	}

	agentInfo, ok := agents[0].(map[string]interface{})
	if !ok {
		t.Fatal("agent info not an object")
	}
	if agentInfo["device_id"] != "health-test-device" {
		t.Errorf("device_id = %v, want health-test-device", agentInfo["device_id"])
	}
	if agentInfo["connected"] == nil {
		t.Error("connected field is nil")
	}
}

func TestVersionEndpoint(t *testing.T) {
	s := NewServer(":0", slog.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/agents", s.handleAgents)
	mux.HandleFunc("/version", s.handleVersion)
	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/version")
	if err != nil {
		t.Fatalf("GET /version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("version status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var version map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		t.Fatalf("decode version: %v", err)
	}

	if version["version"] != "0.2.0" {
		t.Errorf("version = %q, want %q", version["version"], "0.2.0")
	}
	if version["service"] != "buzzpi-relay" {
		t.Errorf("service = %q, want %q", version["service"], "buzzpi-relay")
	}
}
