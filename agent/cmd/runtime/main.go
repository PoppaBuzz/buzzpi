// Command runtime starts the BuzzPi Runtime daemon.
//
// The Runtime is the Go daemon that runs on every BuzzPi device. It manages
// device identity, discovers peers via mDNS, serves BPP over WebSocket,
// and hosts plugins.
//
// Usage:
//
//	buzzpi-runtime                # start with default config
//	buzzpi-runtime --config /etc/buzzpi/runtime.json
//	buzzpi-runtime --device-name kitchen-pi
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/buzzpi/agent/internal/config"
	"github.com/buzzpi/agent/internal/connection"
	"github.com/buzzpi/agent/internal/device"
	"github.com/buzzpi/agent/internal/engine"
	"github.com/buzzpi/agent/internal/file"
	"github.com/buzzpi/agent/internal/identity"
	"github.com/buzzpi/agent/internal/mdns"
	"github.com/buzzpi/agent/internal/relay"
	"github.com/buzzpi/agent/internal/pairing"
	"github.com/buzzpi/agent/internal/screen"
	"github.com/buzzpi/agent/internal/state"
	"github.com/buzzpi/agent/internal/supervisor"
	"github.com/buzzpi/agent/internal/terminal"
	"github.com/buzzpi/agent/internal/version"
	"github.com/buzzpi/agent/internal/ws"
	"github.com/buzzpi/agent/pkg/bpp"
)

