# Contributing to Pablo

Thanks for your interest in contributing. This document describes the practical workflow and conventions used in this repository.

## Ground rules

- Open an issue before starting non-trivial work to align on direction.
- Keep changes focused. Prefer small, reviewable PRs over large mixed ones.
- Match the existing code style and package layout (see `PROJECT STRUCTURE` in [README.md](README.md)).
- Do not introduce dependencies casually. Justify any new direct dependency in the PR description.

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

### Validate manually

There are no Go unit tests. Validate changes by running pipelines against the YAML fixtures under `tests/`:

```bash
cd tests/agnostic/local-deploy
go run ../../../src/main.go run -e production
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
