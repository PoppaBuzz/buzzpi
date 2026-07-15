package file

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestHandleBrowse(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	h := NewHandler(testLogger())
	params, _ := json.Marshal(BrowseRequest{Path: dir})

	result, err := h.HandleBrowse(context.Background(), params)
	if err != nil {
		t.Fatalf("HandleBrowse failed: %v", err)
	}

	resp := result.(*BrowseResponse)
	if len(resp.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(resp.Files))
	}

	foundTxt := false
	foundDir := false
	for _, f := range resp.Files {
		if f.Name == "test.txt" && !f.Directory {
			foundTxt = true
		}
		if f.Name == "subdir" && f.Directory {
			foundDir = true
		}
	}
	if !foundTxt {
		t.Error("test.txt not found")
	}
	if !foundDir {
		t.Error("subdir not found")
	}
}

func TestHandleUpload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "uploaded.txt")

	h := NewHandler(testLogger())
	params, _ := json.Marshal(UploadRequest{
		Path: path,
		Data: "aGVsbG8=", // base64("hello")
	})

	result, err := h.HandleUpload(context.Background(), params)
	if err != nil {
		t.Fatalf("HandleUpload failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(data))
	}

	_ = result
}

func TestHandleDownload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "download.txt")
	os.WriteFile(path, []byte("world"), 0644)

	h := NewHandler(testLogger())
	params, _ := json.Marshal(DownloadRequest{Path: path})

	result, err := h.HandleDownload(context.Background(), params)
	if err != nil {
		t.Fatalf("HandleDownload failed: %v", err)
	}

	resp := result.(*DownloadResponse)
	if resp.Size != 5 {
		t.Errorf("expected size 5, got %d", resp.Size)
	}
	if resp.Data != "d29ybGQ=" { // base64("world")
		t.Errorf("unexpected data: %s", resp.Data)
	}
}

func TestHandleDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "to_delete.txt")
	os.WriteFile(path, []byte("delete me"), 0644)

	h := NewHandler(testLogger())
	params, _ := json.Marshal(DeleteRequest{Path: path})

	_, err := h.HandleDelete(context.Background(), params)
	if err != nil {
		t.Fatalf("HandleDelete failed: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should have been deleted")
	}
}

func TestHandleMkdir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new_dir", "nested")

	h := NewHandler(testLogger())
	params, _ := json.Marshal(MkdirRequest{Path: path})

	_, err := h.HandleMkdir(context.Background(), params)
	if err != nil {
		t.Fatalf("HandleMkdir failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestDetectMIME(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"test.html", "text/html"},
		{"style.css", "text/css"},
		{"app.js", "application/javascript"},
		{"data.json", "application/json"},
		{"image.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"doc.pdf", "application/pdf"},
		{"archive.zip", "application/zip"},
		{"readme.txt", "text/plain"},
		{"unknown.xyz", "application/octet-stream"},
	}

	for _, tt := range tests {
		result := detectMIME(tt.path)
		if result != tt.expected {
			t.Errorf("detectMIME(%s): expected %s, got %s", tt.path, tt.expected, result)
		}
	}
}
