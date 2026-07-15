package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/buzzpi/agent/internal/engine"
	"github.com/buzzpi/agent/internal/identity"
	"github.com/buzzpi/agent/pkg/bpp"
	"github.com/gorilla/websocket"
)

// e2eHandler implements the WebSocket Handler interface for integration tests.
type e2eHandler struct {
	engine   *engine.Manager
	identity *identity.DeviceIdentity
	states   map[string]*e2eState
}

type e2eState struct {
	hs      *bpp.Handshake
	ready   bool
	session string
}

func newE2EHandler(t *testing.T) *e2eHandler {
	t.Helper()

	devID := &identity.DeviceIdentity{
		DeviceID:  "e2e-test-device",
		CreatedAt: time.Now().UTC(),
	}

	eng := engine.NewManager(slog.Default())
	eng.RegisterMethod("ping", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return map[string]string{"pong": "ok"}, nil
	})
	eng.RegisterMethod("echo", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return params, nil
	})

	return &e2eHandler{
		engine:   eng,
		identity: devID,
		states:   make(map[string]*e2eState),
	}
}

func (h *e2eHandler) dispatch(ctx context.Context, connID string, data []byte) ([]byte, error) {
	env, err := bpp.Unmarshal(data)
	if err != nil {
		return nil, nil
	}

	switch env.Method {
	case bpp.MethodHandshake:
		return h.handleHandshake(connID, env)
	case bpp.MethodAuthResponse:
		return h.handleAuthResponse(connID, env)
	}

	state, ok := h.states[connID]
	if !ok || !state.ready {
		return bpp.NewErrorResponse(env, "handshake_required", "complete handshake first").Marshal()
	}

	ctx = context.WithValue(ctx, engine.CtxSessionToken, state.session)
	resp, err := h.engine.Handle(ctx, env)
	if err != nil {
		return nil, nil
	}
	return resp.Marshal()
}

func (h *e2eHandler) handleHandshake(connID string, env *bpp.Envelope) ([]byte, error) {
	var offer bpp.CapabilityOffer
	if err := json.Unmarshal(env.Params, &offer); err != nil {
		return bpp.NewErrorResponse(env, "invalid_params", "bad offer").Marshal()
	}

	hs := bpp.NewHandshake(bpp.HandshakeConfig{
		Role:          bpp.RoleAgent,
		DeviceID:      h.identity.DeviceID,
		DeviceName:    "e2e-test",
		Platform:      "linux/amd64",
		Version:       "0.0.0-test",
		Capabilities:  bpp.DefaultCapabilities,
		TokenDuration: 24 * time.Hour,
	})

	resp, err := hs.HandleOffer(&offer)
	if err != nil {
		return bpp.NewErrorResponse(env, "handshake_error", err.Error()).Marshal()
	}

	switch v := resp.(type) {
	case *bpp.CapabilityAccept:
		h.states[connID] = &e2eState{
			ready:   true,
			session: v.SessionToken,
		}
		r, e := bpp.NewResponse(env, v)
		if e != nil {
			return nil, nil
		}
		return r.Marshal()

	case *bpp.AuthChallenge:
		h.states[connID] = &e2eState{hs: hs}
		r, e := bpp.NewResponse(env, v)
		if e != nil {
			return nil, nil
		}
		return r.Marshal()

	default:
		return bpp.NewErrorResponse(env, "unexpected", "unexpected handshake response").Marshal()
	}
}

func (h *e2eHandler) handleAuthResponse(connID string, env *bpp.Envelope) ([]byte, error) {
	state, ok := h.states[connID]
	if !ok || state.hs == nil {
		return bpp.NewErrorResponse(env, "no_handshake", "no handshake in progress").Marshal()
	}

	var authResp bpp.AuthResponse
	if err := json.Unmarshal(env.Params, &authResp); err != nil {
		return bpp.NewErrorResponse(env, "invalid_params", "bad auth response").Marshal()
	}

	est, err := state.hs.HandleAuthResponse(&authResp)
	if err != nil {
		return bpp.NewErrorResponse(env, "auth_failed", err.Error()).Marshal()
	}

	state.ready = true
	state.session = est.SessionToken

	r, e := bpp.NewResponse(env, est)
	if e != nil {
		return nil, nil
	}
	return r.Marshal()
}

