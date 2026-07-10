# Reference Design Constraints

These constraints define how the Reference book is maintained.

- Every fact exists in exactly one place. Duplication causes drift.
- Reference entries are machine-readable where possible (JSON Schema, protobuf, etc.).
- Design tokens are generated from source of truth, not manually copied.
- Reference is automatically verified against implementations during CI.
- Deprecated entries are marked with their deprecation version and removal date. They are never silently deleted.
- Every error code, event type, and capability ID has a unique, permanent identifier.
- Reference is updated atomically with the implementation change that introduces the reference entry.
