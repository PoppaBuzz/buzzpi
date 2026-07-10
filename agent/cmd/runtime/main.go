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
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/buzzpi/agent/internal/config"
	"github.com/buzzpi/agent/internal/connection"
	"github.com/buzzpi/agent/internal/device"
	"github.com/buzzpi/agent/internal/engine"
	"github.com/buzzpi/agent/internal/mdns"
	"github.com/buzzpi/agent/internal/state"
	"github.com/buzzpi/agent/internal/supervisor"
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

	// Initialize Engine Manager and register built-in methods
	eng := engine.NewManager(logger)

	// Initialize Device service
	devSvc, err := device.NewService(cfg, logger)
	if err != nil {
		logger.Error("failed to initialize device service", "error", err)
		os.Exit(1)
	}
	eng.RegisterMethod("device.info", devSvc.HandleInfo)
	eng.RegisterMethod("device.stats", devSvc.HandleStats)

	// Initialize mDNS advertiser
	mdnsInfo := &mdns.ServiceInfo{
		DeviceID:       "dev_unknown", // TODO: read from store
		FriendlyName:   cfg.Runtime.DeviceName,
		RuntimeVersion: version.Version,
		Platform:       "go/" + version.GoVersion(),
		Port:           cfg.Network.ListenPort,
	}
	mdnsAdv := mdns.NewAdvertiser(mdnsInfo, logger)

	// Initialize WebSocket handler that routes to Engine Manager
	wsHandler := &bppWebSocketHandler{
		engine: eng,
		log:    logger,
	}
	wsServer := ws.NewServer(wsHandler, cfg.Network.ListenPort, logger)

	// Initialize Connection Engine
	connEngine := connection.NewEngine("dev_unknown", logger)

	// Create Supervisor and register all components
	sup := supervisor.New(supervisor.DefaultConfig(), logger)
	sup.Register(store)
	sup.Register(devSvc)
	sup.Register(mdnsAdv)
	sup.Register(wsServer)
	sup.Register(connEngine)
	sup.Register(eng)

	// Run (blocks until shutdown signal)
	logger.Info("all components initialized, starting supervisor")
	if err := sup.Run(context.Background()); err != nil {
		logger.Error("runtime error", "error", err)
		os.Exit(1)
	}

	logger.Info("shutdown complete")
}

// bppWebSocketHandler adapts the Engine Manager to the WebSocket handler interface.
type bppWebSocketHandler struct {
	engine *engine.Manager
	log    *slog.Logger
}

func (h *bppWebSocketHandler) HandleMessage(ctx context.Context, data []byte) ([]byte, error) {
	// Parse BPP envelope
	env, err := bpp.Unmarshal(data)
	if err != nil {
		h.log.Warn("failed to parse BPP message", "error", err)
		return nil, nil // don't respond to malformed messages
	}

	// Route through engine manager
	response, err := h.engine.Handle(ctx, env)
	if err != nil {
		h.log.Error("engine handler error", "error", err)
		return nil, nil
	}

	if response != nil {
		return response.Marshal()
	}
	return nil, nil
}

func (h *bppWebSocketHandler) OnConnect(ctx context.Context, conn *ws.Connection) {
	h.log.Info("client connected", "id", conn.ID, "remote", conn.RemoteAddr)
}

func (h *bppWebSocketHandler) OnDisconnect(ctx context.Context, conn *ws.Connection) {
	h.log.Info("client disconnected", "id", conn.ID)
}
