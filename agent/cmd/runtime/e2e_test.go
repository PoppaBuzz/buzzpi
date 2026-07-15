package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/buzzpi/agent/internal/config"
	"github.com/buzzpi/agent/internal/device"
	"github.com/buzzpi/agent/internal/engine"
	"github.com/buzzpi/agent/internal/identity"
	"github.com/buzzpi/agent/internal/pairing"
	"github.com/buzzpi/agent/internal/state"
	"github.com/buzzpi/agent/internal/terminal"
	"github.com/buzzpi/agent/pkg/bpp"
	"github.com/gorilla/websocket"
)

func resolvePIN(t *testing.T, chalEnv *bpp.Envelope) string {
	t.Helper()
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
	return challenge.PIN
}

// readResponse reads a BPP message from the WebSocket, unmarshals it, and
// fails the test on transport or protocol errors.
func readResponse(t *testing.T, conn *websocket.Conn) *bpp.Envelope {
	t.Helper()
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	env, err := bpp.Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("protocol error: %s: %s", env.Error.Code, env.Error.Message)
	}
	return env
}

// sendRequest marshals and writes a BPP request to the WebSocket.
func sendRequest(t *testing.T, conn *websocket.Conn, method string, params interface{}) *bpp.Envelope {
	t.Helper()
	req, err := bpp.NewRequest(method, params)
	if err != nil {
		t.Fatalf("create %s request: %v", method, err)
	}
	data, _ := req.Marshal()
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("write %s: %v", method, err)
	}
	return req
}