func TestE2E_HandshakeAndMethodCall(t *testing.T) {
	handler := newE2EHandler(t)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}
		defer conn.Close()

		connID := "e2e-conn"

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			resp, err := handler.dispatch(context.Background(), connID, data)
			if err != nil {
				return
			}
			if resp != nil {
				if err := conn.WriteMessage(websocket.TextMessage, resp); err != nil {
					return
				}
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"):] + "/ws"

	dialer := websocket.DefaultDialer
	wsConn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer wsConn.Close()

	// --- Step 1: Handshake (offer -> challenge -> response -> established) ---
	offer, err := bpp.NewRequest(bpp.MethodHandshake, bpp.CapabilityOffer{
		Version:      1,
		Capabilities: []string{"bpp.core"},
	})
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}

	offerData, _ := offer.Marshal()
	if err := wsConn.WriteMessage(websocket.TextMessage, offerData); err != nil {
		t.Fatalf("write offer: %v", err)
	}

	_, challengeData, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read challenge: %v", err)
	}

	chalEnv, err := bpp.Unmarshal(challengeData)
	if err != nil {
		t.Fatalf("unmarshal challenge: %v", err)
	}
	if chalEnv.Error != nil {
		t.Fatalf("handshake error: %s: %s", chalEnv.Error.Code, chalEnv.Error.Message)
	}

	var challenge bpp.AuthChallenge
	if err := json.Unmarshal(chalEnv.Result, &challenge); err != nil {
		t.Fatalf("unmarshal auth challenge: %v", err)
	}
	if challenge.ChallengeType != "pin" {
		t.Fatalf("challenge type = %q, want %q", challenge.ChallengeType, "pin")
	}
	if challenge.PIN == "" {
		t.Fatal("PIN is empty in challenge")
	}
	if challenge.DeviceID == "" {
		t.Fatal("device ID is empty in challenge")
	}

	t.Logf("received PIN challenge: device=%s pin=%s", challenge.DeviceID, challenge.PIN)

	// Respond with the PIN
	authResp, err := bpp.NewRequest(bpp.MethodAuthResponse, bpp.AuthResponse{
		PIN:        challenge.PIN,
		DeviceID:   challenge.DeviceID,
		ClientName: "e2e-test-client",
	})
	if err != nil {
		t.Fatalf("create auth response: %v", err)
	}

	authData, _ := authResp.Marshal()
	if err := wsConn.WriteMessage(websocket.TextMessage, authData); err != nil {
		t.Fatalf("write auth response: %v", err)
	}

	_, sessData, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read session response: %v", err)
	}

	sessEnv, err := bpp.Unmarshal(sessData)
	if err != nil {
		t.Fatalf("unmarshal session response: %v", err)
	}
	if sessEnv.Error != nil {
		t.Fatalf("session error: %s: %s", sessEnv.Error.Code, sessEnv.Error.Message)
	}

	var established bpp.SessionEstablished
	if err := json.Unmarshal(sessEnv.Result, &established); err != nil {
		t.Fatalf("unmarshal session established: %v", err)
	}
	if established.SessionToken == "" {
		t.Fatal("session token is empty")
	}

	t.Logf("handshake complete: session=%s device=%s", established.SessionToken, established.DeviceID)

	// --- Step 2: Call ping method ---
	pingReq, err := bpp.NewRequest("ping", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("create ping: %v", err)
	}
	pingReq.RID = sessEnv.RID // set for routing

	pingData, _ := pingReq.Marshal()
	if err := wsConn.WriteMessage(websocket.TextMessage, pingData); err != nil {
		t.Fatalf("write ping: %v", err)
	}

	_, pingResp, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read ping response: %v", err)
	}

	pingEnv, err := bpp.Unmarshal(pingResp)
	if err != nil {
		t.Fatalf("unmarshal ping response: %v", err)
	}
	if pingEnv.Error != nil {
		t.Fatalf("ping error: %s: %s", pingEnv.Error.Code, pingEnv.Error.Message)
	}

	var pingResult map[string]string
	if err := json.Unmarshal(pingEnv.Result, &pingResult); err != nil {
		t.Fatalf("unmarshal ping result: %v", err)
	}
	if pingResult["pong"] != "ok" {
		t.Errorf("ping result = %v, want {\"pong\":\"ok\"}", pingResult)
	}

	t.Log("ping method call succeeded")

	// --- Step 3: Call echo method ---
	echoReq, err := bpp.NewRequest("echo", map[string]interface{}{
		"hello": "world",
		"n":     42,
	})
	if err != nil {
		t.Fatalf("create echo: %v", err)
	}

	echoData, _ := echoReq.Marshal()
	if err := wsConn.WriteMessage(websocket.TextMessage, echoData); err != nil {
		t.Fatalf("write echo: %v", err)
	}

	_, echoResp, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read echo response: %v", err)
	}

	echoEnv, err := bpp.Unmarshal(echoResp)
	if err != nil {
		t.Fatalf("unmarshal echo response: %v", err)
	}
	if echoEnv.Error != nil {
		t.Fatalf("echo error: %s: %s", echoEnv.Error.Code, echoEnv.Error.Message)
	}

	var echoResult map[string]interface{}
	if err := json.Unmarshal(echoEnv.Result, &echoResult); err != nil {
		t.Fatalf("unmarshal echo result: %v", err)
	}
	if echoResult["hello"] != "world" {
		t.Errorf("echo result = %v, want hello=world", echoResult)
	}
	if echoResult["n"] != float64(42) {
		t.Errorf("echo result n = %v, want 42", echoResult["n"])
	}

	t.Log("echo method call succeeded")

	// --- Step 4: Unknown method returns error ---
	badReq, err := bpp.NewRequest("nonexistent.method", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("create bad request: %v", err)
	}

	badData, _ := badReq.Marshal()
	if err := wsConn.WriteMessage(websocket.TextMessage, badData); err != nil {
		t.Fatalf("write bad request: %v", err)
	}

	_, badResp, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read bad response: %v", err)
	}

	badEnv, err := bpp.Unmarshal(badResp)
	if err != nil {
		t.Fatalf("unmarshal bad response: %v", err)
	}
	if badEnv.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if badEnv.Error.Code != "method_not_found" {
		t.Errorf("error code = %q, want %q", badEnv.Error.Code, "method_not_found")
	}

	t.Log("unknown method correctly rejected")
}

func TestE2E_HandshakeRequired(t *testing.T) {
	handler := newE2EHandler(t)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			resp, err := handler.dispatch(context.Background(), "e2e-conn", data)
			if err != nil {
				return
			}
			if resp != nil {
				conn.WriteMessage(websocket.TextMessage, resp)
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"):] + "/ws"
	wsConn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer wsConn.Close()

	// Send a request without handshake
	req, _ := bpp.NewRequest("ping", json.RawMessage(`{}`))
	reqData, _ := req.Marshal()
	wsConn.WriteMessage(websocket.TextMessage, reqData)

	_, respData, _ := wsConn.ReadMessage()
	respEnv, _ := bpp.Unmarshal(respData)

	if respEnv.Error == nil {
		t.Fatal("expected error for request before handshake")
	}
	if respEnv.Error.Code != "handshake_required" {
		t.Errorf("error code = %q, want %q", respEnv.Error.Code, "handshake_required")
	}
}
