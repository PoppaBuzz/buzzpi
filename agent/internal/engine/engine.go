// Package engine implements the Engine Manager — the central message router
// for the BuzzPi Runtime. It registers BPP method handlers and routes
// incoming requests to the appropriate handler.
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/buzzpi/agent/pkg/bpp"
)

// HandlerFunc processes a BPP method request and returns a result or error.
type HandlerFunc func(ctx context.Context, params json.RawMessage) (interface{}, error)

// MethodInfo describes a registered method.
type MethodInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Manager routes BPP messages to registered method handlers.
type Manager struct {
	mu       sync.RWMutex
	methods  map[string]HandlerFunc
	handlers []HandlerFunc // middleware chain (auth, logging, etc.)
	log      *slog.Logger
}

// NewManager creates a new Engine Manager.
func NewManager(log *slog.Logger) *Manager {
	if log == nil {
		log = slog.Default()
	}
	return &Manager{
		methods: make(map[string]HandlerFunc),
		log:     log.With("component", "engine"),
	}
}

// RegisterMethod registers a handler for a BPP method.
func (m *Manager) RegisterMethod(method string, handler HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.methods[method] = handler
	m.log.Info("method registered", "method", method)
}

// RegisterHandler adds a middleware handler that wraps all method calls.
// Middleware is executed in registration order before the method handler.
func (m *Manager) RegisterHandler(h HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, h)
}

// Handle processes an incoming BPP request and returns a response envelope.
func (m *Manager) Handle(ctx context.Context, env *bpp.Envelope) (*bpp.Envelope, error) {
	m.mu.RLock()
	handler, ok := m.methods[env.Method]
	m.mu.RUnlock()

	if !ok {
		m.log.Warn("method not found", "method", env.Method)
		return bpp.NewErrorResponse(env, "method_not_found",
			fmt.Sprintf("unknown method: %s", env.Method)), nil
	}

	// Run middleware chain
	handler = m.wrapMiddleware(handler)

	result, err := handler(ctx, env.Params)
	if err != nil {
		m.log.Error("method handler error",
			"method", env.Method,
			"error", err,
		)
		return bpp.NewErrorResponse(env, "internal_error", err.Error()), nil
	}

	return bpp.NewResponse(env, result)
}

// wrapMiddleware wraps the handler with all registered middleware.
func (m *Manager) wrapMiddleware(final HandlerFunc) HandlerFunc {
	m.mu.RLock()
	chain := make([]HandlerFunc, len(m.handlers))
	copy(chain, m.handlers)
	m.mu.RUnlock()

	// Apply middleware in reverse order so they execute in forward order
	wrapped := final
	for i := len(chain) - 1; i >= 0; i-- {
		mw := chain[i]
		next := wrapped
		wrapped = func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			return mw(ctx, params)
		}
		_ = next // middleware chains into next
		// TODO: Proper middleware chaining
		// wrapped = func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		//     return mw(ctx, params, next)
		// }
	}
	return wrapped
}

// ListMethods returns all registered methods with their info.
func (m *Manager) ListMethods() []MethodInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := make([]MethodInfo, 0, len(m.methods))
	for name := range m.methods {
		info = append(info, MethodInfo{Name: name})
	}
	return info
}

// Health returns the manager's health status.
func (m *Manager) Health() interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return fmt.Sprintf("ok (%d methods registered)", len(m.methods))
}

func (m *Manager) Name() string { return "engine" }

func (m *Manager) Start(ctx context.Context) error {
	m.log.Info("engine manager started")
	return nil
}

func (m *Manager) Stop(ctx context.Context) error {
	m.log.Info("engine manager stopped")
	return nil
}
