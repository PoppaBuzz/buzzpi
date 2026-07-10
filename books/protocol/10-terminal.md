# BPP Chapter 10: Terminal Service

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The Terminal service provides remote shell access to a device. It is one of the core services that every BPP implementation should support.

## Overview

The Terminal service creates a pseudo-terminal (PTY) on the device and exposes stdin/stdout over a WebRTC data channel. The client renders the terminal output and sends user input (keystrokes, paste events) to the device.

## Methods

### terminal.open

Create a new terminal session.

**Request:**
```json
{
  "method": "terminal.open",
  "params": {
    "shell": "/bin/bash",
    "cols": 80,
    "rows": 24,
    "term": "xterm-256color",
    "env": {
      "BUZZPI_SESSION": "true",
      "BUZZPI_DEVICE_ID": "<device_id>"
    }
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `shell` | `/bin/bash` | Shell to execute (MUST verify against allowed list) |
| `cols` | 80 | Initial terminal width (columns) |
| `rows` | 24 | Initial terminal height (rows) |
| `term` | `xterm-256color` | TERM environment variable |
| `env` | `{}` | Additional environment variables |

**Response:**
```json
{
  "method": "terminal.open",
  "result": {
    "session_id": "term_abc123",
    "pid": 12345,
    "cols": 80,
    "rows": 24
  }
}
```

### terminal.write

Write data to the terminal's stdin (user input).

**Request:**
```json
{
  "method": "terminal.write",
  "params": {
    "session_id": "term_abc123",
    "data": "bHMgLWxhCg==",  // base64-encoded stdin
    "encoding": "base64"
  }
}
```

| Parameter | Description |
|-----------|-------------|
| `data` | Base64-encoded data to write to stdin |
| `encoding` | Always `base64` (binary-safe) |

### terminal.output

Sent by the device when there is new terminal output.

**Event:**
```json
{
  "type": "event",
  "method": "terminal.output",
  "params": {
    "session_id": "term_abc123",
    "data": "b3duZXI6QHNoZWxsOiAvZGF0YS9idXp6cGkkIAo=",
    "encoding": "base64"
  }
}
```

The output is streamed as events (not batched responses). Output events are sent as they become available from the PTY.

### terminal.resize

Resize the terminal emulation window.

**Request:**
```json
{
  "method": "terminal.resize",
  "params": {
    "session_id": "term_abc123",
    "cols": 120,
    "rows": 40
  }
}
```

The device sends SIGWINCH to the child process and updates the PTY window size.

### terminal.close

Close a terminal session.

**Request:**
```json
{
  "method": "terminal.close",
  "params": {
    "session_id": "term_abc123"
  }
}
```

The device sends SIGHUP to the child process, waits for graceful shutdown (3 seconds), then SIGKILL if still running.

## Encoding

All terminal data is base64-encoded UTF-8. This ensures binary-safe transmission of:
- Regular text output
- ANSI escape sequences (colors, cursor movement, clear screen)
- Control characters (Ctrl-C, Ctrl-D, Tab completion)
- Unicode text (international characters, emoji — though discouraged in terminal)

## ANSI Support

The device PTY outputs raw ANSI escape sequences. Interpretation is the client's responsibility.

### Minimum Required Support

Clients MUST support:

| Feature | Sequences |
|---------|-----------|
| Text color (foreground) | 30-37, 38;5;n, 38;2;r;g;b |
| Text color (background) | 40-47, 48;5;n, 48;2;r;g;b |
| Bold/ dim/ italic | 1, 2, 3 |
| Underline | 4, 21, 22, 23, 24 |
| Cursor movement | CUU, CUD, CUF, CUB, CUP |
| Clear screen | ED 0, 1, 2 |
| Clear line | EL 0, 1, 2 |
| SGR reset | 0 |
| Scroll | SU, SD |
| Alternate screen | SM ?1049, RM ?1049 |
| Bracketed paste | SM ?2004, RM ?2004 |

### Recommended Support

Clients SHOULD support 256-color mode and true color (24-bit) for enhanced terminal experiences.

## Session Management

### Multiple Sessions

A client MAY open multiple terminal sessions to the same device simultaneously. Each session is independent (separate PTY, separate process).

### Session Limits

| Limit | Value |
|-------|-------|
| Max concurrent sessions per client | 5 |
| Max concurrent sessions per device | 10 |
| Session idle timeout | 30 minutes |
| Max session duration | 24 hours |

The device MUST enforce these limits. Exceeding limits returns error code `SESSION_LIMIT_EXCEEDED`.

## Security

### Shell Restriction

The device SHOULD restrict which shells can be launched:

| Allowed | Denied |
|---------|--------|
| `/bin/bash` | `/bin/su` |
| `/bin/zsh` | `/bin/sudo` |
| `/bin/sh` | `/bin/login` |
| `/bin/dash` | Custom restricted shells |

The device MAY allow additional shells based on extension configuration.

### Input Validation

The device MUST:
- Validate terminal dimensions (cols: 20-500, rows: 5-200)
- Validate data size per write (max 64KB)
- Rate-limit writes (max 100/sec per session)
- Sanitize control characters (optional, configurable)

### Session Isolation

Terminal sessions MUST be isolated from each other:
- Each session runs in its own process group
- Sessions are not aware of each other
- Sessions terminate when the client disconnects (SIGHUP)

## Protocol Versioning

The Terminal service is versioned independently:

| Version | Changes |
|---------|---------|
| 1 | Initial specification |

The version is negotiated via capability exchange during connection setup.
