# Testing Standards

**BuzzPi ships software that people depend on to manage their devices.** Every change must be verified at the appropriate level.

---

## Test Pyramid

```
         ╱─────╲
        ╱  E2E  ╲          Few: critical user journeys
       ╱─────────╲
      ╱Integration╲        Some: component interactions, protocol
     ╱─────────────╲
    ╱   Unit Tests   ╲     Many: functions, methods, edge cases
   ╱───────────────────╲
```

### Expected Distribution

| Level | Coverage Target | Typical Ratio | Runtime |
|-------|----------------|---------------|---------|
| Unit | 80%+ | 70% of tests | <10ms each |
| Integration | 60%+ | 25% of tests | <1s each |
| E2E | Critical paths | 5% of tests | <30s each |

---

## Unit Tests

**Scope:** A single function, method, or module in isolation.

**Requirements:**
- No network calls, no disk I/O, no system dependencies
- Mock external dependencies (interfaces, not concrete types)
- Test both happy path and error paths
- Table-driven tests for functions with multiple input combinations

**Go example:**

```go
func TestSession_ValidateToken(t *testing.T) {
    tests := []struct {
        name    string
        token   string
        session Session
        want    error
    }{
        {"valid token", "sess_abc", Session{Token: "sess_abc", ExpiresAt: futureTime}, nil},
        {"expired token", "sess_abc", Session{Token: "sess_abc", ExpiresAt: pastTime}, ErrSessionExpired},
        {"wrong token", "sess_def", Session{Token: "sess_abc", ExpiresAt: futureTime}, ErrNotAuthenticated},
        {"empty token", "", Session{Token: "sess_abc", ExpiresAt: futureTime}, ErrNotAuthenticated},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.session.ValidateToken(tt.token)
            assert.ErrorIs(t, got, tt.want)
        })
    }
}
```

**Kotlin example:**

```kotlin
@Test
fun `validateToken returns Session for valid token`() {
    val session = Session(token = "sess_abc", expiresAt = futureTime)
    val result = session.validateToken("sess_abc")
    assertNotNull(result)
}

@Test
fun `validateToken returns null for expired token`() {
    val session = Session(token = "sess_abc", expiresAt = pastTime)
    val result = session.validateToken("sess_abc")
    assertNull(result)
}
```

---

## Integration Tests

**Scope:** Multiple modules working together (e.g., Engine Manager + Device Service + State Store).

**Requirements:**
- Test with real dependencies where feasible (test BoltDB, test PostgreSQL)
- Wire up components using their public interfaces
- Verify data flows end-to-end within the subsystem
- Close and clean up resources after each test

**Go example:**

```go
func TestIntegration_DevicePairing(t *testing.T) {
    // Real BoltDB (temp file)
    store := newTestStateStore(t)
    defer store.Close()

    // Real Engine Manager wired to real Device Service
    engine := NewEngineManager(store, log)
    deviceSvc := NewDeviceService(store)
    engine.RegisterHandler("device.info", deviceSvc.HandleInfo)

    // Execute through public API
    resp, err := engine.HandleRequest(ctx, &Request{
        Method: "device.info",
        Params: nil,
    })

    // Assert
    require.NoError(t, err)
    require.NotEmpty(t, resp.Result["device_id"])
}
```

---

## E2E Tests

**Scope:** Complete system from client to agent to cloud.

**Requirements:**
- Run against a real test environment or local dev stack
- Cover critical user journeys (pair, unpair, screen stream, terminal, file transfer)
- Test both LAN and relay connection paths
- Include destructive tests (agent crash, network drop, simultaneous clients)

**Critical journeys to cover:**

```
1. Launch app → discover device → pair via PIN → see dashboard
2. Dashboard → tap terminal → open session → type command → see output
3. Dashboard → tap screen → see remote desktop → tap/drag input
4. File manager → browse directory → download file
5. System → view monitoring stats
6. Settings → unpair device → device disappears from list
```

---

## Test Infrastructure

### CI Pipeline

```
PR opened → Lint → Unit Tests → Integration Tests → Build → E2E (critical)
                                                         ↓
PR merged → All tests → Build all targets → Publish artifacts
```

### Test Commands

```bash
# Go (Agent)
go test ./...                    # Unit + integration
go test -run Integration ./...   # Integration only
go test -short ./...             # Unit only (skips integration)

# Kotlin (Android)
./gradlew test                   # Unit tests
./gradlew connectedCheck         # Instrumentation tests

# E2E
make test-e2e                    # Full end-to-end suite
```

---

## Coverage Requirements

| Component | Minimum Coverage | Measured By |
|-----------|-----------------|-------------|
| Agent core (Go) | 80% | `go test -cover` |
| Agent protocol handlers | 90% | `go test -cover` |
| Android ViewModels | 85% | `./gradlew koverReport` |
| Android repositories | 80% | `./gradlew koverReport` |
| Backend API handlers | 85% | `go test -cover` |
| Plugin SDK (Go) | 90% | `go test -cover` |

Coverage below minimum blocks merge unless explicitly waived by a Core Maintainer.

---

## Test Quality Guidelines

### Good Tests

- **Deterministic** — same input always produces same output
- **Isolated** — tests do not depend on each other
- **Fast** — unit tests complete in milliseconds
- **Readable** — test name describes the scenario, assertions are clear
- **Minimal** — test only what it claims to test

### Bad Test Signals

- Tests that need `time.Sleep()` — use `clock` interfaces instead
- Tests that call `os.Exit()` — refactor to return errors
- Tests with shared mutable state — risk of flakiness
- Tests that pass without assertions (no `require`, no `assert`)
- Flaky tests — fix or remove; flakes erode trust in the suite

---

## Manual Testing

Some things cannot be automated:

- Screen streaming visual quality assessment
- Touch input latency on real hardware
- Pairing flow UX on a physical phone
- GPIO operation on a real Pi
- Camera preview quality

For these, maintain a **test checklist** in the GitHub release template and verify before each release.

---

## Performance Testing

| Test | Frequency | Target |
|------|-----------|--------|
| Agent CPU profile | Per release | <10% Pi 4 idle |
| Agent memory profile | Per release | <50 MB idle |
| Screen streaming latency | Per release | <200ms LAN |
| Terminal latency | Per release | <50ms LAN |
| Pairing time | Per release | <5s LAN |
| Concurrent client load | Per milestone | 5 clients stable |

---

## Version Compatibility Testing

BuzzPi must maintain backward compatibility within a major version:

- BPP v1 clients must work with BPP v1 agents at any minor version
- Android app v0.x must pair with agent v0.x
- New agent features must not break existing paired clients

Test matrix (run before each release):

```
Agent v0.1.0 + Android v0.1.0  ✓
Agent v0.1.0 + Android v0.2.0  ✓
Agent v0.2.0 + Android v0.1.0  ✓ (no new features, but existing features work)
Agent v0.1.0 + CLI v0.1.0      ✓
```
