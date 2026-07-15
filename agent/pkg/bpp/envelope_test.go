package bpp

import (
	"encoding/json"
	"testing"
)

func TestErrorObj_Error(t *testing.T) {
	t.Parallel()

	e := &ErrorObj{
		Code:    "METHOD_NOT_FOUND",
		Message: "method unknown: foo.bar",
	}
	got := e.Error()
	want := "METHOD_NOT_FOUND: method unknown: foo.bar"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestNewRequest(t *testing.T) {
	t.Parallel()

	t.Run("with params", func(t *testing.T) {
		params := map[string]string{"key": "value"}
		env, err := NewRequest("test.method", params)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		if env.Version != CurrentVersion {
			t.Errorf("Version = %d, want %d", env.Version, CurrentVersion)
		}
		if env.Type != TypeRequest {
			t.Errorf("Type = %q, want %q", env.Type, TypeRequest)
		}
		if env.Method != "test.method" {
			t.Errorf("Method = %q, want %q", env.Method, "test.method")
		}
		if env.ID == "" {
			t.Error("ID should not be empty")
		}
		if env.TS == "" {
			t.Error("TS should not be empty")
		}
		if env.RID == "" {
			t.Error("RID should not be empty")
		}
		if env.Params == nil {
			t.Error("Params should not be nil")
		}
		// Verify params can be unmarshaled.
		var m map[string]string
		if err := json.Unmarshal(env.Params, &m); err != nil {
			t.Fatalf("unmarshal params: %v", err)
		}
		if m["key"] != "value" {
			t.Errorf("params[key] = %q, want %q", m["key"], "value")
		}
	})

	t.Run("nil params", func(t *testing.T) {
		env, err := NewRequest("test.method", nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		if env.Params != nil {
			t.Error("Params should be nil for nil params input")
		}
	})

	t.Run("ID prefix", func(t *testing.T) {
		env, _ := NewRequest("test", nil)
		if len(env.ID) < 5 || env.ID[:4] != "msg_" {
			t.Errorf("ID = %q, want msg_ prefix", env.ID)
		}
		if len(env.RID) < 5 || env.RID[:4] != "req_" {
			t.Errorf("RID = %q, want req_ prefix", env.RID)
		}
	})
}

func TestNewResponse(t *testing.T) {
	t.Parallel()

	t.Run("with result", func(t *testing.T) {
		req, _ := NewRequest("test", nil)
		result := map[string]string{"status": "ok"}
		env, err := NewResponse(req, result)
		if err != nil {
			t.Fatalf("NewResponse: %v", err)
		}
		if env.Type != TypeResponse {
			t.Errorf("Type = %q, want %q", env.Type, TypeResponse)
		}
		if env.RID != req.RID {
			t.Errorf("RID = %q, want %q (match request)", env.RID, req.RID)
		}
		if env.Params != nil {
			t.Error("Params should be nil for response")
		}
		var m map[string]string
		if err := json.Unmarshal(env.Result, &m); err != nil {
			t.Fatalf("unmarshal result: %v", err)
		}
		if m["status"] != "ok" {
			t.Errorf("result[status] = %q", m["status"])
		}
	})

	t.Run("nil result", func(t *testing.T) {
		req, _ := NewRequest("test", nil)
		env, err := NewResponse(req, nil)
		if err != nil {
			t.Fatalf("NewResponse: %v", err)
		}
		if env.Result != nil {
			t.Error("Result should be nil for nil result input")
		}
	})
}

func TestNewErrorResponse(t *testing.T) {
	t.Parallel()

	req, _ := NewRequest("test", nil)
	env := NewErrorResponse(req, "NOT_FOUND", "resource not found")
	if env.Type != TypeError {
		t.Errorf("Type = %q, want %q", env.Type, TypeError)
	}
	if env.RID != req.RID {
		t.Errorf("RID = %q, want %q", env.RID, req.RID)
	}
	if env.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if env.Error.Code != "NOT_FOUND" {
		t.Errorf("Error.Code = %q, want %q", env.Error.Code, "NOT_FOUND")
	}
	if env.Error.Message != "resource not found" {
		t.Errorf("Error.Message = %q, want %q", env.Error.Message, "resource not found")
	}
}

func TestNewEvent(t *testing.T) {
	t.Parallel()

	t.Run("with params", func(t *testing.T) {
		params := map[string]int{"value": 42}
		env, err := NewEvent("test.event", params)
		if err != nil {
			t.Fatalf("NewEvent: %v", err)
		}
		if env.Type != TypeEvent {
			t.Errorf("Type = %q, want %q", env.Type, TypeEvent)
		}
		if env.Method != "test.event" {
			t.Errorf("Method = %q, want %q", env.Method, "test.event")
		}
		if env.RID != "" {
			t.Error("RID should be empty for events")
		}
	})

	t.Run("nil params", func(t *testing.T) {
		env, err := NewEvent("test.event", nil)
		if err != nil {
			t.Fatalf("NewEvent: %v", err)
		}
		if env.Params != nil {
			t.Error("Params should be nil for nil params input")
		}
	})
}

func TestEnvelopeMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	t.Run("round-trip request", func(t *testing.T) {
		env, _ := NewRequest("test.method", map[string]string{"hello": "world"})
		data, err := env.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := Unmarshal(data)
		if err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if got.Type != TypeRequest {
			t.Errorf("Type = %q", got.Type)
		}
		if got.Method != "test.method" {
			t.Errorf("Method = %q", got.Method)
		}
		if got.ID != env.ID {
			t.Errorf("ID = %q, want %q", got.ID, env.ID)
		}
	})

	t.Run("unsupported version", func(t *testing.T) {
		env := &Envelope{Version: 99, Type: TypeRequest}
		data, _ := json.Marshal(env)
		_, err := Unmarshal(data)
		if err == nil {
			t.Fatal("expected error for unsupported version")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := Unmarshal([]byte("{invalid json"))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("round-trip error response", func(t *testing.T) {
		req, _ := NewRequest("test", nil)
		env := NewErrorResponse(req, "CODE", "msg")
		data, err := env.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := Unmarshal(data)
		if err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if got.Error == nil {
			t.Fatal("Error should not be nil")
		}
		if got.Error.Code != "CODE" {
			t.Errorf("Error.Code = %q", got.Error.Code)
		}
	})
}

func TestEnvelopeConstants(t *testing.T) {
	t.Parallel()

	if CurrentVersion != 1 {
		t.Errorf("CurrentVersion = %d, want 1", CurrentVersion)
	}
	if TypeRequest != "request" {
		t.Errorf("TypeRequest = %q", TypeRequest)
	}
	if TypeResponse != "response" {
		t.Errorf("TypeResponse = %q", TypeResponse)
	}
	if TypeEvent != "event" {
		t.Errorf("TypeEvent = %q", TypeEvent)
	}
	if TypeError != "error" {
		t.Errorf("TypeError = %q", TypeError)
	}
}

func TestErrorObjWithData(t *testing.T) {
	t.Parallel()

	e := &ErrorObj{
		Code:    "TEST_ERROR",
		Message: "something went wrong",
		Data:    map[string]string{"detail": "extra info"},
	}
	got := e.Error()
	if got != "TEST_ERROR: something went wrong" {
		t.Errorf("Error() = %q", got)
	}
}

func TestNewRequestMarshalError(t *testing.T) {
	t.Parallel()

	ch := make(chan int)
	_, err := NewRequest("test", ch)
	if err == nil {
		t.Fatal("expected error for unmarshalable params")
	}
}

func TestNewResponseMarshalError(t *testing.T) {
	t.Parallel()

	req, _ := NewRequest("test", nil)
	ch := make(chan int)
	_, err := NewResponse(req, ch)
	if err == nil {
		t.Fatal("expected error for unmarshalable result")
	}
}
