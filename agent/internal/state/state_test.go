package state

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	s, err := Open(path, slog.Default())
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestOpen(t *testing.T) {
	t.Run("creates store with all buckets", func(t *testing.T) {
		s := newTestStore(t)
		if s.db == nil {
			t.Fatal("store.db is nil after Open")
		}
		if s.Path() == "" {
			t.Error("store.Path() is empty")
		}
	})

	t.Run("with nil logger uses default", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test_nil_logger.db")
		s, err := Open(path, nil)
		if err != nil {
			t.Fatalf("Open(nil logger) failed: %v", err)
		}
		s.Close()
	})

	t.Run("file persists on disk", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "persist.db")
		s, err := Open(path, nil)
		if err != nil {
			t.Fatalf("Open() failed: %v", err)
		}
		s.Close()
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("database file was not created on disk")
		}
	})
}

func TestWriteRead(t *testing.T) {
	s := newTestStore(t)

	t.Run("write and read round-trip", func(t *testing.T) {
		type testVal struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		val := testVal{ID: 42, Name: "buzzpi"}

		if err := s.Write("state", "testkey", val); err != nil {
			t.Fatalf("Write() failed: %v", err)
		}

		var got testVal
		if err := s.Read("state", "testkey", &got); err != nil {
			t.Fatalf("Read() failed: %v", err)
		}
		if got.ID != 42 || got.Name != "buzzpi" {
			t.Errorf("Read got %+v, want {ID:42 Name:buzzpi}", got)
		}
	})

	t.Run("read non-existent key returns error", func(t *testing.T) {
		var v interface{}
		err := s.Read("state", "nonexistent", &v)
		if err == nil {
			t.Error("Read(nonexistent) expected error, got nil")
		}
	})

	t.Run("write to unknown bucket returns error", func(t *testing.T) {
		err := s.Write("nobucket", "k", "v")
		if err == nil {
			t.Error("Write(unknown bucket) expected error, got nil")
		}
	})
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)

	if err := s.Write("state", "delkey", "delete me"); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	if err := s.Delete("state", "delkey"); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	var v interface{}
	err := s.Read("state", "delkey", &v)
	if err == nil {
		t.Error("Read after Delete expected error, got nil")
	}
}

func TestList(t *testing.T) {
	s := newTestStore(t)

	items, err := s.List("state")
	if err != nil {
		t.Fatalf("List(empty) failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("List(empty) returned %d items, want 0", len(items))
	}

	if err := s.Write("state", "k1", "v1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Write("state", "k2", "v2"); err != nil {
		t.Fatal(err)
	}

	items, err = s.List("state")
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("List() returned %d items, want 2", len(items))
	}
	if string(items["k1"]) != `"v1"` {
		t.Errorf("items[k1] = %s, want \"v1\"", string(items["k1"]))
	}
}

func TestSessions(t *testing.T) {
	s := newTestStore(t)

	t.Run("SaveSession and GetSession", func(t *testing.T) {
		session := &Session{
			SessionID:  "sess_001",
			DeviceID:   "dev_test",
			ClientID:   "client_abc",
			ClientName: "Test Client",
			Role:       "user",
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(time.Hour),
			LastUsed:   time.Now(),
		}

		if err := s.SaveSession(session); err != nil {
			t.Fatalf("SaveSession() failed: %v", err)
		}

		got, err := s.GetSession("sess_001")
		if err != nil {
			t.Fatalf("GetSession() failed: %v", err)
		}
		if got.SessionID != "sess_001" || got.ClientID != "client_abc" {
			t.Errorf("GetSession got SessionID=%q ClientID=%q", got.SessionID, got.ClientID)
		}
	})

	t.Run("GetSession non-existent", func(t *testing.T) {
		_, err := s.GetSession("sess_nonexistent")
		if err == nil {
			t.Error("GetSession(nonexistent) expected error")
		}
	})

	t.Run("DeleteSession", func(t *testing.T) {
		session := &Session{
			SessionID: "sess_del",
			DeviceID:  "dev_del",
			ClientID:  "client_del",
		}
		if err := s.SaveSession(session); err != nil {
			t.Fatal(err)
		}
		if err := s.DeleteSession("sess_del"); err != nil {
			t.Fatalf("DeleteSession() failed: %v", err)
		}
		_, err := s.GetSession("sess_del")
		if err == nil {
			t.Error("GetSession after DeleteSession expected error")
		}
	})
}

func TestSavePairing(t *testing.T) {
	s := newTestStore(t)

	pairing := &Pairing{
		DeviceID:  "dev_pair",
		AccountID: "acct_001",
		ClientID:  "cli_001",
		PairedAt:  time.Now(),
	}

	if err := s.SavePairing(pairing); err != nil {
		t.Fatalf("SavePairing() failed: %v", err)
	}

	// Verify it was stored by listing pairs bucket
	items, err := s.List("pairs")
	if err != nil {
		t.Fatalf("List(pairs) failed: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("List(pairs) = %d items, want 1", len(items))
	}
}

func TestClose(t *testing.T) {
	t.Run("Close is idempotent", func(t *testing.T) {
		s := newTestStore(t)
		if err := s.Close(); err != nil {
			t.Fatalf("first Close() failed: %v", err)
		}
		if err := s.Close(); err != nil {
			t.Fatalf("second Close() should be no-op: %v", err)
		}
	})
}

func TestSize(t *testing.T) {
	s := newTestStore(t)

	sz, err := s.Size()
	if err != nil {
		t.Fatalf("Size() failed: %v", err)
	}
	if sz <= 0 {
		t.Errorf("Size() = %d, want > 0", sz)
	}

	// Write data and verify size increases
	if err := s.Write("state", "bigkey", map[string]string{"data": "x" + string(make([]byte, 1000))}); err != nil {
		t.Fatal(err)
	}
	sz2, err := s.Size()
	if err != nil {
		t.Fatal(err)
	}
	if sz2 <= sz {
		t.Errorf("Size() after write = %d, should be > %d", sz2, sz)
	}
}

func TestPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pathcheck.db")
	s, err := Open(path, nil)
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer s.Close()

	if s.Path() != path {
		t.Errorf("Path() = %q, want %q", s.Path(), path)
	}
}

func TestHealth(t *testing.T) {
	s := newTestStore(t)

	h := s.Health()
	m, ok := h.(map[string]interface{})
	if !ok {
		t.Fatalf("Health() returned %T, want map[string]interface{}", h)
	}
	if m["status"] != "ok" {
		t.Errorf("Health() status = %v, want \"ok\"", m["status"])
	}
	if _, ok := m["size_bytes"]; !ok {
		t.Error("Health() missing size_bytes")
	}
}

func TestName(t *testing.T) {
	s := newTestStore(t)
	if s.Name() != "state-store" {
		t.Errorf("Name() = %q, want \"state-store\"", s.Name())
	}
}

func TestStartStop(t *testing.T) {
	s := newTestStore(t)

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	// Stop calls Close internally
	if err := s.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
}
