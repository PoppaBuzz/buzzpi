# JSON Schemas

**Formal JSON Schema definitions for all BPP protocol messages.** These schemas are the authoritative specification for message validation. Implementations should use these schemas to validate messages at protocol boundaries.

---

## BPP Envelope

All BPP messages use this envelope schema.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://jphat.net/buzzpi/bpp/schemas/envelope.json",
  "title": "BPP Message Envelope",
  "description": "Universal envelope for all BPP protocol messages",
  "type": "object",
  "required": ["v", "id", "ts", "type"],
  "properties": {
    "v": {
      "type": "integer",
      "description": "Protocol version",
      "minimum": 1,
      "maximum": 1
    },
    "id": {
      "type": "string",
      "description": "Unique message identifier",
      "pattern": "^msg_[a-zA-Z0-9]+$"
    },
    "ts": {
      "type": "string",
      "description": "ISO 8601 timestamp (UTC)",
      "format": "date-time"
    },
    "type": {
      "type": "string",
      "description": "Message type",
      "enum": ["request", "response", "event", "error"]
    },
    "method": {
      "type": "string",
      "description": "Method being called (required for request and event)",
      "pattern": "^[a-z]+\\.[a-z]+(\\.[a-z]+)*$"
    },
    "params": {
      "type": "object",
      "description": "Method parameters (required for request and event)"
    },
    "rid": {
      "type": "string",
      "description": "Request ID (present on requests, echoed on responses)",
      "pattern": "^req_[a-zA-Z0-9]+$"
    },
    "result": {
      "description": "Response payload (present on responses)"
    },
    "error": {
      "type": "object",
      "description": "Error payload (present on error responses)",
      "properties": {
        "code": {
          "type": "string",
          "description": "Error code from registry"
        },
        "message": {
          "type": "string",
          "description": "Human-readable error description"
        },
        "data": {
          "description": "Additional error context"
        }
      },
      "required": ["code", "message"]
    }
  },
  "allOf": [
    {
      "if": {"properties": {"type": {"const": "request"}}},
      "then": {"required": ["method", "rid"]}
    },
    {
      "if": {"properties": {"type": {"const": "event"}}},
      "then": {"required": ["method"]}
    },
    {
      "if": {"properties": {"type": {"const": "response"}}},
      "then": {"required": ["rid"]}
    },
    {
      "if": {"properties": {"type": {"const": "error"}}},
      "then": {"required": ["rid", "error"]}
    }
  ]
}
```

---

## Device Methods

### device.info Response

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://jphat.net/buzzpi/bpp/schemas/device.info.json",
  "title": "Device Info Response",
  "type": "object",
  "required": ["device_id", "friendly_name", "runtime_version", "capabilities"],
  "properties": {
    "device_id": {
      "type": "string",
      "pattern": "^dev_[a-zA-Z0-9]+$"
    },
    "friendly_name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 64
    },
    "model": {
      "type": "string",
      "maxLength": 128
    },
    "runtime_version": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+\\.\\d+$"
    },
    "uptime_seconds": {
      "type": "integer",
      "minimum": 0
    },
    "capabilities": {
      "type": "array",
      "items": {
        "type": "string"
      }
    },
    "platform": {
      "type": "string",
      "pattern": "^[a-z]+/[a-z0-9_]+$"
    }
  }
}
```

### device.stats Response

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://jphat.net/buzzpi/bpp/schemas/device.stats.json",
  "title": "Device Stats Response",
  "type": "object",
  "required": ["cpu", "memory", "storage", "uptime_seconds"],
  "properties": {
    "cpu": {
      "type": "object",
      "required": ["usage_percent", "temperature_celsius"],
      "properties": {
        "usage_percent": {"type": "number", "minimum": 0, "maximum": 100},
        "temperature_celsius": {"type": "number", "minimum": -40, "maximum": 125},
        "frequency_mhz": {"type": "integer", "minimum": 0}
      }
    },
    "memory": {
      "type": "object",
      "required": ["total_mb", "used_mb", "available_mb", "percent"],
      "properties": {
        "total_mb": {"type": "integer", "minimum": 0},
        "used_mb": {"type": "integer", "minimum": 0},
        "available_mb": {"type": "integer", "minimum": 0},
        "percent": {"type": "number", "minimum": 0, "maximum": 100}
      }
    },
    "storage": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["mount", "total_mb", "used_mb", "available_mb", "percent"],
        "properties": {
          "mount": {"type": "string"},
          "total_mb": {"type": "integer", "minimum": 0},
          "used_mb": {"type": "integer", "minimum": 0},
          "available_mb": {"type": "integer", "minimum": 0},
          "percent": {"type": "number", "minimum": 0, "maximum": 100}
        }
      }
    },
    "uptime_seconds": {
      "type": "integer",
      "minimum": 0
    }
  }
}
```

---

## Terminal Methods

### terminal.open Request

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://jphat.net/buzzpi/bpp/schemas/terminal.open.json",
  "title": "Terminal Open Request",
  "type": "object",
  "required": ["rows", "cols"],
  "properties": {
    "rows": {
      "type": "integer",
      "minimum": 10,
      "maximum": 200,
      "default": 40
    },
    "cols": {
      "type": "integer",
      "minimum": 20,
      "maximum": 500,
      "default": 80
    },
    "shell": {
      "type": "string",
      "default": "/bin/bash"
    }
  }
}
```

