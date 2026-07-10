# BPP Chapter 8: Compression

**Layer:** Transport  
**Status:** Draft  
**Version:** 1.0.0

BPP supports optional compression for data channel payloads to reduce bandwidth usage, particularly for terminal output and file transfers.

## Compression Methods

### Supported Methods

| Method | Identifier | Use Case |
|--------|------------|----------|
| None | `none` | Control messages, small payloads |
| gzip | `gzip` | Terminal output, log streaming |
| zstd | `zstd` | File transfer, large payloads |
| Snappy | `snappy` | Real-time data, low-latency preference |

### Selection

Compression is negotiated per-data-channel during connection setup:

```json
{
  "type": "session.capabilities",
  "params": {
    "compression": {
      "available": ["none", "gzip", "zstd"],
      "preferred": "zstd"
    }
  }
}
```

The sender chooses the compression method based on:
- What the receiver supports (from capability negotiation)
- Payload size (gzip for text, zstd for binary, none for messages under 1KB)
- Latency requirements (Snappy for real-time, zstd for throughput)

### Default Behavior

| Channel | Default Compression | Threshold |
|---------|-------------------|-----------|
| `terminal` | gzip | Payloads > 4KB |
| `services` | none | Always |
| `files` | zstd | Payloads > 1KB |
| `events` | none | Always |
| `input` | none | Always |
| `extension.*` | Channel-defined | Channel-defined |

## Framing

Compressed payloads use a flag byte before the message length:

```
┌────┬────────────────────────────────────────┐
│ F  │ Length                                 │
│ 1  │ 4 bytes (big-endian)                   │
│ byte│                                       │
├────┴────────────────────────────────────────┤
│ Payload (compressed or uncompressed)        │
└─────────────────────────────────────────────┘

F (flags byte):
  Bit 0-3: Compression method
    0000 = none
    0001 = gzip
    0010 = zstd
    0011 = snappy
  Bit 4-7: Reserved (set to 0)
```

## Codec Negotiation (Video)

Screen streaming uses H.264 video. Codec negotiation follows standard WebRTC SDP negotiation:

### Offered Profiles

The device (sender) offers these H.264 profiles in SDP:

```
a=rtpmap:100 H264/90000
a=fmtp:100 profile-level-id=42e01f;packetization-mode=1
a=rtpmap:101 H264/90000
a=fmtp:101 profile-level-id=4de01f;packetization-mode=1
```

| Profile | Profile-Level-ID | Quality |
|---------|-----------------|---------|
| Constrained Baseline | 42e01f | Default, widest compatibility |
| Main | 4de01f | Higher quality, better compression |

### Resolution Negotiation

The client (receiver) signals its preferred resolution in the SDP:

```
a=fmtp:100 max-fs=3600;max-fr=30
```

- `max-fs`: Maximum frame size (macroblocks). 3600 = 720p, 8160 = 1080p
- `max-fr`: Maximum frame rate

The device respects these limits and encodes at or below the requested resolution.

## Adaptive Bitrate

Video bitrate adapts to network conditions (see Chapter 6). The compression chapter covers the encoding-level details:

### Rate Control

```
Mode: CBR (Constant Bitrate) for screen streaming
      VBR (Variable Bitrate) for file transfers

Screen streaming CBR parameters:
  Initial bitrate: 2 Mbps
  Minimum: 500 Kbps
  Maximum: 5 Mbps
  Adjustment interval: 5 seconds
  Step size: 20% of current bitrate
```

### Keyframe Interval

```
Screen streaming:
  Keyframe interval (GOP): 2 seconds (60 frames at 30fps)
  Forced keyframe on: first frame, after connection loss, after 5%+ scene change
  Keyframe request: client sends PLI (Picture Loss Indication) via RTCP
```

## Efficiency Targets

| Channel | Uncompressed | Compressed | Target |
|---------|-------------|------------|--------|
| Terminal (text output) | 1 MB | 50 KB | 95% reduction |
| Terminal (ls -la, 50 files) | 3 KB | 1 KB | 66% reduction |
| File (text file, 100KB) | 100 KB | 30 KB | 70% reduction |
| File (binary, 1MB) | 1 MB | 1 MB (no gain) | zstd only if needed |
| Screen (720p frame) | 50 KB (keyframe) | 5 KB (delta) | 90% reduction |
| Events (status update) | 200 bytes | Not compressed | Insufficient gain |

Compression is skipped automatically when the payload is below the threshold or when compression would increase size (common for small or incompressible data).
