//go:build windows

package terminal

import "time"

// startPTY is a no-op stub for Windows.
// PTY terminal sessions are only supported on Unix-like platforms.
func startPTY(shell string) (process, ptyFile, error) {
	return nil, nil, ErrUnsupported
}

// resizePTY is a no-op stub for Windows.
func resizePTY(f ptyFile, rows, cols uint16) error {
	return ErrUnsupported
}

// Create returns ErrUnsupported on Windows.
func Create(id, shell string) (*Session, error) {
	return nil, ErrUnsupported
}

// Read returns ErrUnsupported on Windows.
func (s *Session) Read() ([]byte, error) {
	return nil, ErrUnsupported
}

// ReadOutput returns ErrUnsupported on Windows.
func (s *Session) ReadOutput(timeout time.Duration) ([]byte, error) {
	return nil, ErrUnsupported
}

// Write returns ErrUnsupported on Windows.
func (s *Session) Write(data []byte) (int, error) {
	return 0, ErrUnsupported
}

// Resize returns ErrUnsupported on Windows.
func (s *Session) Resize(rows, cols uint16) error {
	return ErrUnsupported
}

// Close is a no-op on Windows.
func (s *Session) Close() error {
	return nil
}

// Wait is a no-op on Windows.
func (s *Session) Wait() error {
	return nil
}

// StartOutputLoop is a no-op on Windows.
func (s *Session) StartOutputLoop(sender func([]byte) error) {
}
