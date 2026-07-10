# BPP Chapter 15: Docker Service

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The Docker service provides remote management of Docker containers, images, and Compose projects on the device.

## Overview

This service communicates with the local Docker daemon. It requires the Docker extension to be installed on the device.

## Methods

### docker.containers.list

List all containers.

**Request:**
```json
{
  "method": "docker.containers.list",
  "params": {
    "all": false,
    "filters": {
      "status": ["running"],
      "label": ["com.buzzpi.managed=true"]
    }
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `all` | false | Include stopped containers |
| `filters` | null | Filter by status, label, name, etc. |

**Response:**
```json
{
  "method": "docker.containers.list",
  "result": {
    "containers": [
      {
        "id": "a1b2c3d4e5f6",
        "name": "nginx-proxy",
        "image": "nginx:latest",
        "status": "running",
        "state": "running",
        "created": "2026-07-01T10:00:00Z",
        "ports": [
          { "host_port": 80, "container_port": 80, "protocol": "tcp" }
        ],
        "cpu_usage_percent": 2.1,
        "memory_usage_mb": 45,
        "memory_limit_mb": 256,
        "restart_policy": "unless-stopped",
        "health": "healthy"
      }
    ],
    "total": 1
  }
}
```

### docker.containers.inspect

Get detailed container information.

**Request:**
```json
{
  "method": "docker.containers.inspect",
  "params": {
    "id": "a1b2c3d4e5f6"
  }
}
```

### docker.containers.logs

Get container logs.

**Request:**
```json
{
  "method": "docker.containers.logs",
  "params": {
    "id": "a1b2c3d4e5f6",
    "tail": 100,
    "since": "2026-07-07T10:00:00Z",
    "follow": false
  }
}
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `tail` | all | Number of lines from the end |
| `since` | null | Return logs after this timestamp |
| `follow` | false | Stream new log lines |

When `follow` is true, logs are streamed as events until the subscription is closed.

### docker.containers.start / stop / restart

**Request:**
```json
{
  "method": "docker.containers.restart",
  "params": {
    "id": "a1b2c3d4e5f6"
  }
}
```

### docker.images.list

List Docker images.

**Request:**
```json
{
  "method": "docker.images.list",
  "params": {}
}
```

**Response:**
```json
{
  "method": "docker.images.list",
  "result": {
    "images": [
      {
        "id": "sha256:abc123...",
        "repository": "nginx",
        "tag": "latest",
        "size_mb": 187,
        "created": "2026-06-15T00:00:00Z"
      }
    ],
    "total": 1
  }
}
```

### docker.images.pull

Pull a Docker image.

**Request:**
```json
{
  "method": "docker.images.pull",
  "params": {
    "image": "nginx:latest",
    "platform": "linux/arm64"
  }
}
```

For long-running pulls, progress events are streamed:
```json
{
  "type": "event",
  "method": "docker.images.pull.progress",
  "params": {
    "image": "nginx:latest",
    "status": "downloading",
    "progress_percent": 45,
    "downloaded_mb": 84,
    "total_mb": 187
  }
}
```

### docker.compose.list

List Docker Compose projects.

**Request:**
```json
{
  "method": "docker.compose.list",
  "params": {}
}
```

**Response:**
```json
{
  "method": "docker.compose.list",
  "result": {
    "projects": [
      {
        "name": "myapp",
        "path": "/home/pi/myapp/docker-compose.yml",
        "status": "running",
        "services": ["web", "db", "redis"],
        "container_count": 3
      }
    ]
  }
}
```

### docker.compose.up / down / restart

**Request:**
```json
{
  "method": "docker.compose.restart",
  "params": {
    "project": "myapp",
    "services": ["web"]
  }
}
```

### docker.system.info

Get Docker daemon information.

**Request:**
```json
{
  "method": "docker.system.info",
  "params": {}
}
```

**Response:**
```json
{
  "method": "docker.system.info",
  "result": {
    "version": "24.0.7",
    "api_version": "1.43",
    "containers": 5,
    "running": 3,
    "paused": 0,
    "stopped": 2,
    "images": 12,
    "storage_driver": "overlay2",
    "data_usage_mb": 2450,
    "cgroup_driver": "systemd",
    "experimental": false
  }
}
```

### docker.system.prune

Clean up unused Docker resources.

**Request:**
```json
{
  "method": "docker.system.prune",
  "params": {
    "containers": true,
    "images": true,
    "networks": true,
    "volumes": false,
    "builder_cache": true
  }
}
```

## Error Codes

| Code | Meaning |
|------|---------|
| DOCKER_NOT_INSTALLED | Docker daemon is not available on the device |
| DOCKER_PERMISSION_DENIED | User does not have permission to access Docker |
| CONTAINER_NOT_FOUND | Container ID does not exist |
| IMAGE_NOT_FOUND | Image does not exist locally |
| COMPOSE_NOT_FOUND | Compose project file not found |
| PULL_FAILED | Image pull failed (network, registry auth, disk space) |

## Extension

The Docker service requires the Docker extension to be installed on the device. This extension:
- Provides the `docker` binary path configuration
- Manages Docker socket permissions
- Handles `docker` group membership
- Configures auto-start behavior
