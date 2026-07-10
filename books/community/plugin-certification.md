# Plugin Certification

**Plugins extend BuzzPi's capabilities. Certification ensures they meet quality, security, and reliability standards.**

---

## Certification Levels

| Level | Badge | Requirements | Use Case |
|-------|-------|-------------|----------|
| **Verified** | 🟢 | Automated checks pass, no known security issues, documented API | Community plugins, experimental |
| **Approved** | 🔵 | Verified + review by a maintainer, manifest completeness, test coverage, resource limits declared | Reliable plugins for daily use |
| **Official** | ⭐ | Approved + maintained by BuzzPi team, distributed in-app, monitored | First-party capabilities |

### Verified (🟢)

Automated certification — no human review required.

**Requirements:**
- Plugin manifest validates against schema
- No network access to unknown domains (declared in manifest)
- No filesystem access outside plugin directory (declared in manifest)
- Plugin starts and stops without errors
- All declared capabilities respond to `capability.info`
- No hardcoded secrets in the source

**Process:** Submit via GitHub. CI runs automated checks. If all pass, the plugin is listed as Verified.

### Approved (🔵)

Human review by a BuzzPi maintainer or plugin reviewer.

**Requirements (all of Verified plus):**
- Public source repository
- README with installation, configuration, and usage instructions
- At least one working example
- Capability manifest declares all required permissions
- Plugin handles start/stop/restart gracefully
- No excessive resource usage (CPU < 25%, memory < 128 MB)
- Plugin can be uninstalled cleanly

**Process:**
1. Submit PR adding plugin to the community registry
2. Maintainer reviews code, security, and documentation
3. Maintainer tests the plugin on a reference Pi setup
4. If approved, the plugin is published in the BuzzPi Plugin Registry

### Official (⭐)

Plugins maintained by the BuzzPi core team.

**Requirements (all of Approved plus):**
- Source in the BuzzPi GitHub organization
- Maintained by at least one Core Maintainer
- Included in the BuzzPi Agent distribution
- Continuous monitoring (error tracking, performance)
- Support commitment (bugs fixed within release cycle)
- Full test coverage (unit + integration)
- API documentation hosted on docs.buzzpi.dev

---

## Plugin Manifest Requirements

Every plugin must include a valid `plugin.yaml` manifest:

```yaml
# Required
id: docker   # lowercase, no spaces, no underscores
name: Docker Manager
version: "1.0.0"
author:
  name: Your Name
  email: your@email.com

# Capabilities
capabilities:
  - id: container.list
    description: List running containers
  - id: container.logs
    description: View container logs

# Permissions
permissions:
  network:                    # "none" | "registry" | "full"
    level: registry           # registry = docker.io only
    domains:
      - docker.io
      - auth.docker.io
  filesystem:
    paths:
      - /var/run/docker.sock  # Docker socket
    read_only: false
  system: false               # System access (process, network config)

# Resource declarations
resources:
  max_memory_mb: 64
  max_cpu_percent: 25
  startup_timeout_seconds: 15

# Metadata
tags:
  - docker
  - containers
  - devops
homepage: https://github.com/your/docker-plugin
license: MIT
```

---

## Certification Checklist (for Reviewers)

### Security Review

```
[ ] Plugin does not request unnecessary permissions
[ ] Network access is minimal and documented
[ ] Filesystem access is scoped to specific paths
[ ] No hardcoded credentials, tokens, or secrets
[ ] Input is validated before use
[ ] Plugin does not escalate privileges
[ ] No shell injection vectors
```

### Quality Review

```
[ ] Plugin starts and stops cleanly (test 5x)
[ ] Error messages are informative, not crashing
[ ] Plugin handles restart without data loss
[ ] Resource usage within declared limits
[ ] Logging follows BuzzPi structured logging format
[ ] Graceful degradation when dependencies are missing
```

### Documentation Review

```
[ ] README explains installation and configuration
[ ] All capabilities documented with examples
[ ] Troubleshooting section for common issues
[ ] Declares compatible BuzzPi Agent versions
[ ] Includes license file
```

---

## Registry

Certified plugins are listed in the **BuzzPi Plugin Registry** at `plugins.buzzpi.dev`.

### Registry Entry

```json
{
  "id": "docker",
  "name": "Docker Manager",
  "version": "1.0.0",
  "certification": "approved",
  "badge": "blue",
  "author": {
    "name": "Your Name",
    "url": "https://github.com/your"
  },
  "description": "Browse, manage, and inspect Docker containers on your Pi",
  "download_url": "https://plugins.buzzpi.dev/docker/v1.0.0/bundle.tar.gz",
  "checksum": "sha256:abc123...",
  "manifest_url": "https://plugins.buzzpi.dev/docker/v1.0.0/plugin.yaml",
  "tags": ["docker", "containers", "devops"],
  "updated_at": "2026-07-07T00:00:00Z"
}
```

---

## Revocation

A plugin's certification may be revoked if:

- A security vulnerability is discovered and not patched within 30 days (Verified) or 14 days (Approved)
- The plugin causes crashes or data loss in the Agent
- The plugin violates user privacy (exfiltrates data, tracks usage)
- The author requests removal
- The plugin is abandoned for 12+ months without notice

Revoked plugins are removed from the registry and the Agent prevents installation.

---

## Self-Certification

Before submitting for certification, plugin authors should run:

```bash
buzzpi plugin verify ./my-plugin/
```

This checks:
- Manifest schema validity
- Plugin starts and responds to `capability.info`
- All declared capabilities are accessible
- Permission usage matches declaration
- No obvious security issues (hardcoded IPs, secrets in source)

---

## Plugin Sandboxing

All plugins run in a sandboxed environment regardless of certification level:

| Resource | Verified | Approved | Official |
|----------|----------|----------|----------|
| Network | Declared-only | Declared-only | Declared-only |
| Filesystem | Plugin directory | Declared paths | Declared paths |
| System calls | Seccomp default | Seccomp default | Seccomp default |
| Process creation | Forbidden | Forbidden | With permission |
| Memory limit | 64 MB | 128 MB | 256 MB |
| CPU limit | 25% | 25% | 50% |

These limits are enforced by the BuzzPi Agent's Plugin Host and cannot be overridden by the plugin.
