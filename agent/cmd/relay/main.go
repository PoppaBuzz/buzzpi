// Command relay starts the BuzzPi Cloud Relay server.
//
// The relay bridges BPP messages between agents (runtimes) and clients (CLI,
// Android) over WebSocket, enabling off-LAN device access.
//
// Usage:
//
//	buzzpi-relay                    # listen on :8080
//	buzzpi-relay --port 9090        # listen on :9090
//	buzzpi-relay --listen :9090     # listen on :9090
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/buzzpi/agent/internal/relay"
)

type Config struct {
	Listen       string `json:"listen"`
	LogLevel     string `json:"log_level"`
	DrainTimeout string `json:"drain_timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		Listen:       ":8080",
		LogLevel:     "info",
		DrainTimeout: "30s",
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func main() {
	configPath := flag.String("config", "", "path to config file (optional)")
	port := flag.Int("port", 0, "listen port (overrides config)")
	listen := flag.String("listen", "", "listen address (overrides config and --port)")
	logLevel := flag.String("log-level", "", "log level: debug, info, warn, error (overrides config)")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *listen != "" {
		cfg.Listen = *listen
	} else if *port != 0 {
		cfg.Listen = fmt.Sprintf(":%d", *port)
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	drainTimeout := 30 * time.Second
	if cfg.DrainTimeout != "" {
		if d, err := time.ParseDuration(cfg.DrainTimeout); err == nil {
			drainTimeout = d
		}
	}

	server := relay.NewServer(cfg.Listen, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("shutting down", "signal", sig, "drain_timeout", drainTimeout)
		cancel()
	}()

	logger.Info("relay server starting", "listen", cfg.Listen, "drain_timeout", drainTimeout)
	if err := server.ListenAndServe(ctx); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	logger.Info("relay server stopped")
}
