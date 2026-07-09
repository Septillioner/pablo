# Contributing to Pablo

Thanks for your interest in contributing. This document describes the practical workflow and conventions used in this repository.

## Ground rules

- Open an issue before starting non-trivial work to align on direction.
- Keep changes focused. Prefer small, reviewable PRs over large mixed ones.
- Match the existing code style and package layout (see `PROJECT STRUCTURE` in [README.md](README.md)).
- Do not introduce dependencies casually. Justify any new direct dependency in the PR description.
- Check [docs/roadmap.md](docs/roadmap.md) for the current backlog; update checklist items when your PR completes a tracked goal.

## Project docs

| Document | Purpose |
|----------|---------|
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/roadmap.md](docs/roadmap.md) | Public roadmap — shipped features and planned work |
| [docs/reference/cli.md](docs/reference/cli.md) | CLI command reference |
| [docs/reference/capabilities.md](docs/reference/capabilities.md) | Supported types, strategies, limitations |
| [docs/reference/configuration.md](docs/reference/configuration.md) | `pablo.yaml` field reference |
| [PMAP.md](PMAP.md) | Internal architecture map (maintainers and agents; not user-facing) |
| [tests/TEST_SPEC.md](tests/TEST_SPEC.md) | Test catalog and coverage status |

## Project setup

### Prerequisites

- Go 1.25.5 or newer
- Git
- (Optional) Node.js 20+ if you plan to work on `extensions/vscode-pablo`

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
go run main.go run -e <env> -p <profile> -f ../pablo.yaml
```

### Validate changes

Run Go unit tests from `src/`:

```bash
cd src
go test ./...
```

Also validate end-to-end behavior against the YAML fixtures under `tests/`:

```bash
cd tests/agnostic/local-deploy
go run ../../../src/main.go run -e production
```

For remote SSH / docker scenarios, run the Docker-based E2E suite (requires Docker):

```bash
cd tests/e2e
go test -tags=integration -v -timeout 10m ./...
```

## Code conventions

- Go 1.25.5; direct dependencies kept minimal (`cobra`, `fatih/color`, `yaml.v3`, `x/crypto`).
- Concrete struct dependencies, no interfaces / mocking layer.
- DI is wired in `src/main.go`.
- Logging goes through `pkg/ui` (`ui.Log(mark, msg)`):
  `+` success, `-` error, `!` warn, `*` info, `>` action.
- Build commands: `sh -c` on Unix, `cmd /C` on Windows.
- Hooks: `powershell` on Windows, `sh` elsewhere.
- Template variables use `{{KEY}}` and are only expanded for config-like file extensions.
- Avoid emojis in code, comments, logs, and commit messages.
- Comments should explain non-obvious intent only. Do not narrate the code.

## Commit messages

- Use short, imperative subject lines (e.g. `feat: add backup strategy`, `fix: handle empty include list`).
- Conventional prefixes are encouraged but not strictly enforced:
  `feat`, `fix`, `chore`, `docs`, `refactor`, `build`, `ci`.

## Pull requests

When opening a PR, include:

1. A short summary of the change and the motivation.
2. Steps you used to validate it manually (which fixture / environment).
3. Any user-facing impact (CLI flag changes, schema changes, breaking behavior).
4. A note if the change touches one of the items listed under "Known Limitations" in the README.

## Reporting bugs

Please include:

- Pablo version (`pablo version`)
- Host OS / arch
- Target OS / arch (if remote)
- A minimal `pablo.yaml` that reproduces the issue
- Full CLI output

## Security issues

Do **not** report security issues via public GitHub issues. Follow the process in [SECURITY.md](SECURITY.md).
