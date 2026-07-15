package screen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"log/slog"
	"os/exec"
	"runtime"
	"time"
)

type Handler struct {
	capturer *Capturer
	log      *slog.Logger
}

func NewHandler(log *slog.Logger) *Handler {
	return &Handler{
		capturer: NewCapturer(),
		log:      log,
	}
}

type CaptureRequest struct {
	Width   int `json:"width"`
	Height  int `json:"height"`
	Quality int `json:"quality"`
}

type CaptureResponse struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Size     int    `json:"size"`
	Format   string `json:"format"`
	Captured string `json:"captured"`
	Data     []byte `json:"data"`
}

func (h *Handler) HandleCapture(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req CaptureRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}

	if req.Quality <= 0 {
		req.Quality = 80
	}

	opts := &CaptureOptions{
		Width:   req.Width,
		Height:  req.Height,
		Quality: req.Quality,
	}

	result, err := h.capturer.Capture(opts)
	if err != nil {
		h.log.Warn("screen capture failed", "error", err)
		return nil, fmt.Errorf("capture failed: %w", err)
	}

	var imgData []byte
	if req.Quality >= 90 {
		imgData, err = h.capturer.ToPNG(result.Image)
		if err != nil {
			return nil, fmt.Errorf("encode png: %w", err)
		}
	} else {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, result.Image, &jpeg.Options{Quality: req.Quality}); err != nil {
			return nil, fmt.Errorf("encode jpeg: %w", err)
		}
		imgData = buf.Bytes()
	}

	format := "jpeg"
	if req.Quality >= 90 {
		format = "png"
	}

	h.log.Info("screen captured",
		"width", result.Width,
		"height", result.Height,
		"size", len(imgData),
		"format", format,
	)

	return &CaptureResponse{
		Width:    result.Width,
		Height:   result.Height,
		Size:     len(imgData),
		Format:   format,
		Captured: result.Captured.Format(time.RFC3339),
		Data:     imgData,
	}, nil
}

// Input injection via platform-specific tools.
// On Linux: xdotool (X11) or ydotool (Wayland)
// On macOS: osascript (AppleScript)

type InputRequest struct {
	Type  string `json:"type"`            // "key", "mouse_move", "mouse_click", "mouse_down", "mouse_up", "scroll"
	Key   string `json:"key,omitempty"`   // key name for "key" type, e.g. "Return", "a", "ctrl+c"
	X     int    `json:"x,omitempty"`     // x coordinate for mouse events
	Y     int    `json:"y,omitempty"`     // y coordinate for mouse events
	Button int   `json:"button,omitempty"` // mouse button: 1=left, 2=middle, 3=right
	Delta int    `json:"delta,omitempty"`  // scroll delta (positive=up, negative=down)
}

type InputResponse struct {
	Ok     bool   `json:"ok"`
	Method string `json:"method"`
}

func (h *Handler) HandleInput(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req InputRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	method, err := h.sendInput(req)
	if err != nil {
		h.log.Warn("input injection failed", "type", req.Type, "error", err)
		return nil, fmt.Errorf("input failed: %w", err)
	}

	h.log.Info("input sent", "type", req.Type, "method", method)

	return &InputResponse{
		Ok:     true,
		Method: method,
	}, nil
}

func (h *Handler) sendInput(req InputRequest) (string, error) {
	switch runtime.GOOS {
	case "linux":
		return h.sendInputLinux(req)
	case "darwin":
		return h.sendInputDarwin(req)
	default:
		return "", fmt.Errorf("input not supported on %s", runtime.GOOS)
	}
}

func (h *Handler) sendInputLinux(req InputRequest) (string, error) {
	// Try xdotool first (X11), then ydotool (Wayland)
	for _, tool := range []string{"xdotool", "ydotool"} {
		if _, err := exec.LookPath(tool); err != nil {
			continue
		}
		cmd := buildLinuxInputCmd(tool, req)
		if cmd != nil {
			if err := cmd.Run(); err != nil {
				h.log.Warn(tool+" failed", "error", err)
				continue
			}
			return tool, nil
		}
	}
	return "", fmt.Errorf("no input tool available (xdotool/ydotool)")
}

