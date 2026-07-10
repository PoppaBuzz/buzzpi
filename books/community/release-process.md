# Release Process

**BuzzPi follows a time-based release cadence with semantic versioning.** Releases are predictable, automated where possible, and documented.

---

## Versioning Scheme

BuzzPi uses [Semantic Versioning 2.0](https://semver.org/) with the following mapping:

```
v<major>.<minor>.<patch>
```

| Bump | When | Example |
|------|------|---------|
| Major | Breaking protocol change, breaking API change, major architectural shift | v1.0.0 → v2.0.0 |
| Minor | New features, capability additions, non-breaking API changes | v0.1.0 → v0.2.0 |
| Patch | Bug fixes, security patches, performance improvements | v0.1.0 → v0.1.1 |

### Pre-release identifiers

```
v0.1.0-alpha.1     — Internal testing, unstable
v0.1.0-beta.1      — Community testing, feature-complete
v0.1.0-rc.1        — Release candidate, final verification
v0.1.0             — Stable release
```

### v0.x Special Rules

During v0.x development (before v1.0.0), minor versions may include breaking changes. Patch versions are still bug fixes only. Breaking changes must be clearly documented in the release notes.

---

## Release Cadence

| Version Range | Cadence | Notes |
|---------------|---------|-------|
| v0.0.x | On-demand | Foundation phase — releases when ready |
| v0.1.x - v0.9.x | Monthly | Regular development releases |
| v1.0.0+ | Quarterly | Stable releases |

---

## Release Artifacts

Each release produces:

| Artifact | Platform | Format |
|----------|----------|--------|
| BuzzPi Agent | Linux ARM64 (Pi 4/5), ARM (Pi Zero) | `.tar.gz` binary |
| BuzzPi Agent | Linux AMD64 | `.tar.gz` binary |
| BuzzPi CLI | Linux ARM64, AMD64, macOS ARM64 | `.tar.gz` binary |
| BuzzPi Android | Android | `.apk` + `.aab` (Play Store) |
| BuzzPi Backend | Linux AMD64 | Docker image `buzzpi/backend` |
| Source code | All | GitHub tag |
| Checksums | All | `SHA256SUMS` + GPG signature |

---

## Release Process

### Phase 1: Preparation (T-7 days)

1. Create release branch: `release/v<version>` from `main`
2. Run full test suite: `make test-all`
3. Run performance benchmarks, document results
4. Update version strings in:
   - `agent/internal/version/version.go`
   - `android/app/build.gradle.kts`
   - `backend/internal/version/version.go`
   - `docs/reference/platform-reference.md`
5. Update CHANGELOG.md (see changelog format below)

### Phase 2: Validation (T-3 days)

1. Run E2E test suite on physical Pi hardware
2. Verify compatibility matrix (agent + Android + CLI cross-versions)
3. Run security audit (dependency scanning, SAST)
4. Manual QA checklist:
   - [ ] Pairing flow (LAN)
   - [ ] Pairing flow (remote/relay)
   - [ ] Terminal session open/close/resize
   - [ ] Screen streaming start/stop
   - [ ] File browse/download
   - [ ] Device info/stats display
   - [ ] Unpairing and re-pairing
   - [ ] Agent restart with active sessions

### Phase 3: Release (T-0)

1. Tag release: `git tag -s v0.1.0 -m "v0.1.0"`
2. Push tag: `git push origin v0.1.0`
3. CI builds all artifacts automatically
4. CI publishes artifacts to GitHub Releases
5. CI builds and pushes Docker image
6. Create GitHub Release with changelog notes
7. Publish Android release to Play Store (internal track → beta → production)

### Phase 4: Post-Release

1. Merge release branch back to `main`
2. Bump version to next development version (`v0.2.0-dev`)
3. Announce release on:
   - GitHub Discussions
   - Community Discord/Matrix
   - Blog post (for significant releases)
4. Monitor error reports for 48 hours

---

## Hotfix Process

For critical bugs (security vulnerabilities, data loss, broken pairing):

1. Branch from the release tag: `hotfix/v0.1.1`
2. Apply the minimal fix
3. Fast-track through review (one approval from Core Maintainer)
4. Run minimal test suite (unit + integration for affected module)
5. Tag and release: `v0.1.1`
6. Merge hotfix branch back to `main`

Hotfixes skip the preparation and validation phases but still require verification.

---

## Changelog Format

```markdown
# Changelog

## [v0.1.0] - 2026-08-01

### Added
- Initial implementation of mDNS device discovery (#42)
- Terminal session with PTY multiplexer (#58)
- Screen streaming via WebRTC (#64)

### Changed
- Improved reconnection backoff to use jitter (#71)
- Increased default session timeout from 12h to 24h (#73)

### Fixed
- Race condition in session token rotation (#68)
- Memory leak in screen capture buffer pool (#69)
- PIN generation producing ambiguous characters (#70)

### Security
- Switched from SHA-256 to SHA-512 for token hashing (#72)
```

Every changelog entry must reference a GitHub issue or PR number.

---

## Version Compatibility Guarantees

| Component | Stable API | Guarantee |
|-----------|------------|-----------|
| BPP Protocol | From v1.0.0 | Backward compatible within major version |
| Agent CLI flags | From v0.5.0 | Flags stable within minor version |
| Android Intents | From v0.5.0 | Intents stable within minor version |
| Plugin SDK | From v0.5.0 | Public API stable within minor version |
| Config format | From v0.5.0 | Fields additive only, no removals |

---

## Pre-release Verification Checklist

Before cutting any release (including pre-release):

```
[ ] All CI checks pass
[ ] No unresolved CVEs in dependencies
[ ] LSP diagnostics clean on all changed files
[ ] Changelog updated with all changes
[ ] Version strings updated in all locations
[ ] E2E tests pass on physical Pi hardware
[ ] Compatibility verified with previous release
[ ] Release notes drafted
[ ] Manual QA checklist completed
[ ] Core Maintainer sign-off obtained
```
