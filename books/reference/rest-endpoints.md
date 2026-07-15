# REST API Endpoints

**Every backend API endpoint with request/response schemas.** Base URL: `https://jphat.net/buzzpi/api/v1`

---

## Authentication

### POST /v1/auth/signup

Create a new user account.

```http
POST /v1/auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password_123",
  "display_name": "User Name"
}
```

**Response:** `201 Created`

```json
{
  "user_id": "usr_abc12345",
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "rt_def67890",
  "expires_in": 900
}
```

**Errors:** `409 Conflict` (email exists), `422 Unprocessable` (validation)

### POST /v1/auth/login

Authenticate with email and password.

```http
POST /v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password_123"
}
```

**Response:** `200 OK` (same schema as signup)

**Errors:** `401 Unauthorized` (bad credentials), `429 Too Many Requests` (rate limited)

### POST /v1/auth/refresh

Obtain a new access token using a refresh token.

```http
POST /v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "rt_def67890"
}
```

**Response:** `200 OK`

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

### POST /v1/auth/logout

Revoke current refresh token.

```http
POST /v1/auth/logout
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "refresh_token": "rt_def67890"
}
```

**Response:** `204 No Content`

### POST /v1/auth/password-reset

Request a password reset email.

```http
POST /v1/auth/password-reset
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Response:** `202 Accepted` (always succeeds to prevent email enumeration)

### POST /v1/auth/password-reset/confirm

Complete password reset with token.

```http
POST /v1/auth/password-reset/confirm
Content-Type: application/json

{
  "token": "reset_token_abc123",
  "new_password": "new_secure_password"
}
```

**Response:** `200 OK`

```json
{
  "message": "Password reset successful"
}
```

---

## Devices

### GET /v1/devices

List all devices owned by the authenticated user.

```http
GET /v1/devices?state=online
Authorization: Bearer <access_token>
```

**Query Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `state` | string | Filter: `online`, `offline`, `paired`, `all` (default: `all`) |

**Response:** `200 OK`

```json
{
  "devices": [
    {
      "id": "dev_abc12345",
      "friendly_name": "Kitchen Pi",
      "model": "Raspberry Pi 5",
      "state": "online",
      "runtime_version": "0.1.0",
      "last_seen_at": "2026-07-07T11:59:00Z",
      "last_ip": "192.168.1.42",
      "connection_type": "direct",
      "paired_at": "2026-07-01T10:00:00Z",
      "capabilities": ["service.terminal", "service.screen"]
    }
  ]
}
```

### POST /v1/devices

Register a new device.

```http
POST /v1/devices
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "claim_token": "claim_abc12345",
  "identity_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI...",
  "friendly_name": "Kitchen Pi",
  "model": "Raspberry Pi 5",
  "runtime_version": "0.1.0"
}
```

**Response:** `201 Created`

```json
{
  "id": "dev_abc12345",
  "claim_token": "claim_abc12345",
  "friendly_name": "Kitchen Pi",
  "state": "paired"
}
```

### GET /v1/devices/:id

Get device details.

```http
GET /v1/devices/dev_abc12345
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

```json
{
  "id": "dev_abc12345",
  "friendly_name": "Kitchen Pi",
  "model": "Raspberry Pi 5",
  "state": "online",
  "runtime_version": "0.1.0",
  "identity_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI...",
  "last_seen_at": "2026-07-07T11:59:00Z",
  "last_ip": "192.168.1.42",
  "connection_type": "direct",
  "paired_at": "2026-07-01T10:00:00Z",
  "paired_by": "usr_abc12345",
  "capabilities": ["service.terminal", "service.screen"]
}
```

### PATCH /v1/devices/:id

Update device metadata.

```http
PATCH /v1/devices/dev_abc12345
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "friendly_name": "Living Room Pi"
}
```

**Response:** `200 OK`

```json
{
  "id": "dev_abc12345",
  "friendly_name": "Living Room Pi"
}
```

### DELETE /v1/devices/:id

Unpair and remove device.

```http
DELETE /v1/devices/dev_abc12345
Authorization: Bearer <access_token>
```

**Response:** `204 No Content`

### GET /v1/devices/:id/status

Get device online/offline status.

