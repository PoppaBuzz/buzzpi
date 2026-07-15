package engine

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/buzzpi/agent/pkg/bpp"
)

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)
	m.Use(AuthMiddleware())

	called := false
	m.RegisterMethod("test.method", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		called = true
		return "ok", nil
	})

	env, _ := bpp.NewRequest("test.method", json.RawMessage(`{}`))
	_, err := m.Handle(context.Background(), env)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestRequireSessionRejects(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)
	m.Use(RequireSession())
	m.RegisterMethod("protected", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "ok", nil
	})

	env, _ := bpp.NewRequest("protected", json.RawMessage(`{}`))
	resp, err := m.Handle(context.Background(), env)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error response for missing session")
	}
}

func TestRequireSessionAllows(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)
	m.Use(RequireSession())
	m.RegisterMethod("protected", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "ok", nil
	})

	ctx := context.WithValue(context.Background(), CtxSessionToken, "valid-token")
	env, _ := bpp.NewRequest("protected", json.RawMessage(`{}`))
	resp, err := m.Handle(ctx, env)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestRequireSessionEmptyToken(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)
	m.Use(RequireSession())
	m.RegisterMethod("protected", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "ok", nil
	})

	ctx := context.WithValue(context.Background(), CtxSessionToken, "")
	env, _ := bpp.NewRequest("protected", json.RawMessage(`{}`))
	resp, err := m.Handle(ctx, env)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error response for empty session")
	}
}

func TestSessionInfoFromContext(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), CtxSessionToken, "my-token")
	token, ok := SessionInfo(ctx)
	if !ok {
		t.Fatal("should return true")
	}
	if token != "my-token" {
		t.Errorf("token = %q", token)
	}
}

func TestSessionInfoMissing(t *testing.T) {
	t.Parallel()
	_, ok := SessionInfo(context.Background())
	if ok {
		t.Error("should return false")
	}
}

func TestClientIDFromContext(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), CtxClientID, "client-123")
	id, ok := ClientIDFromContext(ctx)
	if !ok {
		t.Fatal("should return true")
	}
	if id != "client-123" {
		t.Errorf("id = %q", id)
	}
}

func TestClientIDMissing(t *testing.T) {
	t.Parallel()
	_, ok := ClientIDFromContext(context.Background())
	if ok {
		t.Error("should return false")
	}
}

func TestSetAuthLevel(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)
	m.RegisterMethod("admin.method", nil)
	m.SetAuthLevel("admin.method", AuthAdmin)

	info, ok := m.MethodInfo("admin.method")
	if !ok {
		t.Fatal("should exist")
	}
	if info.AuthLevel != AuthAdmin {
		t.Errorf("AuthLevel = %q, want %q", info.AuthLevel, AuthAdmin)
	}
}

func TestSetAuthLevelUnregistered(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)
	m.SetAuthLevel("nonexistent", AuthPublic) // should not panic
	_, ok := m.MethodInfo("nonexistent")
	if ok {
		t.Error("should not exist")
	}
}

func TestMethodInfoRegistered(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)
	m.RegisterMethod("test.method", nil)
	info, ok := m.MethodInfo("test.method")
	if !ok {
		t.Fatal("should exist")
	}
	if info.Name != "test.method" {
		t.Errorf("Name = %q", info.Name)
	}
}

func TestMiddlewareChainOrder(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)

	var order []string
	m.Use(func(ctx context.Context, method string, params json.RawMessage, next HandlerFunc) (interface{}, error) {
		order = append(order, "mw1-before")
		result, err := next(ctx, params)
		order = append(order, "mw1-after")
		return result, err
	})
	m.Use(func(ctx context.Context, method string, params json.RawMessage, next HandlerFunc) (interface{}, error) {
		order = append(order, "mw2-before")
		result, err := next(ctx, params)
		order = append(order, "mw2-after")
		return result, err
	})
	m.RegisterMethod("test", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		order = append(order, "handler")
		return "ok", nil
	})

	env, _ := bpp.NewRequest("test", json.RawMessage(`{}`))
	m.Handle(context.Background(), env)

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("order = %v, want %v", order, expected)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Errorf("order[%d] = %q, want %q", i, order[i], expected[i])
		}
	}
}

func TestMiddlewareShortCircuit(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)

	handlerCalled := false
	m.Use(func(ctx context.Context, method string, params json.RawMessage, next HandlerFunc) (interface{}, error) {
		return "blocked", nil // don't call next
	})
	m.RegisterMethod("test", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	})

	env, _ := bpp.NewRequest("test", json.RawMessage(`{}`))
	resp, _ := m.Handle(context.Background(), env)
	if handlerCalled {
		t.Error("handler should not be called")
	}
	var result string
	json.Unmarshal(resp.Result, &result)
	if result != "blocked" {
		t.Errorf("result = %q, want blocked", result)
	}
}

func TestHandleResponseHasCorrectRID(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)
	m.RegisterMethod("test", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "ok", nil
	})

	env, _ := bpp.NewRequest("test", json.RawMessage(`{}`))
	resp, _ := m.Handle(context.Background(), env)
	if resp.RID != env.RID {
		t.Errorf("RID = %q, want %q", resp.RID, env.RID)
	}
}

func TestHandleErrorResponseHasCorrectRID(t *testing.T) {
	t.Parallel()
	m := newTestManager(t)
	m.RegisterMethod("test", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return nil, json.Unmarshal([]byte("bad"), nil)
	})

	env, _ := bpp.NewRequest("test", json.RawMessage(`{}`))
	resp, _ := m.Handle(context.Background(), env)
	if resp.RID != env.RID {
		t.Errorf("RID = %q, want %q", resp.RID, env.RID)
	}
}
