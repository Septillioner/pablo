# Pablo

**Pablo** is a CLI deployment helper that automates building, filtering, and deploying artifacts across local and remote (SSH) environments from a single YAML manifest (`pablo.yaml`).

> **Documentation:** [docs/](docs/README.md) — installation, guides, reference, and troubleshooting.

---

## Key Features

- **Multi-profile manifests** — manage frontend, backend, and infra in one `pablo.yaml`.
- **Multi-stage pipeline** — hooks, build, filter, deploy, template variables, health check.
- **Remote SSH deploy** — tar-based streaming for fast bulk transfers.
- **Safety checks** — protected system directory detection and automatic backups.
- **Self-deploy** — Pablo can build and install itself.
- **VS Code extension** — LSP-powered completion, hover docs, validation, and Run commands for `pablo.yaml`.

---

## Requirements

| Component | When needed |
|-----------|-------------|
| **Go 1.25.5+** | Building from source only |
| **Git** | `git-sync` and `docker` deployment types |
| **Docker** | `docker` deployment type |
| **OpenSSH client** | Remote SSH deploys |

Supported host platforms: Windows, macOS, Linux.

---

## Install

**Windows (PowerShell):**

```powershell
$s="$env:TEMP\pablo-install.ps1"; [Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; irm 'https://raw.githubusercontent.com/septillioner/pablo/master/install.ps1' -OutFile $s; powershell -NoProfile -ExecutionPolicy Bypass -File $s
```

Downloads the installer to a temp file and runs it in a clean PowerShell session (avoids `iex` / profile issues).

**Windows (cmd):**

```bat
curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/master/install.cmd -o install.cmd && install.cmd
```

Run from **PowerShell**, not cmd.

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/master/install.sh | bash
```

Pin a version: set `PABLO_VERSION=v1.4.0` before running the one-liner.

Or download from [Releases](https://github.com/septillioner/pablo/releases), build from source, or use self-deploy.

Full instructions: [docs/getting-started/installation.md](docs/getting-started/installation.md)

---

## Quick Start

```bash
pablo init
pablo check
pablo run -p default -e production
# or: pablo run default/production
```

See [docs/getting-started/quick-start.md](docs/getting-started/quick-start.md).

---

## CLI

```
pablo run [profile/env]  -p <profile> -e <env> [-f pablo.yaml] [--force]
pablo check              -f <file> [-p profile] [-e env]
pablo inspect            -f <file> [--json]
pablo init              [-t|--template]
pablo uninstall          -p <profile> -e <env> [--remove-backups]
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

- `blue-green` strategy — declared in schema, not implemented at runtime.
- `deploy.service` (systemd/PM2) — schema only; use `post_commands` instead.
- SSH host key verification disabled — see [SECURITY.md](SECURITY.md).

More: [docs/reference/capabilities.md](docs/reference/capabilities.md) · [docs/roadmap.md](docs/roadmap.md)

---

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) and [docs/development/contributing.md](docs/development/contributing.md).

## Security

Report vulnerabilities per [SECURITY.md](SECURITY.md). Do not open public issues for security reports.

## Releasing

[RELEASING.md](RELEASING.md) · [docs/development/release-process.md](docs/development/release-process.md)

## License

Apache License 2.0 — see [LICENSE](LICENSE).

## Author

**Ege Ismail Kosedag** · [github.com/septillioner](https://github.com/septillioner)