```http
GET /v1/devices/dev_abc12345/status
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

```json
{
  "device_id": "dev_abc12345",
  "state": "online",
  "last_seen_at": "2026-07-07T11:59:00Z",
  "connection_type": "direct"
}
```

---

## Pairing

### POST /v1/devices/:id/pair

Initiate pairing with a device.

```http
POST /v1/devices/dev_abc12345/pair
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "pairing_code": "A3K9M7"
}
```

**Response:** `200 OK`

```json
{
  "device_id": "dev_abc12345",
  "status": "paired",
  "paired_at": "2026-07-07T12:00:00Z"
}
```

### POST /v1/devices/:id/unpair

Remove pairing with a device.

```http
POST /v1/devices/dev_abc12345/unpair
Authorization: Bearer <access_token>
```

**Response:** `204 No Content`

---

## Discovery

### GET /v1/discover

Get all paired devices with current online state.

```http
GET /v1/discover
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

```json
{
  "devices": [
    {
      "id": "dev_abc12345",
      "friendly_name": "Kitchen Pi",
      "state": "online",
      "connection_type": "direct",
      "last_seen_at": "2026-07-07T11:59:00Z",
      "capabilities": ["service.terminal", "service.screen"]
    },
    {
      "id": "dev_fedcba98",
      "friendly_name": "Garage Pi",
      "state": "offline",
      "connection_type": null,
      "last_seen_at": "2026-07-06T08:00:00Z",
      "capabilities": ["service.terminal"]
    }
  ]
}
```

---

## Sessions (Admin)

### GET /v1/admin/sessions

List active relay sessions.

```http
GET /v1/admin/sessions?limit=50
Authorization: Bearer <admin_token>
```

**Response:** `200 OK`

```json
{
  "sessions": [
    {
      "id": "ses_abc12345",
      "device_id": "dev_abc12345",
      "user_id": "usr_abc12345",
      "transport": "direct",
      "started_at": "2026-07-07T11:00:00Z",
      "bytes_sent": 1500000,
      "bytes_received": 3000000
    }
  ],
  "total": 1
}
```

---

## Health

### GET /v1/health

Basic health check (no auth required).

```http
GET /v1/health
```

**Response:** `200 OK`

```json
{
  "status": "ok",
  "version": "0.1.0",
  "uptime_seconds": 86400,
  "db_connected": true,
  "relay_sessions": 5
}
```

### GET /v1/health/ready

Readiness check (database, relay, TURN).

```http
GET /v1/health/ready
```

**Response:** `200 OK` or `503 Service Unavailable`

```json
{
  "status": "ready",
  "checks": {
    "database": "ok",
    "relay": "ok",
    "turn": "ok"
  }
}
```

---

## Common Error Responses

### 400 Bad Request

```json
{
  "error": "validation_error",
  "message": "Invalid request body",
  "fields": {
    "email": "must be a valid email address",
    "password": "must be at least 8 characters"
  }
}
```

### 401 Unauthorized

```json
{
  "error": "unauthorized",
  "message": "Invalid or expired access token"
}
```

### 403 Forbidden

```json
{
  "error": "forbidden",
  "message": "You do not have access to this device"
}
```

### 404 Not Found

```json
{
  "error": "not_found",
  "message": "Device not found"
}
```

### 409 Conflict

```json
{
  "error": "conflict",
  "message": "Email already registered"
}
```

### 422 Unprocessable

```json
{
  "error": "unprocessable",
  "message": "Pairing code expired. Generate a new one."
}
```

### 429 Too Many Requests

```json
{
  "error": "rate_limited",
  "message": "Too many requests. Try again in 30 seconds.",
  "retry_after_seconds": 30
}
```

### 500 Internal Server Error

```json
{
  "error": "internal_error",
  "message": "An unexpected error occurred"
}
```

---

## Rate Limit Headers

All API responses include rate limit headers:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1688734800
```

On rate limit hit:

```http
Retry-After: 30
```

---

## Pagination

List endpoints support cursor-based pagination:

```http
GET /v1/devices?limit=20&cursor=dev_abc12345
```

```json
{
  "devices": [...],
  "next_cursor": "dev_xyz78901",
  "has_more": true
}
```

---

## WebSocket Upgrade

The relay WebSocket endpoint upgrades from HTTP:

```http
GET /v1/relay
Upgrade: websocket
Connection: upgrade
Sec-WebSocket-Version: 13
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Authorization: Bearer <access_token>
```

---

## API Versioning

The API version is specified in the URL path: `/v1/`, `/v2/`, etc.

Versioning policy:
- Backwards-compatible changes (new fields, new endpoints): MINOR version bump
- Breaking changes (removed fields, changed response shapes): MAJOR version bump
- Previous major version supported for 6 months after new major release
- Deprecation announced via `Sunset` and `Deprecation` response headers

```http
Deprecation: true
Sunset: Sat, 01 Jan 2027 00:00:00 GMT
```
