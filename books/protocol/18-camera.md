# BPP Chapter 18: Camera Service

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The Camera service provides access to cameras connected to the device (Raspberry Pi Camera Module, USB cameras).

## Overview

This service allows clients to:
- View a live video stream from the camera
- Capture still images (snapshots)
- Record video clips
- Control camera parameters (resolution, exposure, white balance)

## Methods

### camera.list

List available cameras.

**Request:**
```json
{
  "method": "camera.list",
  "params": {}
}
```

**Response:**
```json
{
  "method": "camera.list",
  "result": {
    "cameras": [
      {
        "id": "cam0",
        "name": "Camera Module 3",
        "model": "IMX708",
        "connection": "csi",
        "max_resolution": { "width": 4608, "height": 2592 },
        "supported_resolutions": [
          { "width": 1920, "height": 1080 },
          { "width": 1280, "height": 720 },
          { "width": 640, "height": 480 }
        ],
        "supports_preview": true,
        "supports_recording": true,
        "supports_controls": true,
        "is_available": true
      }
    ]
  }
}
```

### camera.preview.start

Start a live camera preview stream.

**Request:**
```json
{
  "method": "camera.preview.start",
  "params": {
    "camera_id": "cam0",
    "resolution": { "width": 1920, "height": 1080 },
    "fps": 30,
    "bitrate_kbps": 5000,
    "rotation": 0,
    "hflip": false,
    "vflip": false
  }
}
```

The preview stream is delivered as an H.264 video track over WebRTC (same mechanism as screen streaming). The client renders it in a video view.

**Response:**
```json
{
  "method": "camera.preview.start",
  "result": {
    "camera_id": "cam0",
    "session_id": "cam_preview_abc123",
    "actual_resolution": { "width": 1920, "height": 1080 },
    "actual_fps": 30,
    "codec": "H.264"
  }
}
```

### camera.preview.stop

Stop the preview stream.

**Request:**
```json
{
  "method": "camera.preview.stop",
  "params": {
    "camera_id": "cam0",
    "session_id": "cam_preview_abc123"
  }
}
```

### camera.snapshot

Capture a still image.

**Request:**
```json
{
  "method": "camera.snapshot",
  "params": {
    "camera_id": "cam0",
    "resolution": { "width": 4608, "height": 2592 },
    "format": "jpeg",
    "quality": 95
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `format` | `jpeg` | `jpeg`, `png`, `bmp` |
| `quality` | 95 | JPEG quality 1-100 |

**Response:**
```json
{
  "method": "camera.snapshot",
  "result": {
    "camera_id": "cam0",
    "format": "jpeg",
    "size_bytes": 2450000,
    "resolution": { "width": 4608, "height": 2592 },
    "timestamp": "2026-07-07T12:00:00Z",
    "data": "<base64_encoded_image_data>",
    "encoding": "base64"
  }
}
```

For large images, the data is chunked (same protocol as file transfers).

### camera.record.start / stop

Start and stop recording a video clip.

**Request:**
```json
{
  "method": "camera.record.start",
  "params": {
    "camera_id": "cam0",
    "resolution": { "width": 1920, "height": 1080 },
    "fps": 30,
    "max_duration_seconds": 300,
    "max_size_mb": 500,
    "segment_duration_seconds": 60
  }
}
```

**Response:**
```json
{
  "method": "camera.record.start",
  "result": {
    "session_id": "cam_rec_abc123",
    "status": "recording",
    "output_path": "/home/pi/Videos/buzzpi_rec_20260707_120000.mp4"
  }
}
```

**Request (stop):**
```json
{
  "method": "camera.record.stop",
  "params": {
    "camera_id": "cam0",
    "session_id": "cam_rec_abc123"
  }
}
```

**Response:**
```json
{
  "method": "camera.record.stop",
  "result": {
    "session_id": "cam_rec_abc123",
    "status": "completed",
    "output_path": "/home/pi/Videos/buzzpi_rec_20260707_120000.mp4",
    "duration_seconds": 45,
    "size_mb": 180,
    "format": "mp4"
  }
}
```

### camera.controls

Get or set camera controls.

**Get controls:**
```json
{
  "method": "camera.controls",
  "params": {
    "camera_id": "cam0"
  }
}
```

**Response:**
```json
{
  "method": "camera.controls",
  "result": {
    "controls": {
      "brightness": { "value": 0, "min": -100, "max": 100, "step": 1, "default": 0 },
      "contrast": { "value": 100, "min": 0, "max": 200, "step": 1, "default": 100 },
      "saturation": { "value": 100, "min": 0, "max": 200, "step": 1, "default": 100 },
      "exposure_mode": { "value": "auto", "options": ["auto", "night", "backlight", "spotlight"] },
      "white_balance": { "value": "auto", "options": ["auto", "incandescent", "fluorescent", "daylight"] },
      "iso": { "value": 0, "min": 100, "max": 3200, "step": 100, "default": 0 },
      "sharpness": { "value": 0, "min": -100, "max": 100, "step": 1, "default": 0 }
    }
  }
}
```

**Set controls:**
```json
{
  "method": "camera.controls",
  "params": {
    "camera_id": "cam0",
    "controls": {
      "brightness": 10,
      "exposure_mode": "night"
    }
  }
}
```

## Hardware Support

| Platform | Camera Support | Method |
|----------|---------------|--------|
| Raspberry Pi (libcamera) | Full | libcamera-vid, libcamera-still |
| Raspberry Pi (legacy) | Full | raspivid, raspistill |
| USB camera (V4L2) | Partial | Video4Linux2 |
| Other CSI cameras | Partial | Platform-specific |

The Runtime detects available camera backends and uses the best available.

## Security

| Concern | Mitigation |
|---------|------------|
| Unauthorized camera access | Camera requires explicit extension permission; visual indicator when active |
| Privacy (remote spying) | Camera is disabled by default; user must enable per-device in settings |
| Recording without consent | Active recording is indicated by the device's camera LED (hardware) |
| Stream interception | End-to-end WebRTC encryption |

## Rate Limits

| Operation | Limit |
|-----------|-------|
| Concurrent preview streams | 1 per camera |
| Snapshot requests | 10 per minute |
| Recording sessions | 1 per camera |
| Control changes | 10 per second |
