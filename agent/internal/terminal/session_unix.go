//go:build !windows

package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

// cmdProcess wraps *exec.Cmd to implement the process interface.
// This avoids cross-compilation issues where *exec.Cmd method sets
// may not satisfy the interface on non-host platforms.
type cmdProcess struct {
	cmd *exec.Cmd
}

func (p *cmdProcess) Kill() error { return p.cmd.Process.Kill() }
func (p *cmdProcess) Wait() error { return p.cmd.Wait() }

// startPTY starts a new shell process with a PTY.
func startPTY(shell string) (process, ptyFile, error) {
	if shell == "" {
		shell = defaultShell()
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	f, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("start pty: %w", err)
	}

	return &cmdProcess{cmd: cmd}, f, nil
}

// resizePTY changes the PTY window size.
func resizePTY(f ptyFile, rows, cols uint16) error {
	p, ok := f.(*os.File)
	if !ok {
		return fmt.Errorf("not a pty file")
	}
	return pty.Setsize(p, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
}

func defaultShell() string {
	for _, s := range []string{"/bin/bash", "/bin/sh", "/usr/bin/zsh"} {
		if _, err := os.Stat(s); err == nil {
			return s
		}
	}
	return "/bin/sh"
}

// Create starts a new PTY terminal session.
// timeout sets the inactivity timeout; 0 disables auto-close.
func Create(id, shell string, timeout time.Duration) (*Session, error) {
	cmd, f, err := startPTY(shell)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	s := &Session{
		ID:                id,
		CreatedAt:         now,
		cmd:               cmd,
		pty:               f,
		rows:              24,
		cols:              80,
		done:              make(chan struct{}),
		lastActivity:      now,
		inactivityTimeout: timeout,
	}

	if timeout > 0 {
		s.startIdleChecker()
	}

	return s, nil
}

// Read reads output from the PTY (blocking).
func (s *Session) Read() ([]byte, error) {
	if s.closed {
		return nil, fmt.Errorf("session closed")
	}
	buf := make([]byte, 4096)
	n, err := s.pty.Read(buf)
	if n > 0 {
		s.touchActivity()
		return buf[:n], err
	}
	return nil, err
}

// ReadOutput reads available PTY output with a timeout.
func (s *Session) ReadOutput(timeout time.Duration) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fmt.Errorf("session closed")
	}

	f, ok := s.pty.(*os.File)
	if !ok {
		return nil, fmt.Errorf("pty is not an os.File")
	}

	f.SetReadDeadline(time.Now().Add(timeout))
	defer f.SetReadDeadline(time.Time{})

	buf := make([]byte, 4096)
	n, err := f.Read(buf)
	if n > 0 {
		s.touchActivity()
		return buf[:n], nil
	}
	if os.IsTimeout(err) {
		return nil, nil
	}
	return nil, err
}

// Write sends input to the PTY.
func (s *Session) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, fmt.Errorf("session closed")
	}
	n, err := s.pty.Write(data)
	if n > 0 {
		s.touchActivity()
	}
	return n, err
}

// Resize sets the PTY terminal dimensions.
func (s *Session) Resize(rows, cols uint16) error {
	if s.closed {
		return fmt.Errorf("session closed")
	}
	if err := resizePTY(s.pty, rows, cols); err != nil {
		return fmt.Errorf("resize: %w", err)
	}
	s.rows = rows
	s.cols = cols
	return nil
}

// Close terminates the PTY session.
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	close(s.done)
	s.pty.Close()
	return s.cmd.Kill()
}

// StartOutputLoop reads PTY output and pushes it to the client via the sender.
func (s *Session) StartOutputLoop(sender func([]byte) error) {
	s.sender = sender
	go func() {
		buf := make([]byte, 4096)
		for {
			select {
			case <-s.done:
				return
			default:
			}

			s.mu.Lock()
			if s.closed {
				s.mu.Unlock()
				return
			}
			if f, ok := s.pty.(*os.File); ok {
				f.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			}
			n, readErr := s.pty.Read(buf)
			if f, ok := s.pty.(*os.File); ok {
				f.SetReadDeadline(time.Time{})
			}
			s.mu.Unlock()

			if n > 0 {
				eventParams := map[string]interface{}{
					"session_id": s.ID,
					"data":       string(buf[:n]),
				}
				paramsJSON, _ := json.Marshal(eventParams)
				event := map[string]interface{}{
					"v":      1,
					"id":     "evt_" + fmt.Sprintf("%d", time.Now().UnixNano()),
					"ts":     time.Now().UTC().Format(time.RFC3339),
					"type":   "event",
					"method": "terminal.output",
					"params": json.RawMessage(paramsJSON),
				}
				data, err := json.Marshal(event)
				if err != nil {
					continue
				}
				if err := sender(data); err != nil {
					return
				}
			}

			if readErr != nil {
				if os.IsTimeout(readErr) {
					continue
				}
				return
			}
		}
	}()
}

// Wait blocks until the process exits.
func (s *Session) Wait() error {
	return s.cmd.Wait()
}

func (s *Session) startIdleChecker() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.mu.Lock()
				if s.closed {
					s.mu.Unlock()
					return
				}
				idle := time.Since(s.lastActivity)
				if idle >= s.inactivityTimeout {
					s.mu.Unlock()
					s.Close()
					return
				}
				s.mu.Unlock()
			}
		}
	}()
}

func (s *Session) touchActivity() {
	s.lastActivity = time.Now()
}
