# BPP Chapter 12: File Service

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The File service provides remote file system access — browsing, reading, writing, uploading, and downloading files on the device.

## Overview

The File service operates over the WebRTC data channel. File transfers are chunked for efficiency and support progress tracking.

## Methods

### files.list

List directory contents.

**Request:**
```json
{
  "method": "files.list",
  "params": {
    "path": "/home/pi",
    "include_hidden": false,
    "sort": "name",
    "order": "asc"
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `path` | — | Absolute path to list |
| `include_hidden` | false | Include dot-files |
| `sort` | `name` | `name`, `size`, `modified`, `type` |
| `order` | `asc` | `asc`, `desc` |

**Response:**
```json
{
  "method": "files.list",
  "result": {
    "path": "/home/pi",
    "total_entries": 15,
    "entries": [
      {
        "name": "Documents",
        "type": "directory",
        "size": 4096,
        "mode": "drwxr-xr-x",
        "modified": "2026-07-07T12:00:00Z",
        "permissions": "0755",
        "owner": "pi",
        "group": "pi"
      },
      {
        "name": "config.txt",
        "type": "file",
        "size": 2340,
        "mode": "-rw-r--r--",
        "modified": "2026-06-15T08:30:00Z",
        "permissions": "0644",
        "owner": "pi",
        "group": "pi"
      }
    ]
  }
}
```

### files.read

Read a file's contents.

**Request:**
```json
{
  "method": "files.read",
  "params": {
    "path": "/home/pi/config.txt",
    "encoding": "base64",
    "offset": 0,
    "limit": 65536
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `encoding` | `base64` | `base64` (binary-safe) or `text` |
| `offset` | 0 | Read starting at this byte |
| `limit` | 65536 | Maximum bytes to read (max 1MB) |

**Response:**
```json
{
  "method": "files.read",
  "result": {
    "path": "/home/pi/config.txt",
    "size": 2340,
    "offset": 0,
    "data": "IyBCdXp6UGkgQ29uZmlndXJhdGlvbgpuZXR3b3JrLmhvc3RuYW1lPSBidXp6cGktYnJvYWQK",
    "encoding": "base64"
  }
}
```

For large files, read in chunks using offset and limit. Each chunk is a separate request.

### files.write

Write data to a file.

**Request:**
```json
{
  "method": "files.write",
  "params": {
    "path": "/home/pi/new_config.txt",
    "data": "IyBOZXcgQ29uZmlndXJhdGlvbg==",
    "encoding": "base64",
    "mode": "create",
    "permissions": "0644"
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `mode` | `create` | `create` (fail if exists), `overwrite`, `append` |
| `permissions` | `0644` | File permission bits |

**Response:**
```json
{
  "method": "files.write",
  "result": {
    "path": "/home/pi/new_config.txt",
    "size": 25,
    "mode": "create"
  }
}
```

For large files, split into chunks (max 64KB per chunk) and send sequentially. Each chunk includes an index for verification.

### files.delete

Delete a file or empty directory.

**Request:**
```json
{
  "method": "files.delete",
  "params": {
    "path": "/home/pi/old_file.txt",
    "recursive": false
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `recursive` | false | Delete directories recursively |

### files.rename

Rename or move a file.

**Request:**
```json
{
  "method": "files.rename",
  "params": {
    "source": "/home/pi/temp.txt",
    "destination": "/home/pi/Documents/final.txt"
  }
}
```

### files.mkdir

Create a directory.

**Request:**
```json
{
  "method": "files.mkdir",
  "params": {
    "path": "/home/pi/NewFolder",
    "permissions": "0755",
    "parents": false
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `parents` | false | Create parent directories (-p) |

### files.stat

Get detailed file information.

**Request:**
```json
{
  "method": "files.stat",
  "params": {
    "path": "/home/pi/project"
  }
}
```

**Response:**
```json
{
  "method": "files.stat",
  "result": {
    "name": "project",
    "type": "directory",
    "size": 4096,
    "mode": "drwxr-xr-x",
    "modified": "2026-07-07T12:00:00Z",
    "permissions": "0755",
    "owner": "pi",
    "group": "pi",
    "nlink": 3
  }
}
```

### files.watch

Watch a directory for changes (inotify).

**Request:**
```json
{
  "method": "files.watch",
  "params": {
    "path": "/home/pi/Documents",
    "recursive": true,
    "events": ["create", "modify", "delete"]
  }
}
```

**Events (streamed):**
```json
{
  "type": "event",
  "method": "files.watch.event",
  "params": {
    "path": "/home/pi/Documents/new_file.txt",
    "event": "create",
    "timestamp": "2026-07-07T12:00:05Z"
  }
}
```

Unwatch by closing the subscription (no explicit method — the client closes the event stream).

## Transfer Protocol

For large file transfers (uploads and downloads), BPP uses a chunked transfer protocol over the data channel:

### Upload (Client → Device)

1. Client sends `files.write` with first chunk
2. Device responds with `files.write.chunk_ack` (received offset and size)
3. Client sends next chunk
4. Repeat until all chunks sent
5. Device sends final `files.write.complete` with checksum

### Download (Device → Client)

1. Client sends `files.read` with offset and limit
2. Device responds with chunk data
3. Client sends `files.read.next` for next chunk
4. Repeat until device returns data with `size < limit` (end of file)
5. Client reassembles chunks

### Checksums

Each chunk includes an optional CRC32 checksum for integrity verification:

```json
{
  "method": "files.write",
  "params": {
    "path": "/home/pi/large_file.bin",
    "data": "<chunk_data>",
    "encoding": "base64",
    "offset": 65536,
    "checksum": "a1b2c3d4",
    "checksum_type": "crc32"
  }
}
```

## Path Security

### Path Validation

The device MUST validate all file paths:

| Rule | Enforcement |
|------|-------------|
| No symlink traversal | Resolve symlinks and verify path is within allowed roots |
| No escape sequences | Reject paths containing `..` outside allowed roots |
| No system paths | Block access to `/etc`, `/sys`, `/proc`, `/dev` by default |
| Root restriction | Confine all operations to `/home` and `/var/lib/buzzpi` by default |
| Permissions | Enforce OS file permissions (cannot read `/root/.ssh/id_rsa`) |

### Allowed Roots (Default)

| Path | Access |
|------|--------|
| `/home/*` | Read/write |
| `/var/lib/buzzpi/*` | Read/write |
| `/tmp/*` | Read/write |
| `/etc/buzzpi/*` | Read-only |
| Everywhere else | Blocked |

The device MAY extend allowed roots via extension configuration.

## Rate Limits

| Limit | Value |
|-------|-------|
| Max file listing entries | 1000 per directory |
| Max file size (single chunk) | 64 KB |
| Max file size (total upload) | 100 MB (configurable) |
| Max concurrent transfers | 3 per client |
| Max transfers per minute | 60 per device |
