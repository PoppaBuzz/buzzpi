# BuzzPi

## Vision

**BuzzPi** is a native Android companion app for Raspberry Pi that makes
connecting and managing Raspberry Pis effortless.

**Core principle:** \> The user should never have to know their IP
address.

## Goals

-   One-tap connection
-   Zero networking knowledge required
-   Native Android experience
-   Automatic discovery
-   Secure remote access
-   Powerful management tools

## Architecture

``` text
Android App
    │
Connection Engine
    │
Automatically selects:
- Existing session
- Local mDNS
- Local hostname
- Tailscale/MagicDNS
- WebRTC
- Cloud relay

        │

BuzzPi Agent (Go)

        │

Raspberry Pi
```

## BuzzPi Agent

Runs as a systemd service.

Responsibilities:

-   Device registration
-   Secure authentication
-   WebSocket communications
-   System monitoring
-   Remote command execution
-   Plugin hosting
-   Automatic updates

### Local API

-   /info
-   /stats
-   /files
-   /docker
-   /services
-   /logs
-   /camera
-   /gpio
-   /terminal

## Connection Philosophy

Users connect to devices---not IP addresses.

### Connection Engine Priority

1.  Existing session
2.  Local discovery (mDNS/DNS-SD)
3.  Local hostname
4.  Tailscale/MagicDNS
5.  WebRTC
6.  Cloud relay

Switch automatically if connectivity changes.

## Discovery Methods

### Local Discovery

Use Avahi/mDNS.

Device advertises:

-   Friendly name
-   Model
-   Available services

### QR Pairing

Installer generates a QR code containing:

-   Device ID
-   Public key
-   One-time pairing token

Scan and pair instantly.

### Bluetooth Pairing

Nearby BLE discovery for initial setup.

### USB Pairing

Optional pairing over USB for first-time configuration.

### Cloud Registration

Each Pi maintains a lightweight outbound connection to the BuzzPi cloud.

Stores:

-   Device ID
-   Friendly name
-   Last seen
-   Available transports
-   Capabilities

## Android Features

### Dashboard

-   CPU
-   Memory
-   Temperature
-   Storage
-   Network
-   Uptime

### Terminal

-   ANSI colors
-   Multiple tabs
-   Search
-   Copy/paste
-   History
-   Landscape support

### File Manager

-   Upload/download
-   Rename
-   Preview
-   Archive
-   Syntax highlighting

### Docker

-   Containers
-   Images
-   Compose
-   Logs
-   Shell
-   Resource monitoring

### GPIO

Interactive pin controls.

### Camera

-   Live stream
-   Snapshots
-   Recording

### Notifications

-   Overheating
-   Low storage
-   Service failures
-   Updates
-   UPS alerts

## Adaptive Dashboard

Detect installed software automatically.

Examples:

-   Pi-hole
-   Home Assistant
-   Jellyfin
-   CasaOS
-   Immich
-   Frigate
-   Node-RED
-   Syncthing

## AI Features

Built-in assistant capable of:

-   Explaining high CPU usage
-   Reading logs
-   Restarting services
-   Diagnosing issues
-   Answering Raspberry Pi questions

## Security

-   TLS
-   Certificate pinning
-   Public/private keys
-   Android Keystore
-   Biometrics
-   End-to-end encrypted sessions

## Technology Stack

### Android

-   Kotlin
-   Jetpack Compose
-   Material 3
-   Ktor
-   Room
-   Hilt
-   WorkManager
-   DataStore

### Agent

Go

### Backend

-   Go
-   PostgreSQL
-   Redis
-   Object storage

## Development Roadmap

### Phase 1

-   Android app
-   Agent
-   Pairing
-   LAN discovery
-   Terminal
-   File transfer
-   Live stats

### Phase 2

-   Remote access
-   Cloud registry
-   Notifications
-   Auto reconnect

### Phase 3

-   Docker
-   Camera
-   GPIO
-   Services
-   Logs

### Phase 4

-   AI assistant
-   Plugin SDK
-   Adaptive dashboards
-   Automation recipes

## Long-Term Vision

BuzzPi becomes the definitive companion app for Raspberry Pi.

Remote access is only one feature. The goal is a seamless, intelligent
experience where users manage every Pi they own without ever thinking
about networking or IP addresses.
