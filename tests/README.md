# Pablo Test Scenarios

This directory contains test scenarios organized by their target operating system and compatibility.

## Structure

- **`e2e/`**: Docker-based integration tests (Ubuntu SSH target). Run with `cd tests/e2e && go test -tags=integration -v ./...`. See [e2e/README.md](e2e/README.md).
- **`agnostic/`**: Platform-independent tests that should work on any OS using relative paths and generic features.
- **`windows/`**: Tests specifically for Windows, involving `C:\` paths, `.exe` builds, or Windows-specific system integration.
- **`linux/`**: Tests for Linux systems, involving absolute Unix paths, systemd services, etc.
- **`macos/`**: Tests for macOS environments.

## Schema v2 manifests

Fixtures use Schema v2: all deploy config under `profiles`, artifacts via `deploy.source` on each environment, SSH via `remote` (no `remote.method`), remote transfer via `deploy.transfer`.

## Scenarios (Agnostic)

### 1. Multi-Profile (`agnostic/multi-profile/`)

**One file, multiple profiles**

All applications in a single `pablo.yaml`:

```bash
pablo run -p frontend -e production -f agnostic/multi-profile/pablo.yaml
pablo run -p backend-api -e staging -f agnostic/multi-profile/pablo.yaml
```

**Best for:**
- Monorepos
- Centralized credential management
- Related applications deployed together

---

### 2. Separate Apps (`agnostic/separate-apps/`)

**One app per directory**

Each application has its own `pablo.yaml`:

```bash
cd agnostic/separate-apps/frontend && pablo run -e production
cd agnostic/separate-apps/backend && pablo run -e staging
```

**Best for:**
- Independent repositories
- Different teams
- Simpler, focused configurations

---

## Deployment Types Covered

Both scenarios include examples of all deployment types:

| Type | Description | Example |
|------|-------------|---------|
| `static` | Frontend SPA | React, Vue, Angular |
| `docker` | Containerized | Node.js, Python API |
| `binary` | Compiled | Go, Rust services |
| `git-sync` | Interpreted | PHP, Python apps |

---

## Quick Start

```bash
# Test multi-profile
cd agnostic/multi-profile
pablo check -f pablo.yaml

# Test separate apps
cd agnostic/separate-apps/frontend
pablo check -f pablo.yaml
```
