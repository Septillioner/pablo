# Pablo 🚀

**Pablo** is a production-ready CLI deployment automation tool designed to simplify and standardize the deployment process across multiple environments and application types.

## What is Pablo?

Pablo automates the entire deployment pipeline from build to health check, supporting multiple deployment strategies and application types. Whether you're deploying a React frontend, a Dockerized backend, a Go binary, or a PHP application, Pablo provides a unified, declarative approach to deployment.

### Key Features

✨ **Type-Based Deployments**
- `static`: Frontend SPAs (React, Vue, Angular)
- `docker`: Containerized applications
- `binary`: Compiled executables (Go, Rust)
- `git-sync`: Interpreted languages (PHP, Python)

🔐 **Centralized Credentials**
- SSH keys, passwords, tokens
- Reference credentials across profiles
- Secure credential management

🎯 **Multi-Environment Support**
- Development, staging, production
- Environment-specific variables
- Different deployment strategies per environment

🔄 **Complete Pipeline**
- Pre/post deployment hooks
- Build automation
- Artifact filtering
- Template variable injection
- Health checks
- Service management (systemd, PM2)

📦 **Flexible Configuration**
- Multi-profile: All apps in one file
- Separate configs: One file per app
- Backward compatible with legacy configs

### Quick Example

```yaml
name: my-app
version: 1.0.0

profiles:
  frontend:
    type: static
    build:
      command: npm run build
      output_dir: ./dist
    environments:
      production:
        deploy:
          target_path: /var/www/html
          strategy: backup
```

```bash
pablo run -p frontend -e production
```

---

# Pablo Test Scenarios

This directory contains test scenarios organized by their target operating system and compatibility.

## Structure

- **`agnostic/`**: Platform-independent tests that should work on any OS using relative paths and generic features.
- **`windows/`**: Tests specifically for Windows, involving `C:\` paths, `.exe` builds, or Windows-specific system integration.
- **`linux/`**: Tests for Linux systems, involving absolute Unix paths, systemd services, etc.
- **`macos/`**: Tests for macOS environments.

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
