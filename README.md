<p align="center">
  <img src="assets/logo.svg" width="120" height="120" alt="Pablo Logo">
</p>

# Pablo

**Pablo** is a CLI deployment helper that automates building, filtering, and deploying artifacts across multiple environments. It supports local and remote (SSH) targets driven by a single YAML manifest.

> Full schema reference: [`schema/schema.md`](schema/schema.md)
> Wiki (extended docs): [github.com/septillioner/pablo/wiki](https://github.com/septillioner/pablo/wiki)

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

### Option A — Download a pre-built binary (recommended)

1. Open the latest release: [Releases page](https://github.com/septillioner/pablo/releases)
2. Download the archive matching your OS / arch
   (`pablo-<os>-<arch>` for macOS/Linux, `pablo-<os>-<arch>.exe` for Windows).
3. Verify the SHA-256 checksum against `checksums.txt`.
4. Move the binary to a directory on your `PATH`
   (e.g. `/usr/local/bin/pablo` or `C:\Program Files\Pablo\pablo.exe`).
5. Verify the install:

```bash
pablo version
```

### Option B — Build from source

```bash
git clone https://github.com/septillioner/pablo.git
cd pablo

# Build for the current platform (output goes to ./build)
./build.sh

# Or build for all supported platforms
./build.sh all
```

The resulting binary is written to `build/pablo` (or `build/pablo.exe` on Windows).

### Option C — Self-deploy

Pablo can build and register itself in your `PATH` using its own pipeline:

```bash
# macOS / Linux
./publish-self.sh

# Windows (PowerShell, runs elevated)
./publish-self.ps1
```

After self-deploy, the `pablo` command is available globally.

---

## Quick Start

Create a minimal `pablo.yaml` (or run `pablo init` to generate a sample):

```yaml
name: my-app
version: 0.1.0
profiles:
  default:
    type: static
    build:
      command: npm run build
      output_dir: ./dist
    artifacts:
      include: ["**/*"]
    environments:
      production:
        deploy:
          target_path: /var/www/my-app
          strategy: backup
```

Then:

```bash
pablo check                         # Validate the manifest
pablo run -e production             # Execute the deployment
```

---

## CLI Commands

```
pablo run        -p <profile> -e <env> [-f pablo.yaml] [--force]
pablo check      -f <file> [-p profile] [-e env]
pablo init
pablo uninstall  -p <profile> -e <env> [--remove-backups]
pablo version
```

Defaults: manifest = `pablo.yaml`, profile = `default`, env = `production`.

---

## Deployment Types

| Type | Description | Local | Remote SSH | Status |
|------|-------------|-------|------------|--------|
| `static` | Frontend / SPA — build, filter artifacts, deploy files | Yes | Yes | Working |
| `binary` | Compiled executables — build, deploy, PATH register | Yes | Yes | Working |
| `docker` | Docker Compose — git clone/pull, compose up | Yes | No | Working (local) |
| `git-sync` | Interpreted languages — git pull, post commands | Yes | Yes | Working |

## Deploy Strategies

| Strategy | Description | Status |
|----------|-------------|--------|
| `overwrite` | Copy files over existing (default) | Working |
| `backup` | Rename existing dir with timestamp, then deploy | Working |
| `recreate` | Delete target dir, create fresh, deploy | Working |
| `blue-green` | Zero-downtime swap | Not implemented |

## Pipeline Phases

1. Pre-deploy hooks (`hooks.pre`)
2. Build (`build.command`)
3. Pre-deployment commands (`deploy.pre_commands`)
4. Deployment (local copy or SSH tar stream)
5. Post-deployment commands (`deploy.post_commands`)
6. Post-deploy hooks (`hooks.post`)
7. Health check (`pipeline.health_check`)

---

## What Works

- Full local deploy pipeline for `static` and `binary` types.
- Remote SSH deploy with tar-streaming and SCP fallback.
- Glob-based artifact filtering (include/exclude patterns).
- Template variable substitution (`{{VAR}}` in config files).
- Config inheritance — profile settings cascade into environments.
- Automatic PATH registration (Windows, macOS, Linux user scope).
- Backup and recreate strategies with protected path detection.
- `docker` type with local Docker Compose orchestration.
- `git-sync` with local and remote (SSH) git clone/pull.
- Environment variable injection via `.env` file generation.
- LSP-powered VS Code extension with completion, hover, and YAML validation.

## Known Limitations

- **No unit tests** — only YAML test fixtures exist under `tests/`.
- **`blue-green` strategy** — declared but not implemented (returns error).
- **SSH host key verification** — currently disabled (`InsecureIgnoreHostKey`); see [SECURITY.md](SECURITY.md).
- **Windows PATH uninstall** — `RemovePath` returns "not yet implemented" on Windows.
- **Remote docker deploy** — `docker` type only works locally; no remote SSH compose support.
- **LSP schema validation** — only YAML syntax errors are reported; semantic checks are TODO.
- **`filepath.Join` on remote paths** — may produce backslashes when a Windows host deploys to Linux.
- **`builder.Service`** — exists as a standalone service but is currently unused; builds run inline.
- **Snippet versions** — hardcoded; not synced with the `VERSION` file.

---

## Project Structure

```
src/                     Go CLI source (module: pablo)
  main.go                Entry point - cobra commands
  internal/
    domain/              Config, Profile, Environment, Deploy types
    config/              YAML loader + inheritance resolver
    services/
      pipeline/          Orchestrator - full deploy lifecycle
      deployer/          Local + SSH deploy with safety checks
      builder/           Build command executor (unused)
      filter/            Include/exclude file filtering
      scm/               Git clone/pull
      hooks/             Lifecycle hook runner
      health/            HTTP health check
      template/          {{VAR}} replacement
    adapters/
      ssh/               SSH connect, SCP, tar pipeline
      docker/            Docker Compose wrapper
      system/            PATH registration (cross-platform)
  pkg/ui/                Colored terminal output

extensions/
  pablo-lsp/             Go LSP server (completion, hover, validation)
  vscode-pablo/          VS Code extension (language client, snippets, syntax)

tests/                   YAML test fixtures
schema/                  Schema documentation
```

---

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening an issue or pull request.

## Security

If you discover a security vulnerability, please follow the disclosure process described in [SECURITY.md](SECURITY.md). Do **not** open a public issue for security reports.

## Releasing

The release process (version bump, build matrix, checksums, tagging, GitHub Release) is documented in [RELEASING.md](RELEASING.md).

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.

## Author

**Ege Ismail Kosedag**
[github.com/septillioner](https://github.com/septillioner)
