package screen

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"log/slog"
	"os"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestCaptureOptions(t *testing.T) {
	opts := &CaptureOptions{
		Width:  640,
		Height: 480,
		Quality: 80,
	}

	if opts.Width != 640 {
		t.Errorf("expected width 640, got %d", opts.Width)
	}
	if opts.Height != 480 {
		t.Errorf("expected height 480, got %d", opts.Height)
	}
}

func TestResizeImage(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	for y := 0; y < 1080; y++ {
		for x := 0; x < 1920; x++ {
			src.SetRGBA(x, y, color.RGBA{R: 255, G: 128, B: 64, A: 255})
		}
	}

	dst := resizeImage(src, 640, 360)
	rgba, ok := dst.(*image.RGBA)
	if !ok {
		t.Fatal("resizeImage did not return *image.RGBA")
	}

	bounds := rgba.Bounds()
	if bounds.Dx() != 640 {
		t.Errorf("expected width 640, got %d", bounds.Dx())
	}
	if bounds.Dy() != 360 {
		t.Errorf("expected height 360, got %d", bounds.Dy())
	}

	if rgba.RGBAAt(0, 0).R != 255 {
		t.Errorf("expected red 255, got %d", rgba.RGBAAt(0, 0).R)
	}
}

func TestResizeImageScaleWidthOnly(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1920, 1080))

	dst := resizeImage(src, 640, 0)

	bounds := dst.Bounds()
	if bounds.Dx() != 640 {
		t.Errorf("expected width 640, got %d", bounds.Dx())
	}
	expectedHeight := 1080 * 640 / 1920
	if bounds.Dy() != expectedHeight {
		t.Errorf("expected height %d, got %d", expectedHeight, bounds.Dy())
	}
}

func TestResizeImageNoResize(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1920, 1080))

	dst := resizeImage(src, 0, 0)

	bounds := dst.Bounds()
	if bounds.Dx() != 1920 {
		t.Errorf("expected width 1920, got %d", bounds.Dx())
	}
	if bounds.Dy() != 1080 {
		t.Errorf("expected height 1080, got %d", bounds.Dy())
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"line1\nline2\nline3", []string{"line1", "line2", "line3"}},
		{"single", []string{"single"}},
		{"", nil},
		{"line1\n\nline3", []string{"line1", "", "line3"}},
	}

	for _, tt := range tests {
		result := splitLines(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitLines(%q): expected %d lines, got %d", tt.input, len(tt.expected), len(result))
			continue
		}
		for i, line := range result {
			if line != tt.expected[i] {
				t.Errorf("splitLines(%q)[%d]: expected %q, got %q", tt.input, i, tt.expected[i], line)
			}
		}
	}
}

func TestParseFBSet(t *testing.T) {
	output := []byte(`framebuffer device
  id:          fb0
  flags:       fb
  geometry:    1920 1080 32 32
  timings:     ...
  bpp:         32
  ...`)

	var width, height, bpp int
	parseFBSet(output, &width, &height, &bpp)

	if width != 1920 {
		t.Errorf("expected width 1920, got %d", width)
	}
	if height != 1080 {
		t.Errorf("expected height 1080, got %d", height)
	}
	if bpp != 32 {
		t.Errorf("expected bpp 32, got %d", bpp)
	}
}

func TestParseFBSetMinimal(t *testing.T) {
	output := []byte(`geometry 800 600 16 16
bpp 16`)

	var width, height, bpp int
	parseFBSet(output, &width, &height, &bpp)

	if width != 800 {
		t.Errorf("expected width 800, got %d", width)
	}
	if height != 600 {
		t.Errorf("expected height 600, got %d", height)
	}
	if bpp != 16 {
		t.Errorf("expected bpp 16, got %d", bpp)
	}
}

func TestCapturerToPNG(t *testing.T) {
	c := NewCapturer()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	data, err := c.ToPNG(img)
	if err != nil {
		t.Fatalf("ToPNG failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ToPNG returned empty data")
	}

	decoded, err := png.Decode(createTempFile(t, data))
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("decoded image dimensions: expected 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestHandlerCaptureUnsupportedPlatform(t *testing.T) {
	h := NewHandler(testLogger())

	req := CaptureRequest{
		Width:   640,
		Height:  480,
		Quality: 80,
	}

	params, _ := json.Marshal(req)

	_, err := h.HandleCapture(context.Background(), params)

	if err == nil {
		t.Skip("capture may work on this platform")
	}
}

func TestHandlerCaptureEmptyParams(t *testing.T) {
	h := NewHandler(testLogger())

	_, err := h.HandleCapture(context.Background(), nil)

	if err == nil {
		t.Skip("capture may work on this platform with default params")
	}
}

func createTempFile(t *testing.T, data []byte) *os.File {
	t.Helper()

	f, err := os.CreateTemp("", "test-*.png")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		f.Close()
		os.Remove(f.Name())
	})

	if _, err := f.Write(data); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("failed to seek temp file: %v", err)
	}

	return f
}
