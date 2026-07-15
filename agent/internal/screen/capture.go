package screen

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Capturer struct {
	mu       sync.Mutex
	lastCapt time.Time
}

func NewCapturer() *Capturer {
	return &Capturer{}
}

type CaptureOptions struct {
	Width  int
	Height int
	Quality int
}

type CaptureResult struct {
	Image    image.Image
	Width    int
	Height   int
	Captured time.Time
}

func (c *Capturer) Capture(opts *CaptureOptions) (*CaptureResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var img image.Image
	var err error

	switch runtime.GOOS {
	case "linux":
		img, err = c.captureLinux()
	case "darwin":
		img, err = c.captureDarwin()
	default:
		return nil, fmt.Errorf("screen capture not supported on %s", runtime.GOOS)
	}

	if err != nil {
		return nil, err
	}

	if opts != nil && (opts.Width > 0 || opts.Height > 0) {
		img = resizeImage(img, opts.Width, opts.Height)
	}

	bounds := img.Bounds()
	c.lastCapt = time.Now()

	return &CaptureResult{
		Image:    img,
		Width:    bounds.Dx(),
		Height:   bounds.Dy(),
		Captured: c.lastCapt,
	}, nil
}

func (c *Capturer) captureLinux() (image.Image, error) {
	if img, err := c.captureFramebuffer(); err == nil {
		return img, nil
	}

	if img, err := c.captureX11(); err == nil {
		return img, nil
	}

	if img, err := c.captureWayland(); err == nil {
		return img, nil
	}

	return nil, fmt.Errorf("no screen capture method available")
}

func (c *Capturer) captureDarwin() (image.Image, error) {
	tmpFile, err := os.CreateTemp("", "buzzpi-screenshot-*.png")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := exec.Command("screencapture", "-x", "-C", tmpFile.Name())
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("screencapture failed: %w", err)
	}

	return loadPNG(tmpFile.Name())
}

func (c *Capturer) captureFramebuffer() (image.Image, error) {
	f, err := os.Open("/dev/fb0")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	var fbWidth, fbHeight, bpp int
	if out, err := exec.Command("fbset", "-s").Output(); err == nil {
	 parseFBSet(out, &fbWidth, &fbHeight, &bpp)
	}

	if fbWidth == 0 || fbHeight == 0 {
		fbWidth = 1920
		fbHeight = 1080
		bpp = 32
	}

	data := make([]byte, info.Size())
	if _, err := f.Read(data); err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, fbWidth, fbHeight))
	pixelSize := bpp / 8
	for y := 0; y < fbHeight; y++ {
		for x := 0; x < fbWidth; x++ {
			offset := (y*fbWidth + x) * pixelSize
			if offset+3 >= len(data) {
				continue
			}
			img.SetRGBA(x, y, color.RGBA{
				R: data[offset+2],
				G: data[offset+1],
				B: data[offset],
				A: 255,
			})
		}
	}

	return img, nil
}

func parseFBSet(output []byte, width, height, bpp *int) {
	text := string(output)
	for _, line := range splitLines(text) {
		line = strings.TrimSpace(line)
		var w, h, b int
		if _, err := fmt.Sscanf(line, "geometry %d %d", &w, &h); err == nil && w > 0 {
			*width = w
			*height = h
		}
		if _, err := fmt.Sscanf(line, "geometry: %d %d", &w, &h); err == nil && w > 0 {
			*width = w
			*height = h
		}
		if _, err := fmt.Sscanf(line, "bpp %d", &b); err == nil && b > 0 {
			*bpp = b
		}
		if _, err := fmt.Sscanf(line, "bpp: %d", &b); err == nil && b > 0 {
			*bpp = b
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func (c *Capturer) captureX11() (image.Image, error) {
	tmpFile, err := os.CreateTemp("", "buzzpi-screenshot-*.png")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	for _, tool := range []string{"scrot", "import", "gnome-screenshot"} {
		cmd := exec.Command(tool, "-o", tmpFile.Name())
		if err := cmd.Run(); err == nil {
			return loadPNG(tmpFile.Name())
		}
	}

	return nil, fmt.Errorf("no X11 capture tool available")
}

func (c *Capturer) captureWayland() (image.Image, error) {
	tmpFile, err := os.CreateTemp("", "buzzpi-screenshot-*.png")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cmd := exec.Command("grim", tmpFile.Name())
	if err := cmd.Run(); err == nil {
		return loadPNG(tmpFile.Name())
	}

	return nil, fmt.Errorf("grim not available for Wayland capture")
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func resizeImage(img image.Image, targetW, targetH int) image.Image {
	if targetW <= 0 && targetH <= 0 {
		return img
	}

	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if targetW <= 0 {
		targetW = srcW * targetH / srcH
	}
	if targetH <= 0 {
		targetH = srcH * targetW / srcW
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	for y := 0; y < targetH; y++ {
		for x := 0; x < targetW; x++ {
			srcX := bounds.Min.X + x*srcW/targetW
			srcY := bounds.Min.Y + y*srcH/targetH
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}

	return dst
}

func (c *Capturer) ToPNG(img image.Image) ([]byte, error) {
	var buf []byte
	err := func() error {
		f, err := os.CreateTemp("", "buzzpi-*.png")
		if err != nil {
			return err
		}
		defer func() {
			f.Close()
			data, readErr := os.ReadFile(f.Name())
			if readErr == nil {
				buf = data
			}
			os.Remove(f.Name())
		}()
		return png.Encode(f, img)
	}()
	return buf, err
}
