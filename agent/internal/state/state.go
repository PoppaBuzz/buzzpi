// Package state provides an embedded key-value store for runtime state.
// Uses BoltDB for ACID-compliant, single-file persistence.
package state

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

// Common bucket names.
var (
	BucketPairs     = []byte("pairs")
	BucketSessions  = []byte("sessions")
	BucketConfig    = []byte("config")
	BucketPlugins   = []byte("plugins")
	BucketState     = []byte("state")
	BucketTelemetry = []byte("telemetry")
)

// Store provides key-value access to the runtime state.
type Store struct {
	db     *bbolt.DB
	path   string
	log    *slog.Logger
	closeOnce sync.Once
}

// Session represents an active client session stored in BoltDB.
type Session struct {
	SessionID  string    `json:"session_id"`
	DeviceID   string    `json:"device_id"`
	ClientID   string    `json:"client_id"`
	ClientName string    `json:"client_name"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsed   time.Time `json:"last_used"`
}

// Pairing represents a paired client.
type Pairing struct {
	DeviceID   string    `json:"device_id"`
	AccountID  string    `json:"account_id"`
	ClientID   string    `json:"client_id"`
	ClientName string    `json:"client_name"`
	Role       string    `json:"role"`
	PairedAt   time.Time `json:"paired_at"`
	PublicKey  string    `json:"public_key"`
}

// Open opens or creates a BoltDB database at the given path.
func Open(path string, log *slog.Logger) (*Store, error) {
	if log == nil {
		log = slog.Default()
	}
	db, err := bbolt.Open(path, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("open state store: %w", err)
	}

	store := &Store{db: db, path: path, log: log.With("component", "state-store")}

	// Ensure buckets exist
	if err := store.initBuckets(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) initBuckets() error {
	buckets := [][]byte{
		BucketPairs, BucketSessions, BucketConfig,
		BucketPlugins, BucketState, BucketTelemetry,
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("create bucket %s: %w", string(bucket), err)
			}
		}
		return nil
	})
}

// Write stores a value in the specified bucket.
func (s *Store) Write(bucket, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}
		return b.Put([]byte(key), data)
	})
}

// Read retrieves a value from the specified bucket and unmarshals it into out.
func (s *Store) Read(bucket, key string, out interface{}) error {
	return s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key not found: %s/%s", bucket, key)
		}
		return json.Unmarshal(data, out)
	})
}

// Delete removes a key from the specified bucket.
func (s *Store) Delete(bucket, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}
		return b.Delete([]byte(key))
	})
}

// List returns all key-value pairs in a bucket.
func (s *Store) List(bucket string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}
		return b.ForEach(func(k, v []byte) error {
			result[string(k)] = append([]byte{}, v...)
			return nil
		})
	})
	return result, err
}

// SaveSession stores a session in the sessions bucket.
func (s *Store) SaveSession(session *Session) error {
	return s.Write("sessions", session.SessionID, session)
}

// GetSession retrieves a session by token.
func (s *Store) GetSession(sessionID string) (*Session, error) {
	var session Session
	if err := s.Read("sessions", sessionID, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteSession removes a session.
func (s *Store) DeleteSession(sessionID string) error {
	return s.Delete("sessions", sessionID)
}

// SavePairing stores a pairing record.
func (s *Store) SavePairing(pairing *Pairing) error {
	key := pairing.DeviceID + "/" + pairing.ClientID
	return s.Write("pairs", key, pairing)
}

// Close closes the database.
func (s *Store) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.db.Close()
	})
	return err
}

// Path returns the database file path.
func (s *Store) Path() string {
	return s.path
}

// Size returns the database file size in bytes.
func (s *Store) Size() (int64, error) {
	var size int64
	err := s.db.View(func(tx *bbolt.Tx) error {
		size = tx.Size()
		return nil
	})
	return size, err
}

func (s *Store) Name() string { return "state-store" }

func (s *Store) Start(ctx context.Context) error { return nil }

func (s *Store) Stop(ctx context.Context) error {
	s.log.Info("closing state store")
	return s.Close()
}

func (s *Store) Health() interface{} {
	sz, err := s.Size()
	if err != nil {
		return map[string]interface{}{"status": "error", "error": err.Error()}
	}
	return map[string]interface{}{"status": "ok", "size_bytes": sz}
}