func buildLinuxInputCmd(tool string, req InputRequest) *exec.Cmd {
	switch req.Type {
	case "key":
		if tool == "xdotool" {
			return exec.Command("xdotool", "key", req.Key)
		}
		return exec.Command("ydotool", "key", req.Key)
	case "mouse_move":
		if tool == "xdotool" {
			return exec.Command("xdotool", "mousemove", fmt.Sprintf("%d", req.X), fmt.Sprintf("%d", req.Y))
		}
		return exec.Command("ydotool", "mousemove", "--absolute", fmt.Sprintf("%d", req.X), fmt.Sprintf("%d", req.Y))
	case "mouse_click":
		btn := "1"
		if req.Button == 2 {
			btn = "2"
		} else if req.Button == 3 {
			btn = "3"
		}
		if tool == "xdotool" {
			return exec.Command("xdotool", "click", btn)
		}
		return exec.Command("ydotool", "click", btn)
	case "scroll":
		if tool == "xdotool" {
			dir := "5" // button 5 = scroll up
			if req.Delta < 0 {
				dir = "4" // button 4 = scroll down
			}
			return exec.Command("xdotool", "click", dir)
		}
		return exec.Command("ydotool", "mousemove", "--vertical", fmt.Sprintf("%d", req.Delta))
	default:
		return nil
	}
}

func (h *Handler) sendInputDarwin(req InputRequest) (string, error) {
	switch req.Type {
	case "key":
		// Convert key names to AppleScript key codes
		keyCode := macKeyCode(req.Key)
		script := fmt.Sprintf(`tell application "System Events" to keystroke "%s" using key code %d`, req.Key, keyCode)
		return "osascript", exec.Command("osascript", "-e", script).Run()
	case "mouse_move", "mouse_click":
		script := fmt.Sprintf(`tell application "System Events" to click at {%d, %d}`, req.X, req.Y)
		return "osascript", exec.Command("osascript", "-e", script).Run()
	case "scroll":
		dir := "up"
		if req.Delta < 0 {
			dir = "down"
		}
		script := fmt.Sprintf(`tell application "System Events" to scroll area {0, 0} direction %q`, dir)
		return "osascript", exec.Command("osascript", "-e", script).Run()
	default:
		return "", fmt.Errorf("unsupported input type on macOS: %s", req.Type)
	}
}

func macKeyCode(key string) int {
	// Common key mappings
	switch key {
	case "Return", "return", "enter":
		return 36
	case "space":
		return 49
	case "tab":
		return 48
	case "escape":
		return 53
	case "delete", "backspace":
		return 51
	case "a":
		return 0
	case "c":
		return 8
	case "v":
		return 9
	case "x":
		return 7
	default:
		return 0
	}
}

type StreamRequest struct {
	Width   int `json:"width"`
	Height  int `json:"height"`
	FPS     int `json:"fps"`
	Quality int `json:"quality"`
}

type StreamFrame struct {
	Frame   int    `json:"frame"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Size    int    `json:"size"`
	Format  string `json:"format"`
	Data    []byte `json:"data"`
}

func (h *Handler) StartStream(ctx context.Context, req StreamRequest, onFrame func(StreamFrame) error) error {
	if req.FPS <= 0 {
		req.FPS = 1
	}
	if req.FPS > 30 {
		req.FPS = 30
	}
	if req.Quality <= 0 {
		req.Quality = 70
	}

	interval := time.Second / time.Duration(req.FPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	frame := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			opts := &CaptureOptions{
				Width:  req.Width,
				Height: req.Height,
			}

			result, err := h.capturer.Capture(opts)
			if err != nil {
				h.log.Warn("stream capture failed", "frame", frame, "error", err)
				continue
			}

			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, result.Image, &jpeg.Options{Quality: req.Quality}); err != nil {
				h.log.Warn("stream encode failed", "frame", frame, "error", err)
				continue
			}

			frame++
			if err := onFrame(StreamFrame{
				Frame:  frame,
				Width:  result.Width,
				Height: result.Height,
				Size:   buf.Len(),
				Format: "jpeg",
				Data:   buf.Bytes(),
			}); err != nil {
				return err
			}
		}
	}
}
