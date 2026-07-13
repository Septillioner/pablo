# Contributing to Pablo

Thanks for your interest in contributing. This document is the canonical contribution guide. A short summary also lives at [docs/development/contributing.md](docs/development/contributing.md).

## Ground rules

- Open an issue before starting non-trivial work to align on direction.
- Keep changes focused. Prefer small, reviewable PRs over large mixed ones.
- Match the existing code style and package layout — see [Architecture](docs/development/architecture.md) and [PMAP.md](PMAP.md).
- Do not introduce dependencies casually. Justify any new direct dependency in the PR description.
- Check [docs/roadmap.md](docs/roadmap.md) for the current backlog; update checklist items when your PR completes a tracked goal.

## Project docs

| Document | Purpose |
|----------|---------|
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/roadmap.md](docs/roadmap.md) | Public roadmap — shipped features and planned work |
| [docs/development/architecture.md](docs/development/architecture.md) | Components, pipeline, services and adapters |
| [docs/development/testing.md](docs/development/testing.md) | Unit, E2E, and manual fixture testing |
| [docs/reference/cli.md](docs/reference/cli.md) | CLI command reference |
| [docs/reference/capabilities.md](docs/reference/capabilities.md) | Supported types, strategies, limitations |
| [docs/reference/configuration.md](docs/reference/configuration.md) | `pablo.yaml` field reference |
| [docs/reference/api.md](docs/reference/api.md) | `inspect --json` and LSP protocol surface |
| [PMAP.md](PMAP.md) | Internal architecture map (maintainers and agents; not user-facing) |
| [tests/TEST_SPEC.md](tests/TEST_SPEC.md) | Test catalog and coverage status |
| [RELEASING.md](RELEASING.md) | Version bump and publish process |

## Project setup

### Prerequisites

| Tool | When needed |
|------|-------------|
| Go 1.25.5+ | Building or changing the CLI |
| Git | Always |
| Docker + `ssh-keygen` | E2E suite (`tests/e2e/`) |
| Node.js 20+ | `extensions/vscode-pablo` |
| Visual Studio 2022/2026 + VSSDK workload | `extensions/vs2026` |

### Build

```bash
# Current platform
./build.sh

# All supported platforms
./build.sh all
```

Binaries are written to `build/`.

### Run from source

```bash
cd src
go run main.go check -f ../pablo.yaml
go run main.go run -f ../pablo.yaml default/production
go run main.go run -e <env> -p <profile> -f ../pablo.yaml
go run main.go run sequence <name> -f ../pablo.yaml
go run main.go inspect -f ../pablo.yaml --json
```

### Validate changes

Prefer the test runner from the repo root:

```bash
./test.sh all          # unit + integration + e2e (Docker required for e2e)
./test.sh unit         # fast package tests only
```

Windows:

```powershell
.\test.ps1 unit
.\test.ps1 e2e
# or: test.bat unit
```

Or run Go tests directly:

```bash
cd src
go test ./...
```

Manual fixtures under `tests/`:

```bash
# Validate only
go run ./src/main.go check -f tests/agnostic/local-deploy/pablo.yaml

# Deploy (writes real files)
cd tests/agnostic/local-deploy
go run ../../../src/main.go run -p local-test -e dev
```

For remote SSH / Docker scenarios (requires Docker):

```bash
./test.sh e2e
# or:
cd tests/e2e
go test -tags=integration -v -timeout 10m ./...
```

Full testing guide: [docs/development/testing.md](docs/development/testing.md).

## Code conventions

- Go 1.25.5; keep direct dependencies minimal (`cobra`, `fatih/color`, `yaml.v3`, `x/crypto`, `glsp` for LSP).
- Concrete struct dependencies, no interfaces / mocking layer.
- DI is wired in `src/main.go`.
- Semantic validation lives in `pkg/validate` — shared by `check`, `run`, and `pablo lsp`.
- Logging goes through `pkg/ui` (`ui.Log(mark, msg)`):
  `+` success, `-` error, `!` warn, `*` info, `>` action.
- Build commands: `sh -c` on Unix, `cmd /C` on Windows.
- Hooks: `powershell` on Windows, `sh` elsewhere.
- Template variables use `{{KEY}}` and are only expanded for config-like file extensions.
- Remote paths: `pathutil.JoinRemote`; local paths: `filepath`.
- Avoid emojis in code, comments, logs, and commit messages.
- Comments should explain non-obvious intent only. Do not narrate the code.

### Where to change what

| Change | Touch |
|--------|-------|
| Manifest field | `pkg/domain` → `pkg/config` → `pkg/validate` → `pkg/schema` → services/adapters → pipeline |
| CLI command / flag | `src/main.go` + [CLI reference](docs/reference/cli.md) |
| Validation rule | `pkg/validate` (single source for CLI + LSP) |
| LSP completion / hover | `pkg/schema` + `internal/lsp` |
| VS Code extension | `extensions/vscode-pablo/` |
| Visual Studio extension | `extensions/vs2026/` |

## Extensions

### VS Code (`extensions/vscode-pablo`)

- Spawns `pablo lsp` for diagnostics, completion, hover, and CodeLens Run.
- After CLI or LSP changes that affect the editor, rebuild the CLI and verify against a real `pablo.yaml`.
- See [VS Code guide](docs/guides/vscode.md) and the extension [README](extensions/vscode-pablo/README.md).

### Visual Studio (`extensions/vs2026`)

- VSIX with LSP, tool window, toolbar, and feature parity goals with the VS Code extension.
- Build: `extensions\vs2026\build-vs2026.bat`
- Debug: open `extensions\vs2026\Pablo.sln` in Visual Studio, F5 into the Experimental Instance.
- See [Visual Studio guide](docs/guides/visual-studio.md) and the extension [README](extensions/vs2026/README.md).

## Documentation and CHANGELOG

When behavior changes, update the matching docs:

| Change | Update |
|--------|--------|
| Schema / manifest field | [configuration.md](docs/reference/configuration.md) |
| CLI command / flag | [cli.md](docs/reference/cli.md) |
| Feature status / limitation | [capabilities.md](docs/reference/capabilities.md) |
| Architecture / maintainer note | [PMAP.md](PMAP.md) |
| Completed roadmap item | [roadmap.md](docs/roadmap.md) |
| User-facing release note | [CHANGELOG.md](CHANGELOG.md) under `## [Unreleased]` |

User-facing changes (CLI, schema, behavior, breaking) must get a CHANGELOG entry before the PR is considered done.

## Commit messages

- Use short, imperative subject lines (e.g. `feat: add backup strategy`, `fix: handle empty include list`).
- Conventional prefixes are encouraged but not strictly enforced:
  `feat`, `fix`, `chore`, `docs`, `refactor`, `build`, `ci`.

## Pull requests

When opening a PR, include:

1. A short summary of the change and the motivation.
2. Steps you used to validate it manually (which fixture / environment / E2E scenario).
3. Any user-facing impact (CLI flag changes, schema changes, breaking behavior).
4. A note if the change touches an item in [Known limitations](docs/reference/capabilities.md) (or related SECURITY constraints).
5. Whether [CHANGELOG.md](CHANGELOG.md) was updated (and which bullet), when the change is user-facing.

## Reporting bugs

Please include:

- Pablo version (`pablo version`)
- Host OS / arch
- Target OS / arch (if remote)
- A minimal `pablo.yaml` that reproduces the issue
- Full CLI output

## Security issues

Do **not** report security issues via public GitHub issues. Follow the process in [SECURITY.md](SECURITY.md).