func TestE2E_FullStack(t *testing.T) {
	// ── Setup ────────────────────────────────────────────────────────────────

	tmpDir, err := os.MkdirTemp("", "buzzpi-e2e-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "e2e.db")
	idPath := identity.IdentityPath(tmpDir)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn, // quiet during tests
	}))

	// State store
	store, err := state.Open(dbPath, logger)
	if err != nil {
		t.Fatalf("open state: %v", err)
	}
	defer store.Close()

	// Identity
	devID, err := identity.Ensure(idPath)
	if err != nil {
		t.Fatalf("ensure identity: %v", err)
	}

	// Engine
	eng := engine.NewManager(logger)

	// Device service
	cfg, _ := config.Load("")
	cfg.Runtime.DeviceName = "e2e-test-device"
	devSvc, err := device.NewService(cfg, devID.DeviceID, logger)
	if err != nil {
		t.Fatalf("new device service: %v", err)
	}
	eng.RegisterMethod("device.info", devSvc.HandleInfo)
	eng.RegisterMethod("device.stats", devSvc.HandleStats)

	pairH := pairing.NewHandler(store, devID.DeviceID, logger)
	eng.RegisterMethod("pair.initiate", pairH.HandleInitiate)
	eng.RegisterMethod("pair.verify", pairH.HandleVerify)
	eng.RegisterMethod("pair.status", pairH.HandleStatus)
	eng.RegisterMethod("pair.unpair", pairH.HandleUnpair)

	termMgr := terminal.NewManager(logger)
	eng.RegisterMethod("terminal.open", termMgr.HandleOpen)
	eng.RegisterMethod("terminal.input", termMgr.HandleInput)
	eng.RegisterMethod("terminal.resize", termMgr.HandleResize)
	eng.RegisterMethod("terminal.close", termMgr.HandleClose)

	// WebSocket test server
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		var (
			hs      *bpp.Handshake
			ready   bool
			session string
		)

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			env, err := bpp.Unmarshal(data)
			if err != nil {
				continue
			}

			var respData []byte

			switch env.Method {
			case bpp.MethodHandshake:
				hs = bpp.NewHandshake(bpp.HandshakeConfig{
					Role:          bpp.RoleAgent,
					DeviceID:      devID.DeviceID,
					DeviceName:    "e2e-test-device",
					Platform:      runtime.GOOS + "/" + runtime.GOARCH,
					Version:       "0.0.0-e2e-test",
					Capabilities:  bpp.DefaultCapabilities,
					TokenDuration: 24 * time.Hour,
				})
				var offer bpp.CapabilityOffer
				if err := json.Unmarshal(env.Params, &offer); err != nil {
					resp := bpp.NewErrorResponse(env, "invalid_params", "bad offer")
					respData, _ = resp.Marshal()
					break
				}
				r, err := hs.HandleOffer(&offer)
				if err != nil {
					resp := bpp.NewErrorResponse(env, "handshake_error", err.Error())
					respData, _ = resp.Marshal()
					break
				}
				switch v := r.(type) {
				case *bpp.CapabilityAccept:
					ready = true
					session = v.SessionToken
					resp, _ := bpp.NewResponse(env, v)
					respData, _ = resp.Marshal()
				case *bpp.AuthChallenge:
					resp, _ := bpp.NewResponse(env, v)
					respData, _ = resp.Marshal()
				default:
					resp := bpp.NewErrorResponse(env, "unexpected", "unexpected handshake response")
					respData, _ = resp.Marshal()
				}

			case bpp.MethodAuthResponse:
				if hs == nil {
					resp := bpp.NewErrorResponse(env, "no_handshake", "no handshake in progress")
					respData, _ = resp.Marshal()
					break
				}
				var authResp bpp.AuthResponse
				if err := json.Unmarshal(env.Params, &authResp); err != nil {
					resp := bpp.NewErrorResponse(env, "invalid_params", "bad auth response")
					respData, _ = resp.Marshal()
					break
				}
				est, err := hs.HandleAuthResponse(&authResp)
				if err != nil {
					resp := bpp.NewErrorResponse(env, "auth_failed", err.Error())
					respData, _ = resp.Marshal()
					break
				}
				ready = true
				session = est.SessionToken
				resp, _ := bpp.NewResponse(env, est)
				respData, _ = resp.Marshal()

			default:
				if !ready {
					resp := bpp.NewErrorResponse(env, "handshake_required", "complete handshake first")
					respData, _ = resp.Marshal()
				} else {
					mctx := context.WithValue(context.Background(), engine.CtxSessionToken, session)
					resp, err := eng.Handle(mctx, env)
					if err != nil {
						continue
					}
					respData, _ = resp.Marshal()
				}
			}

			if respData != nil {
				conn.WriteMessage(websocket.TextMessage, respData)
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"):] + "/ws"

	// ── Connect ──────────────────────────────────────────────────────────────

	dialer := websocket.DefaultDialer
	wsConn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer wsConn.Close()

	// ── Step 1: Handshake ────────────────────────────────────────────────────

	t.Log("=== Step 1: Handshake ===")

	sendRequest(t, wsConn, bpp.MethodHandshake, bpp.CapabilityOffer{
		Version:      1,
		Capabilities: []string{"bpp.core"},
	})

	chalEnv := readResponse(t, wsConn)
	var challenge bpp.AuthChallenge
	if err := json.Unmarshal(chalEnv.Result, &challenge); err != nil {
		t.Fatalf("unmarshal challenge: %v", err)
	}
	if challenge.ChallengeType != "pin" {
		t.Fatalf("challenge type = %q, want %q", challenge.ChallengeType, "pin")
	}
	if challenge.PIN == "" {
		t.Fatal("empty PIN in challenge")
	}
	if challenge.DeviceID == "" {
		t.Fatal("empty DeviceID in challenge")
	}
	t.Logf("handshake → pin challenge: device=%s pin=%s", challenge.DeviceID, challenge.PIN)

	// ── Step 2: Auth Response ────────────────────────────────────────────────

	t.Log("=== Step 2: Auth Response ===")

	sendRequest(t, wsConn, bpp.MethodAuthResponse, bpp.AuthResponse{
		PIN:        challenge.PIN,
		DeviceID:   challenge.DeviceID,
		ClientName: "e2e-test-client",
	})

	sessEnv := readResponse(t, wsConn)
	var established bpp.SessionEstablished
	if err := json.Unmarshal(sessEnv.Result, &established); err != nil {
		t.Fatalf("unmarshal session established: %v", err)
	}
	if established.SessionToken == "" {
		t.Fatal("empty session token")
	}
	t.Logf("auth complete: session=%s device=%s", established.SessionToken, established.DeviceID)

	// ── Step 3: Pair.initiate ────────────────────────────────────────────────

	t.Log("=== Step 3: Pair.initiate ===")

	sendRequest(t, wsConn, "pair.initiate", pairing.PairInitiateParams{
		DeviceID:       challenge.DeviceID,
		ClientName:     "e2e-test-client",
		SupportedAuths: []string{"pin"},
	})

	pairInitEnv := readResponse(t, wsConn)
	var pairInitResult pairing.PairInitiateResult
	if err := json.Unmarshal(pairInitEnv.Result, &pairInitResult); err != nil {
		t.Fatalf("unmarshal pair.initiate result: %v", err)
	}
	if pairInitResult.PIN == "" {
		t.Fatal("empty PIN in pair.initiate")
	}
	if pairInitResult.SessionID == "" {
		t.Fatal("empty SessionID in pair.initiate")
	}
	t.Logf("pair.initiate: session=%s pin=%s", pairInitResult.SessionID, pairInitResult.PIN)

	// ── Step 4: Pair.verify ──────────────────────────────────────────────────

	t.Log("=== Step 4: Pair.verify ===")

	sendRequest(t, wsConn, "pair.verify", pairing.PairVerifyParams{
		SessionID:  pairInitResult.SessionID,
		PIN:        pairInitResult.PIN,
		ClientName: "e2e-test-client",
	})

	pairVerifyEnv := readResponse(t, wsConn)
	var pairVerifyResult pairing.PairVerifyResult
	if err := json.Unmarshal(pairVerifyEnv.Result, &pairVerifyResult); err != nil {
		t.Fatalf("unmarshal pair.verify result: %v", err)
	}
	if pairVerifyResult.SessionToken == "" {
		t.Fatal("empty session token after verify")
	}
	t.Logf("pair.verify: paired, token=%s", pairVerifyResult.SessionToken[:16]+"...")

	// ── Step 5: Device.info (post-pairing method call) ───────────────────────

	t.Log("=== Step 5: Device.info ===")

	sendRequest(t, wsConn, "device.info", json.RawMessage(`{}`))
	devInfoEnv := readResponse(t, wsConn)
	var devInfoResult map[string]interface{}
	if err := json.Unmarshal(devInfoEnv.Result, &devInfoResult); err != nil {
		t.Fatalf("unmarshal device.info: %v", err)
	}
	if devInfoResult["device_id"] == nil {
		t.Fatal("device.info missing device_id")
	}
	t.Logf("device.info: id=%s name=%s", devInfoResult["device_id"], devInfoResult["name"])

	// ── Step 6: Terminal session ─────────────────────────────────────────────

	t.Log("=== Step 6: Terminal session ===")

	// Check if terminal is supported on this platform
	termSupported := true
	_, err = termMgr.Open("")
	if err != nil {
		if strings.Contains(err.Error(), "not supported") {
			termSupported = false
		}
	} else {
		termMgr.CloseAll()
	}

	if !termSupported {
		t.Logf("terminal not supported on %s, skipping terminal test", runtime.GOOS)
	} else {
		sendRequest(t, wsConn, "terminal.open", json.RawMessage(`{}`))
		termOpenEnv := readResponse(t, wsConn)

		var termOpenResult struct {
			SessionID string `json:"session_id"`
		}
		if err := json.Unmarshal(termOpenEnv.Result, &termOpenResult); err != nil {
			t.Fatalf("unmarshal terminal.open: %v", err)
		}
		if termOpenResult.SessionID == "" {
			t.Fatal("empty session_id in terminal.open")
		}
		t.Logf("terminal.open: session=%s", termOpenResult.SessionID)

		// Send a command — use "echo hello" which works on both Unix and Windows
		sendRequest(t, wsConn, "terminal.input", map[string]interface{}{
			"session_id": termOpenResult.SessionID,
			"data":       []byte("echo hello\n"),
		})
		termInputEnv := readResponse(t, wsConn)
		var termInputResult struct {
			Output []byte `json:"output"`
		}
		if err := json.Unmarshal(termInputEnv.Result, &termInputResult); err != nil {
			t.Fatalf("unmarshal terminal.input: %v", err)
		}
		t.Logf("terminal.input: %d bytes output", len(termInputResult.Output))
		if len(termInputResult.Output) > 0 {
			t.Logf("output (first 200 bytes): %s", string(termInputResult.Output[:min(len(termInputResult.Output), 200)]))
		}

		// Close terminal
		sendRequest(t, wsConn, "terminal.close", map[string]string{
			"session_id": termOpenResult.SessionID,
		})
		termCloseEnv := readResponse(t, wsConn)
		var termCloseResult struct {
			Closed bool `json:"closed"`
		}
		if err := json.Unmarshal(termCloseEnv.Result, &termCloseResult); err != nil {
			t.Fatalf("unmarshal terminal.close: %v", err)
		}
		if !termCloseResult.Closed {
			t.Fatal("terminal.close returned closed=false")
		}
		t.Log("terminal.close: OK")
	}

	// ── Step 7: Pair.status ──────────────────────────────────────────────────

	t.Log("=== Step 7: Pair.status ===")

	sendRequest(t, wsConn, "pair.status", map[string]string{
		"device_id": challenge.DeviceID,
	})
	statusEnv := readResponse(t, wsConn)
	var statusResult pairing.PairStatusResult
	if err := json.Unmarshal(statusEnv.Result, &statusResult); err != nil {
		t.Fatalf("unmarshal pair.status: %v", err)
	}
	if !statusResult.Paired {
		t.Fatal("pair.status: expected paired=true")
	}
	t.Logf("pair.status: paired=%v client=%s", statusResult.Paired, statusResult.ClientName)

	// ── Step 8: Pair.unpair ──────────────────────────────────────────────────

	t.Log("=== Step 8: Pair.unpair ===")

	sendRequest(t, wsConn, "pair.unpair", map[string]string{
		"session_token": established.SessionToken,
		"device_id":     challenge.DeviceID,
	})
	unpairEnv := readResponse(t, wsConn)
	var unpairResult map[string]interface{}
	if err := json.Unmarshal(unpairEnv.Result, &unpairResult); err != nil {
		t.Fatalf("unmarshal pair.unpair: %v", err)
	}
	if unpairResult["unpaired"] != true {
		t.Fatal("pair.unpair returned unpaired=false")
	}
	t.Log("pair.unpair: OK")

	t.Log("=== E2E test passed ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
