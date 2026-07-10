package engine

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/buzzpi/agent/pkg/bpp"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	return NewManager(slog.Default())
}

func TestNewManager(t *testing.T) {
	m := newTestManager(t)
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	methods := m.ListMethods()
	if len(methods) != 0 {
		t.Errorf("NewManager() has %d methods, want 0", len(methods))
	}

	t.Run("nil logger uses default", func(t *testing.T) {
		m2 := NewManager(nil)
		if m2 == nil {
			t.Fatal("NewManager(nil) returned nil")
		}
	})
}

func TestRegisterMethod(t *testing.T) {
	m := newTestManager(t)

	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "ok", nil
	}

	m.RegisterMethod("test.method", handler)

	methods := m.ListMethods()
	if len(methods) != 1 {
		t.Fatalf("ListMethods() = %d, want 1", len(methods))
	}
	if methods[0].Name != "test.method" {
		t.Errorf("Method name = %q, want \"test.method\"", methods[0].Name)
	}
}

func TestHandle(t *testing.T) {
	m := newTestManager(t)

	t.Run("dispatches to correct handler", func(t *testing.T) {
		m.RegisterMethod("ping", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			return "pong", nil
		})

		env, reqErr := bpp.NewRequest("ping", json.RawMessage(`{}`))
		if reqErr != nil {
			t.Fatalf("NewRequest() failed: %v", reqErr)
		}
		resp, err := m.Handle(context.Background(), env)
		if err != nil {
			t.Fatalf("Handle() returned error: %v", err)
		}
		if resp == nil {
			t.Fatal("Handle() returned nil response")
		}
	})

	t.Run("unknown method returns error response", func(t *testing.T) {
		env, reqErr := bpp.NewRequest("unknown.method", json.RawMessage(`{}`))
		if reqErr != nil {
			t.Fatalf("NewRequest() failed: %v", reqErr)
		}
		resp, err := m.Handle(context.Background(), env)
		if err != nil {
			t.Fatalf("Handle(unknown) returned error: %v", err)
		}
		if resp == nil {
			t.Fatal("Handle() returned nil response")
		}
	})

	t.Run("handler error returns error response", func(t *testing.T) {
		m.RegisterMethod("error.method", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			return nil, errors.New("handler failed")
		})

		env, reqErr := bpp.NewRequest("error.method", json.RawMessage(`{}`))
		if reqErr != nil {
			t.Fatalf("NewRequest() failed: %v", reqErr)
		}
		resp, err := m.Handle(context.Background(), env)
		if err != nil {
			t.Fatalf("Handle() returned error: %v", err)
		}
		if resp == nil {
			t.Fatal("Handle() returned nil response")
		}
	})
}

func TestListMethods(t *testing.T) {
	m := newTestManager(t)

	m.RegisterMethod("a", nil)
	m.RegisterMethod("b", nil)
	m.RegisterMethod("c", nil)

	methods := m.ListMethods()
	if len(methods) != 3 {
		t.Errorf("ListMethods() = %d, want 3", len(methods))
	}

	names := make(map[string]bool)
	for _, mi := range methods {
		names[mi.Name] = true
	}
	for _, name := range []string{"a", "b", "c"} {
		if !names[name] {
			t.Errorf("ListMethods() missing %q", name)
		}
	}
}

func TestHealth(t *testing.T) {
	m := newTestManager(t)
	m.RegisterMethod("test", nil)

	h := m.Health()
	s, ok := h.(string)
	if !ok {
		t.Fatalf("Health() returned %T, want string", h)
	}
	if s != "ok (1 methods registered)" {
		t.Errorf("Health() = %q, want \"ok (1 methods registered)\"", s)
	}
}

func TestName(t *testing.T) {
	m := newTestManager(t)
	if m.Name() != "engine" {
		t.Errorf("Name() = %q, want \"engine\"", m.Name())
	}
}

func TestStartStop(t *testing.T) {
	m := newTestManager(t)

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	if err := m.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
}
