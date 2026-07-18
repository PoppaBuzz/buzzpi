package file

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Handler struct {
	log *slog.Logger
}

func NewHandler(log *slog.Logger) *Handler {
	return &Handler{log: log}
}

type BrowseRequest struct {
	Path string `json:"path"`
}

type FileInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Directory bool   `json:"is_directory"`
	Size      int64  `json:"size"`
	Modified  string `json:"modified"`
}

type BrowseResponse struct {
	Files []FileInfo `json:"files"`
	Path  string     `json:"path"`
}

func (h *Handler) HandleBrowse(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req BrowseRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}

	if req.Path == "" {
		req.Path = "/home"
	}

	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	// If /home is empty, fall back to /home/pi (common on Raspberry Pi)
	if len(entries) == 0 && absPath == "/home" {
		fallback := "/home/pi"
		if fi, err := os.Stat(fallback); err == nil && fi.IsDir() {
			absPath = fallback
			entries, err = os.ReadDir(absPath)
			if err != nil {
				return nil, fmt.Errorf("read directory: %w", err)
			}
		}
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(absPath, entry.Name())
		files = append(files, FileInfo{
			Name:      entry.Name(),
			Path:      fullPath,
			Directory: entry.IsDir(),
			Size:      info.Size(),
			Modified:  info.ModTime().Format(time.RFC3339),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Directory != files[j].Directory {
			return files[i].Directory
		}
		return files[i].Name < files[j].Name
	})

	h.log.Info("directory browsed", "path", absPath, "count", len(files))

	return &BrowseResponse{
		Files: files,
		Path:  absPath,
	}, nil
}

type UploadRequest struct {
	Path string `json:"path"`
	Data string `json:"data"`
}

func (h *Handler) HandleUpload(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req UploadRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}

	if req.Path == "" || req.Data == "" {
		return nil, fmt.Errorf("path and data are required")
	}

	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		return nil, fmt.Errorf("decode data: %w", err)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	h.log.Info("file uploaded", "path", absPath, "size", len(data))

	return map[string]interface{}{
		"path": absPath,
		"size": len(data),
	}, nil
}

type DownloadRequest struct {
	Path string `json:"path"`
}

type DownloadResponse struct {
	Path string `json:"path"`
	Data string `json:"data"`
	Size int64  `json:"size"`
	MIME string `json:"mime_type"`
}

func (h *Handler) HandleDownload(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req DownloadRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}

	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	mimeType := detectMIME(absPath)

	h.log.Info("file downloaded", "path", absPath, "size", len(data))

	return &DownloadResponse{
		Path: absPath,
		Data: base64.StdEncoding.EncodeToString(data),
		Size: int64(len(data)),
		MIME: mimeType,
	}, nil
}

type DeleteRequest struct {
	Path string `json:"path"`
}

func (h *Handler) HandleDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req DeleteRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}

	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	if err := os.RemoveAll(absPath); err != nil {
		return nil, fmt.Errorf("delete: %w", err)
	}

	h.log.Info("file deleted", "path", absPath)

	return map[string]interface{}{
		"path":    absPath,
		"deleted": true,
	}, nil
}

type RenameRequest struct {
	Path    string `json:"path"`
	NewName string `json:"new_name"`
}

func (h *Handler) HandleRename(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req RenameRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}

	if req.Path == "" || req.NewName == "" {
		return nil, fmt.Errorf("path and new_name are required")
	}

	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	parent := filepath.Dir(absPath)
	newPath := filepath.Join(parent, req.NewName)

	if err := os.Rename(absPath, newPath); err != nil {
		return nil, fmt.Errorf("rename: %w", err)
	}

	h.log.Info("file renamed", "old", absPath, "new", newPath)

	return map[string]interface{}{
		"old_path": absPath,
		"new_path": newPath,
		"renamed":  true,
	}, nil
}

type MkdirRequest struct {
	Path string `json:"path"`
}

func (h *Handler) HandleMkdir(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req MkdirRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}

	if req.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	h.log.Info("directory created", "path", absPath)

	return map[string]interface{}{
		"path":    absPath,
		"created": true,
	}, nil
}

func detectMIME(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".txt", ".log", ".md":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