func main() {
	// Parse flags
	configPath := flag.String("config", "", "path to config file (optional)")
	deviceName := flag.String("device-name", "", "override device name")
	showVersion := flag.Bool("version", false, "show version and exit")
	dbPath := flag.String("db", "/var/lib/buzzpi/state.db", "path to state database")
	relayURL := flag.String("relay", "", "cloud relay server URL (e.g. wss://relay.buzzpi.io/ws)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Info())
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// CLI flag overrides config file
	if *deviceName != "" {
		cfg.Runtime.DeviceName = *deviceName
	}

	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("starting BuzzPi Runtime",
		"version", version.Version,
		"commit", version.Commit,
		"go_version", version.GoVersion(),
	)

	// Open state store
	store, err := state.Open(*dbPath, logger)
	if err != nil {
		logger.Error("failed to open state store", "error", err, "path", *dbPath)
		os.Exit(1)
	}
	defer store.Close()
	logger.Info("state store opened", "path", *dbPath)

	idPath := identity.IdentityPath(filepath.Dir(*dbPath))
	devID, err := identity.Ensure(idPath)
	if err != nil {
		logger.Error("failed to load device identity", "error", err, "path", idPath)
		os.Exit(1)
	}
	logger.Info("device identity loaded",
		"device_id", devID.DeviceID,
		"created", devID.CreatedAt,
	)

	platform := runtime.GOOS + "/" + runtime.GOARCH
	if cfg.Runtime.DeviceName == "" {
		cfg.Runtime.DeviceName = devID.DeviceID
	}

	// Initialize Engine Manager and register built-in methods
	eng := engine.NewManager(logger)

	// Initialize Device service
	devSvc, err := device.NewService(cfg, *configPath, devID.DeviceID, logger)
	if err != nil {
		logger.Error("failed to initialize device service", "error", err)
		os.Exit(1)
	}
	eng.RegisterMethod("device.info", devSvc.HandleInfo)
	eng.RegisterMethod("device.stats", devSvc.HandleStats)
	eng.RegisterMethod("device.rename", devSvc.HandleRename)

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

	screenH := screen.NewHandler(logger)
	eng.RegisterMethod("screen.capture", screenH.HandleCapture)
	eng.RegisterMethod("screen.input", screenH.HandleInput)

	fileH := file.NewHandler(logger)
	eng.RegisterMethod("file.browse", fileH.HandleBrowse)
	eng.RegisterMethod("file.upload", fileH.HandleUpload)
	eng.RegisterMethod("file.download", fileH.HandleDownload)
	eng.RegisterMethod("file.delete", fileH.HandleDelete)
	eng.RegisterMethod("file.mkdir", fileH.HandleMkdir)
	eng.RegisterMethod("file.rename", fileH.HandleRename)

	relayCfg := relay.DefaultConfig()
	relayCfg.ServerURL = *relayURL
	if *relayURL == "" && len(cfg.Network.RelayServers) > 0 {
		relayCfg.ServerURL = cfg.Network.RelayServers[0]
	}
	relayCfg.DeviceID = devID.DeviceID
	relayClient := relay.NewClient(relayCfg, logger)
	relayClient.OnMessage = func(ctx context.Context, env *bpp.Envelope) (*bpp.Envelope, error) {
		return eng.Handle(ctx, env)
	}

	// Initialize mDNS advertiser
	mdnsInfo := &mdns.ServiceInfo{
		DeviceID:       devID.DeviceID,
		FriendlyName:   cfg.Runtime.DeviceName,
		RuntimeVersion: version.Version,
		Platform:       platform,
		Port:           cfg.Network.ListenPort,
	}
	mdnsAdv := mdns.NewAdvertiser(mdnsInfo, logger)

	wsHandler := &bppWebSocketHandler{
		engine:     eng,
		identity:   devID,
		deviceName: cfg.Runtime.DeviceName,
		platform:   platform,
		version:    version.Version,
		log:        logger,
	}
	wsServer := ws.NewServer(wsHandler, cfg.Network.ListenPort, logger)

	// Initialize Connection Engine
	connEngine := connection.NewEngine(devID.DeviceID, logger)

	// Create Supervisor and register all components
	sup := supervisor.New(supervisor.DefaultConfig(), logger)
	sup.Register(store)
	sup.Register(devSvc)
	sup.Register(mdnsAdv)
	sup.Register(wsServer)
	sup.Register(connEngine)
	sup.Register(eng)
	sup.Register(termMgr)
	sup.Register(relayClient)

	// Run (blocks until shutdown signal)
	logger.Info("all components initialized, starting supervisor")
	if err := sup.Run(context.Background()); err != nil {
		logger.Error("runtime error", "error", err)
		os.Exit(1)
	}

	logger.Info("shutdown complete")
}

type handshakeState struct {
	mu       sync.Mutex
	hs       *bpp.Handshake
	ready    bool
	session  string
	deviceID string
	clientID string
}

// bppWebSocketHandler adapts the Engine Manager to the WebSocket handler interface.
type bppWebSocketHandler struct {
	engine     *engine.Manager
	pairing    *pairing.Handler
	identity   *identity.DeviceIdentity
	deviceName string
	platform   string
	version    string
	log        *slog.Logger
	states     sync.Map // connID (string) -> *handshakeState
}

func (h *bppWebSocketHandler) OnConnect(ctx context.Context, conn *ws.Connection) {
	s := &handshakeState{
		hs: bpp.NewHandshake(bpp.HandshakeConfig{
			Role:          bpp.RoleAgent,
			DeviceID:      h.identity.DeviceID,
			DeviceName:    h.deviceName,
			Platform:      h.platform,
			Version:       h.version,
			Capabilities:  bpp.DefaultCapabilities,
			TokenDuration: 24 * time.Hour,
		}),
	}
	h.states.Store(conn.ID, s)
	h.log.Info("client connected", "id", conn.ID, "remote", conn.RemoteAddr)
}

func (h *bppWebSocketHandler) OnDisconnect(ctx context.Context, conn *ws.Connection) {
	h.states.Delete(conn.ID)
	h.log.Info("client disconnected", "id", conn.ID)
}

func (h *bppWebSocketHandler) HandleMessage(ctx context.Context, data []byte) ([]byte, error) {
	connID, _ := ctx.Value(ws.ContextKeyConnID).(string)
	env, err := bpp.Unmarshal(data)
	if err != nil {
		h.log.Warn("failed to parse BPP message", "error", err)
		return nil, nil
	}

	switch env.Method {
	case bpp.MethodHandshake:
		return h.handleHandshakeMessage(connID, env)
	case bpp.MethodAuthResponse:
		return h.handleAuthResponse(connID, env)
	case "pair.initiate", "pair.verify", "pair.status", "pair.unpair":
		resp, err := h.engine.Handle(ctx, env)
		if err != nil {
			h.log.Error("engine handler error", "error", err)
			return nil, nil
		}
		if resp != nil {
			return resp.Marshal()
		}
		return nil, nil
	}

	s, ok := h.states.Load(connID)
	if !ok {
		return bpp.NewErrorResponse(env, "handshake_required", "send bpp.handshake first").Marshal()
	}
	hs := s.(*handshakeState)
	hs.mu.Lock()
	ready := hs.ready
	session := hs.session
	clientID := hs.clientID
	hs.mu.Unlock()

	if !ready {
		return bpp.NewErrorResponse(env, "handshake_required", "send bpp.handshake first").Marshal()
	}

	ctx = context.WithValue(ctx, engine.CtxSessionToken, session)
	ctx = context.WithValue(ctx, engine.CtxClientID, clientID)

	resp, err := h.engine.Handle(ctx, env)
	if err != nil {
		h.log.Error("engine handler error", "error", err)
		return nil, nil
	}
	if resp != nil {
		return resp.Marshal()
	}
	return nil, nil
}

func (h *bppWebSocketHandler) handleHandshakeMessage(connID string, env *bpp.Envelope) ([]byte, error) {
	var offer bpp.CapabilityOffer
	if err := json.Unmarshal(env.Params, &offer); err != nil {
		return bpp.NewErrorResponse(env, "invalid_params", "invalid handshake offer").Marshal()
	}

	hs, ok := h.getHandshake(connID)
	if !ok {
		return bpp.NewErrorResponse(env, "no_handshake", "no handshake state").Marshal()
	}

	resp, err := hs.hs.HandleOffer(&offer)
	if err != nil {
		return bpp.NewErrorResponse(env, "handshake_error", err.Error()).Marshal()
	}

	switch v := resp.(type) {
	case *bpp.CapabilityAccept:
		hs.mu.Lock()
		hs.ready = true
		hs.session = v.SessionToken
		hs.deviceID = v.DeviceID
		hs.mu.Unlock()
		r, e := bpp.NewResponse(env, v)
		if e != nil {
			return nil, nil
		}
		return r.Marshal()
	case *bpp.AuthChallenge:
		r, e := bpp.NewResponse(env, v)
		if e != nil {
			return nil, nil
		}
		return r.Marshal()
	}
	return bpp.NewErrorResponse(env, "unexpected", "unexpected handshake response").Marshal()
}

func (h *bppWebSocketHandler) handleAuthResponse(connID string, env *bpp.Envelope) ([]byte, error) {
	var authResp bpp.AuthResponse
	if err := json.Unmarshal(env.Params, &authResp); err != nil {
		return bpp.NewErrorResponse(env, "invalid_params", "invalid auth response").Marshal()
	}

	hs, ok := h.getHandshake(connID)
	if !ok {
		return bpp.NewErrorResponse(env, "no_handshake", "no handshake state").Marshal()
	}

	est, err := hs.hs.HandleAuthResponse(&authResp)
	if err != nil {
		return bpp.NewErrorResponse(env, "auth_failed", err.Error()).Marshal()
	}

	hs.mu.Lock()
	hs.ready = true
	hs.session = est.SessionToken
	hs.deviceID = est.DeviceID
	hs.clientID = authResp.ClientName
	if hs.clientID == "" {
		hs.clientID = authResp.DeviceID
	}
	hs.mu.Unlock()

	h.log.Info("handshake complete", "conn", connID, "device", est.DeviceID)
	r, e := bpp.NewResponse(env, est)
	if e != nil {
		return nil, nil
	}
	return r.Marshal()
}

func (h *bppWebSocketHandler) getHandshake(connID string) (*handshakeState, bool) {
	v, ok := h.states.Load(connID)
	if !ok {
		return nil, false
	}
	return v.(*handshakeState), true
}
