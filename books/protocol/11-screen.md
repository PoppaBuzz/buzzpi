# BPP Chapter 11: Screen Service

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The Screen service provides remote graphical desktop access to a device. It captures the device's display output, encodes it as H.264 video, and streams it to the client over a WebRTC media track.

## Overview

The Screen service is the most technically demanding BPP service. It requires:
- Hardware-accelerated video encoding (VideoCore on Raspberry Pi, V4L2 on other platforms)
- Real-time frame capture from the device's framebuffer or display server
- Adaptive bitrate and resolution based on network conditions
- Input injection (mouse, keyboard, touch) relayed back to the device

## Methods

### screen.start

Start screen streaming.

**Request:**
```json
{
  "method": "screen.start",
  "params": {
    "quality": "adaptive",
    "max_fps": 30,
    "max_width": 1920,
    "max_height": 1080,
    "capture_mode": "auto",
    "show_cursor": true
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `quality` | `adaptive` | `adaptive`, `high`, `medium`, `low` |
| `max_fps` | 30 | Maximum frames per second (5-30) |
| `max_width` | 1920 | Maximum resolution width |
| `max_height` | 1080 | Maximum resolution height |
| `capture_mode` | `auto` | `auto`, `drm`, `fbdev`, `x11`, `wayland`, `vnc` |
| `show_cursor` | true | Whether to include cursor in capture |

**Response:**
```json
{
  "method": "screen.start",
  "result": {
    "session_id": "scrn_abc123",
    "actual_width": 1920,
    "actual_height": 1080,
    "capture_mode": "drm",
    "codec": "H.264",
    "pixel_format": "NV12"
  }
}
```

### screen.stop

Stop screen streaming.

**Request:**
```json
{
  "method": "screen.stop",
  "params": {
    "session_id": "scrn_abc123"
  }
}
```

### screen.input

Send input events to the device.

**Request:**
```json
{
  "method": "screen.input",
  "params": {
    "session_id": "scrn_abc123",
    "events": [
      {
        "type": "mousemove",
        "x": 800,
        "y": 540,
        "timestamp": 1234567890
      },
      {
        "type": "mousedown",
        "button": "left",
        "x": 800,
        "y": 540,
        "timestamp": 1234567891
      },
      {
        "type": "mouseup",
        "button": "left",
        "x": 800,
        "y": 540,
        "timestamp": 1234567892
      }
    ]
  }
}
```

Input events are batched (up to 10 events per message) for efficiency. Each event has a relative timestamp for accurate replay.

### Input Event Types

| Type | Parameters | Description |
|------|------------|-------------|
| `mousemove` | `x`, `y` | Move mouse to absolute position |
| `mousedown` | `x`, `y`, `button` | Mouse button pressed |
| `mouseup` | `x`, `y`, `button` | Mouse button released |
| `mousescroll` | `x`, `y`, `delta_x`, `delta_y` | Scroll wheel |
| `keydown` | `key`, `code`, `modifiers` | Key pressed |
| `keyup` | `key`, `code`, `modifiers` | Key released |
| `touchstart` | `touches[]` | Touch started |
| `touchmove` | `touches[]` | Touch moved |
| `touchend` | `touches[]` | Touch ended |

### Input Coordinate Mapping

The client sends input coordinates relative to the video frame dimensions. The device maps coordinates to its actual display resolution:

```
device_x = (client_x / frame_width) * device_display_width
device_y = (client_y / frame_height) * device_display_height
```

## Capture Methods

### DRM/KMS (Preferred)

Uses the Direct Rendering Manager / Kernel Mode Setting interface:

| Platform | Support | Quality |
|----------|---------|---------|
| Raspberry Pi (all models) | Full via VC4/V3D driver | Highest |
| Other Linux with DRM | Full | High |

Advantages:
- Captures the actual framebuffer (boot console, login screen, desktop)
- No display server dependency (works without X11 or Wayland)
- Low overhead (direct kernel interface)
- Supports hardware encoding pipeline

### fbdev (Fallback)

Uses the Linux framebuffer device (`/dev/fb0`):

| Platform | Support | Quality |
|----------|---------|---------|
| Raspberry Pi | Available | Medium |
| Other Linux | Available | Medium |

Limitations:
- No cursor overlay (cursor must be rendered into framebuffer)
- No damage tracking (full frames only, higher bandwidth)
- Limited resolution support

### X11/Wayland (Desktop Fallback)

For devices running a full desktop environment:

Captures via:
- `XGetImage` / `xcb_composite` (X11)
- `wl_subsurface` / `pipewire` (Wayland)

### VNC (Compatibility Fallback)

If no native capture method works, the Runtime can start a VNC server and capture from it. This is the least preferred method (adds latency, requires VNC client on device).

## Video Pipeline

```
 ┌──────────┐    ┌────────────┐    ┌───────────┐    ┌──────────┐
 │          │    │            │    │           │    │          │
 │ Capture  │───▶│ Color      │───▶│ H.264     │───▶│ WebRTC   │
 │ (DRM)    │    │ Convert    │    │ Encoder   │    │ Track    │
 │          │    │ (NV12)     │    │ (Video4)  │    │          │
 └──────────┘    └────────────┘    └───────────┘    └──────────┘
      │                                                │
      │                                            RTP/RTX
      │                                               │
   Damage                                       DTLS-SRTP
   tracking                                         │
      │                                               v
      │                                          Client App
      └────── Frame comparison → skip unchanged ────→ (decoder)
```

### Damage Tracking

To reduce bandwidth, the device tracks which regions of the screen have changed:

- Frame divided into 64x64 pixel tiles
- Only tiles that changed since the last frame are encoded
- Full frame sent as keyframe every 2 seconds (or on request)
- Subtle changes (clock updates, cursor movement) may not trigger full tile updates

## Adaptive Quality

The device monitors WebRTC statistics and adjusts encoding:

### Quality Profiles

| Profile | Max Resolution | Max FPS | Max Bitrate | Use Case |
|---------|---------------|---------|-------------|----------|
| `high` | 1920x1080 | 30 | 5 Mbps | LAN, strong WiFi |
| `medium` | 1280x720 | 15 | 2 Mbps | Typical remote access |
| `low` | 854x480 | 10 | 1 Mbps | Cellular, weak connection |
| `minimum` | 640x360 | 5 | 500 Kbps | Edge connectivity |

### Adaptation Logic

```
Monitor WebRTC stats every 5 seconds:

IF packet_loss > 5% for 2 consecutive intervals:
    Move to next lower profile

IF packet_loss < 1% for 5 consecutive intervals:
    Move to next higher profile

IF round_trip_time > 500ms:
    Reduce FPS by 5 (minimum 5)

IF available_bitrate < current_bitrate * 1.2:
    Reduce profile
```

## Client Rendering

The client renders the video stream using the platform's hardware-accelerated video decoder:

| Platform | Decoder |
|----------|---------|
| Android | MediaCodec (H.264 hardware decoder) |

The video view overlays:
- Touch/mouse input layer (transparent, captures gestures)
- Status bar (connection quality, FPS, resolution — optional, debug)
- Toolbar (keyboard toggle, home button, orientation lock)

## Security

| Concern | Mitigation |
|---------|------------|
| Unauthorized screen access | Authenticated + encrypted WebRTC channel |
| Screen recording without consent | Visual indicator in device's system tray when screen is being streamed |
| Input injection | Only authorized clients can send input (mutual auth) |
| Keystroke capture (middleware attack) | End-to-end WebRTC encryption prevents relay from seeing input |