### terminal.open Response

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://jphat.net/buzzpi/bpp/schemas/terminal.open.response.json",
  "title": "Terminal Open Response",
  "type": "object",
  "required": ["session_id", "rows", "cols"],
  "properties": {
    "session_id": {
      "type": "string",
      "pattern": "^term_[a-zA-Z0-9]+$"
    },
    "rows": {"type": "integer"},
    "cols": {"type": "integer"}
  }
}
```

---

## Screen Methods

### screen.start Request

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://jphat.net/buzzpi/bpp/schemas/screen.start.json",
  "title": "Screen Start Request",
  "type": "object",
  "properties": {
    "quality": {
      "type": "string",
      "enum": ["high", "medium", "low", "minimum"],
      "default": "high"
    },
    "max_fps": {
      "type": "integer",
      "minimum": 1,
      "maximum": 60,
      "default": 30
    },
    "max_resolution": {
      "type": "string",
      "enum": ["640x360", "854x480", "1280x720", "1920x1080"],
      "default": "1920x1080"
    },
    "cursor_visible": {
      "type": "boolean",
      "default": true
    }
  }
}
```

---

## Capability Methods

### capabilities.list Response

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://jphat.net/buzzpi/bpp/schemas/capabilities.list.json",
  "title": "Capability List Response",
  "type": "object",
  "required": ["capabilities"],
  "properties": {
    "capabilities": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "version", "available"],
        "properties": {
          "id": {
            "type": "string",
            "pattern": "^[a-z]+\\.[a-z]+(\\.[a-z]+)*$"
          },
          "version": {
            "type": "string",
            "pattern": "^\\d+\\.\\d+$"
          },
          "available": {
            "type": "boolean"
          },
          "params": {
            "type": "object",
            "description": "Capability-specific parameters"
          }
        }
      }
    }
  }
}
```

---

## Error Response

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://jphat.net/buzzpi/bpp/schemas/error.json",
  "title": "BPP Error",
  "type": "object",
  "required": ["code", "message"],
  "properties": {
    "code": {
      "type": "string",
      "description": "Error code from the BPP error registry"
    },
    "message": {
      "type": "string",
      "description": "Human-readable error description",
      "maxLength": 500
    },
    "data": {
      "type": "object",
      "description": "Additional error context (optional)"
    }
  }
}
```

---

## Common Method Parameter Schemas

### Pagination

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://jphat.net/buzzpi/bpp/schemas/pagination.json",
  "title": "Pagination Parameters",
  "type": "object",
  "properties": {
    "cursor": {
      "type": "string",
      "description": "Pagination cursor for next page"
    },
    "limit": {
      "type": "integer",
      "minimum": 1,
      "maximum": 100,
      "default": 20
    }
  }
}
```

---

## Validation Helper Functions

### Go

```go
import (
    "github.com/santhosh-tekuri/jsonschema/v5"
    "github.com/santhosh-tekuri/jsonschema/v5/httploader"
)

var schemaCache = make(map[string]*jsonschema.Schema)

func ValidateBPPEnvelope(data []byte) error {
    return validate("https://jphat.net/buzzpi/bpp/schemas/envelope.json", data)
}

func validate(schemaURL string, data []byte) error {
    schema, ok := schemaCache[schemaURL]
    if !ok {
        var err error
        schema, err = jsonschema.Compile(schemaURL)
        if err != nil {
            return err
        }
        schemaCache[schemaURL] = schema
    }

    var v interface{}
    if err := json.Unmarshal(data, &v); err != nil {
        return err
    }

    return schema.Validate(v)
}
```

### Python

```python
import json
import jsonschema
from referencing import Registry, Resource

# Load schemas from registry
with open("schemas/envelope.json") as f:
    envelope_schema = json.load(f)

def validate_bpp_message(data: dict) -> None:
    """Validate a BPP message against the envelope schema."""
    jsonschema.validate(data, envelope_schema)
```

---

## Schema Distribution

Schemas are distributed as part of the BPP specification and can be fetched from:

```
https://jphat.net/buzzpi/bpp/schemas/{schema_name}.json
```

The complete schema set is also bundled with each Runtime release in the `schemas/` directory and published as a standalone package: `@buzzpi/bpp-schemas` (npm) and `buzzpi-bpp-schemas` (PyPI).
