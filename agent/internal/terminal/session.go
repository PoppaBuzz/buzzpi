package terminal

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNoSession   = errors.New("session not found")
	ErrUnsupported = errors.New("terminal not supported on this platform")
)

// Session represents a single PTY terminal session.
type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	cmd       process
	pty       ptyFile
	closed    bool
	rows      uint16
	cols      uint16
	mu        sync.Mutex       // protects pty reads, writes, and closed
	done      chan struct{}     // closed when session closes
	sender    func([]byte) error // push output to client (nil if no client)
}

// process is the interface for the underlying OS process.
type process interface {
	Kill() error
	Wait() error
}

// ptyFile is the interface for the pseudo-terminal file.
type ptyFile interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}
