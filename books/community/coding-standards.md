# Coding Standards

**BuzzPi's code is written once, read many times.** Consistency matters more than cleverness. These standards apply to all code in the BuzzPi repository.

---

## Language-Specific Standards

### Go (Agent / CLI / Backend)

**Style:**
- Follow `gofmt` and `go vet` — they are not optional
- Use `golangci-lint` with BuzzPi's configuration in `.golangci.yml`
- Maximum line length: 120 characters
- Use `log/slog` for structured logging, never `log.Printf`

**Naming:**
- Receiver names: 1-3 characters (`r *Runtime`, not `runtime *Runtime`)
- Interfaces: add `er` suffix for single-method (`Streamer`, `Dialer`)
- No stuttering: `device.DeviceInfo` → `device.Info`
- Package names: lowercase, single word, no underscores

**Error handling:**
- Wrap errors with context: `fmt.Errorf("read config: %w", err)`
- Never use `panic` in production code (test code is exempt)
- Sentinel errors use `var ErrNotFound = errors.New("not found")`
- Error types for structured errors: `type ConfigError struct { ... }`

**Concurrency:**
- Prefer goroutines + channels for concurrency, not `sync.Mutex` directly
- Use `context.Context` as first parameter for all blocking functions
- Never leak goroutines — ensure cleanup with `defer cancel()` or `sync.WaitGroup`

**Imports:**
```
// Group order: stdlib, third-party, internal
import (
    "context"
    "fmt"
    "os"

    "github.com/gorilla/websocket"
    "go.etcd.io/bbolt"

    "github.com/buzzpi/agent/internal/config"
)
```

### Kotlin (Android)

**Style:**
- Follow official Kotlin coding conventions
- Use `ktlint` for automated formatting
- No wildcard imports (`import java.io.*`) — always explicit

**Naming:**
- Variables and functions: `camelCase`
- Classes and interfaces: `PascalCase`
- Composable functions: `PascalCase` (`DeviceCard`, `TerminalScreen`)
- Constants: `SCREAMING_SNAKE_CASE` (top-level `const val`)

**Compose:**
- State hoisting: state flows down, events flow up
- Preview parameter providers for all reusable composables
- `remember` and `derivedStateOf` for derived state, never compute in composition
- Extract modifier chains into named val when reused

**Coroutines:**
- `viewModelScope.launch` for ViewModel coroutines
- Never use `GlobalScope` in production code
- Use `StateFlow` for UI state, never `MutableState` outside ViewModels

---

## Cross-Language Standards

### Error Handling

```
Every error must be:
1. Detected (check return value)
2. Logged (with context via slog)
3. Handled (retry, degrade, or propagate)
4. NEVER ignored (no empty catch/check blocks)
```

**Anti-patterns (forbidden):**

```go
// NEVER: Ignoring errors
val, _ := someFunc()

// NEVER: Empty catch
try { something() } catch { }
```

### Logging

```go
// GOOD - structured with context
slog.Info("device paired", "device_id", id, "client", client)

// BAD - unstructured
slog.Info(fmt.Sprintf("device %s paired by %s", id, client))
```

**Log levels:**
- `Debug` — detailed diagnostic info (disabled in production by default)
- `Info` — normal operational messages (pairing, starting, stopping)
- `Warn` — unexpected but handled situations (retry, degraded mode)
- `Error` — failures that need investigation (connection loss, config error)
- Never use `Fatal` outside of `main()` — return errors instead

### Testing

**Naming:**
- Test functions: `TestPackage_Behavior` (`TestDevice_Pairing`, `TestTerminal_Resize`)
- Test files: `<file>_test.go` or `<file>Test.kt`

**Assertions:**
- Go: use `testing` package + `sling` or `testify/assert`
- Kotlin: use kotlin.test + `Truth` or `kotlin.test.assert`

**Test structure (Go):**

```go
func TestDevice_PairingSuccess(t *testing.T) {
    // Setup
    device := newTestDevice(t)

    // Execute
    result, err := device.Pair(ctx, pairReq)
    t.Log("paired", "device", device.ID())

    // Assert
    require.NoError(t, err)
    assert.True(t, result.Paired)
}
```

---

## Documentation Standards

### Code Comments

```go
// GOOD - explains WHY, not WHAT
// Use different nonce for retry to prevent replay attacks
nonce := generateNonce(ctx, seed+retryCount)

// BAD - states the obvious
// Generate a nonce
nonce := generateNonce(ctx, seed+retryCount)
```

**Rules:**
- Comments explain *why* something is done a certain way, not *what* the code does
- Every exported symbol must have a doc comment
- TODO comments: include author and issue reference: `// TODO(pjhatfiel): handle edge case BUZZ-123`
- Fixme comments: include severity: `// FIXME(security): validate input before use`

---

## Commit Standards

**Format:**

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`

**Examples:**
```
feat(agent): add mDNS device discovery
fix(terminal): handle SIGWINCH race on resize
docs: update pairing protocol reference
```

**Rules:**
- First line ≤ 72 characters
- Use imperative mood ("add" not "added" or "adds")
- Reference issues/RFCs in footer: `RFC-0003`, `Closes #42`
- No co-authors or multiple authors — squash before merge

---

## Code Review Standards

### What reviewers look for:

1. **Correctness** — Does the code do what it claims?
2. **Testing** — Are there tests for the new behavior?
3. **Error handling** — Are all error paths handled?
4. **Concurrency safety** — Are there races, deadlocks, or leaks?
5. **API design** — Is the public API consistent with existing patterns?
6. **Performance** — Are there obvious inefficiencies?
7. **Security** — Are inputs validated? Are secrets exposed?

### Review velocity:

- Small PR (< 200 lines): review within 24 hours
- Medium PR (200-500 lines): review within 48 hours
- Large PR (> 500 lines): review within 72 hours (consider splitting)

---

## Enforcement

These standards are enforced by:
1. **CI automation**: `golangci-lint`, `ktlint`, `gofmt`, `go vet`
2. **Code review**: All PRs require at least one approval
3. **The BuzzPi SDK**: Official SDK packages codify best practices

Exceptions require explicit approval from a Core Maintainer and must be documented in the PR.
