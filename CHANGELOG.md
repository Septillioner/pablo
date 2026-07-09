# Changelog

All notable changes to Pablo are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.4.0] - 2026-07-09

### Added

- `pablo inspect` — list profiles and environments from a manifest (`--json` for machine-readable output).
- `pkg/target` — parse positional `profile/env` run targets shared by CLI and LSP.
- LSP CodeLens — `Run` on environment lines (`pablo.runWithArgs`).
- LSP custom request `pablo/listProfiles` for editor profile/environment pickers.
- VS Code extension: binary picker (`executable.ts`), inspect fallback, shell quoting helpers, and profile/environment gutter decorations.
- Public docs tree under `docs/` — getting started, guides, reference, development, FAQ, and troubleshooting.
- `docs/roadmap.md` (moved from `docs/goals.md`).

### Changed

- README slimmed down; install and usage detail live under `docs/`.
- `build.sh` accepts `BUILD_DIR` override for release artifact output.
- Local deploy prep errors for missing artifact / target directories are clearer.

## [1.3.0] - 2026-07-09

### Added

- Docker-based E2E integration tests under `tests/e2e/` (Ubuntu SSH target, `go test -tags=integration`).
- Go unit tests for priority packages: `filter`, `pathutil`, `config/loader`, `template`, `deployer`, `health`, `hooks`, and `system` (`cd src && go test ./...`).
- `pkg/pathutil` — `JoinRemote` / `DirRemote` for POSIX remote paths (Windows host → Linux target).
- Remote SSH `docker` deploy — git sync, env file, and `docker compose up -d` over SSH.
- Linux system-scope PATH registration via `/etc/profile.d/pablo.sh`.
- `goals.md` — public roadmap and feature backlog.
- Public-facing project metadata: `LICENSE` (MIT), `CONTRIBUTING.md`, `SECURITY.md`, `RELEASING.md`, and this changelog.
- README sections for prerequisites, install from release, install from source, and self-deploy.
- `.gitignore` coverage for LSP build outputs, VS Code extension `dist/` / `out/` / `*.vsix`, and Go coverage files.
- `pablo lsp` subcommand — single-binary Language Server Protocol for VS Code and other editors.
- Shared manifest validation (`pkg/validate`) with line/column diagnostics in `pablo check`, `pablo run`, and LSP.
- VS Code extension v1.3.0 published to Marketplace (`septillioner.pablo`).

### Changed

- Windows `RemovePath` during `pablo uninstall` — removes Pablo PATH entries via PowerShell (User/Machine scope).
- Pipeline remote path building uses `pathutil` instead of `filepath.Join` / `filepath.Dir`.
- `README.md` restructured for first-time external users; release-binary install path documented.
- LSP version reports the same value as `pablo version` (from `src/VERSION`).
- VS Code extension spawns `pablo lsp` instead of a separate `pablo-lsp` binary.

## [1.0.46] - 2025

Initial public release baseline tracked in `src/VERSION`.

### Highlights

- Multi-profile, multi-environment YAML manifests with profile-to-environment inheritance.
- Deployment types: `static`, `binary`, `docker` (local), `git-sync`.
- Deploy strategies: `overwrite`, `backup`, `recreate` (`blue-green` is a stub).
- Local copy and remote SSH deploy (tar-streaming with SCP fallback).
- Glob-based artifact filtering with include / exclude patterns.
- Template variable substitution (`{{VAR}}`) for config-like files.
- Cross-platform PATH registration (Windows, macOS, Linux user scope).
- VS Code extension with LSP-backed completion, hover, and YAML validation.

### Known limitations

- No Go unit tests; only YAML fixtures under `tests/`.
- `blue-green` strategy not implemented (returns error).
- SSH host key verification disabled (`InsecureIgnoreHostKey`); see `SECURITY.md`.
- Windows `RemovePath` returns "not yet implemented" during `pablo uninstall`.
- `docker` deployment type has no remote SSH support.
- LSP validator only catches YAML syntax errors, not schema-level issues.
- `filepath.Join` may produce backslashes when a Windows host targets Linux.
- `builder.Service` is unused; pipeline runs builds inline.
- VS Code snippets hardcode an older version string instead of reading `src/VERSION`.

[Unreleased]: https://github.com/septillioner/pablo/compare/v1.4.0...HEAD
[1.4.0]: https://github.com/septillioner/pablo/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/septillioner/pablo/compare/v1.0.46...v1.3.0
[1.0.46]: https://github.com/septillioner/pablo/releases/tag/v1.0.46
