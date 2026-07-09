# Pablo

**Pablo** is a CLI deployment helper that automates building, filtering, and deploying artifacts across multiple environments. It supports local and remote (SSH) targets driven by a single YAML manifest.

> **Documentation:** [docs/](docs/README.md) — installation, guides, reference, and troubleshooting.

---

## Key Features

- **Multi-profile manifests** — manage frontend, backend, and infra in one `pablo.yaml`.
- **Multi-stage pipeline** — hooks, build, filter, deploy, template variables, health check.
- **Remote SSH deploy** — tar-based streaming for fast bulk transfers.
- **Safety checks** — protected system directory detection and automatic backups.
- **Self-deploy** — Pablo can build and install itself.
- **VS Code extension** — syntax highlighting, autocomplete, hover docs, and snippets for `pablo.yaml`.

---

## Requirements

- **Go** 1.25.5 or newer (only required to build from source)
- **Git** (only required for `git-sync` deployment type)
- **Docker** (only required for `docker` deployment type)
- **OpenSSH client / private key** (only required for remote SSH deploys)
- Supported host platforms: Windows, macOS, Linux

---

## Install

```bash
# Download from Releases, or build from source:
git clone https://github.com/septillioner/pablo.git
cd pablo && ./build.sh
pablo version
```

Full instructions: [docs/getting-started/installation.md](docs/getting-started/installation.md)

---

## Quick Start

```bash
pablo init
pablo check
pablo run -e production
```

See [docs/getting-started/quick-start.md](docs/getting-started/quick-start.md).

---

## CLI

```
pablo run        -p <profile> -e <env> [-f pablo.yaml] [--force]
pablo check      -f <file> [-p profile] [-e env]
pablo inspect    -f <file> [--json]
pablo init
pablo uninstall  -p <profile> -e <env> [--remove-backups]
pablo version
pablo lsp
```

Defaults: manifest = `pablo.yaml`, profile = `default`, env = `production`.

Full reference: [docs/reference/cli.md](docs/reference/cli.md)

---

## Deployment Types

| Type | Description | Local | Remote SSH |
|------|-------------|-------|------------|
| `static` | Frontend / SPA | Yes | Yes |
| `binary` | Compiled executables + PATH | Yes | Yes |
| `docker` | Docker Compose | Yes | Yes |
| `git-sync` | Git pull + post commands | Yes | Yes |

Details: [docs/reference/capabilities.md](docs/reference/capabilities.md)

---

## Known Limitations

- `blue-green` strategy and `deploy.service` (systemd/PM2) — schema only, not implemented at runtime.
- SSH host key verification disabled — see [SECURITY.md](SECURITY.md).
- Partial unit test coverage — see [docs/roadmap.md](docs/roadmap.md).

---

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) and [docs/development/](docs/development/contributing.md).

## Security

Report vulnerabilities per [SECURITY.md](SECURITY.md). Do not open public issues for security reports.

## Releasing

[RELEASING.md](RELEASING.md) · [docs/development/release-process.md](docs/development/release-process.md)

## License

Apache License 2.0 — see [LICENSE](LICENSE).

## Author

**Ege Ismail Kosedag** · [github.com/septillioner](https://github.com/septillioner)
