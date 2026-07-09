# Changelog

All notable changes to Pablo are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Docker-based E2E integration tests under `tests/e2e/` (Ubuntu SSH target, `go test -tags=integration`).
- Go unit tests for priority packages: `filter`, `pathutil`, `config/loader`, `template`, `deployer`, `health`, `hooks`, and `system` (`cd src && go test ./...`).
- `pkg/pathutil` — `JoinRemote` / `DirRemote` for POSIX remote paths (Windows host → Linux target).
- Remote SSH `docker` deploy — git sync, env file, and `docker compose up -d` over SSH.
- Linux system-scope PATH registration via `/etc/profile.d/pablo.sh`.
- `goals.md` — eksik özellikler ve hedef backlog (P0–P5 öncelik matrisi).
- `goals.md` — P2 altında genişletilmiş Go unit test hedefleri (paket önceliği, kapsam, CI bağlantısı).
- Public-facing project metadata: `LICENSE` (MIT), `CONTRIBUTING.md`, `SECURITY.md`, `RELEASING.md`, and this changelog.
- README sections for prerequisites, install from release, install from source, and self-deploy.
- `.gitignore` coverage for LSP build outputs, VS Code extension `dist/` / `out/` / `*.vsix`, and Go coverage files.

### Changed
- Windows `RemovePath` during `pablo uninstall` — removes Pablo PATH entries via PowerShell (User/Machine scope).
- Pipeline remote path building uses `pathutil` instead of `filepath.Join` / `filepath.Dir`.
- `README.md` restructured for first-time external users; release-binary install path documented.

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

[Unreleased]: https://github.com/septillioner/pablo/compare/v1.0.46...HEAD
[1.0.46]: https://github.com/septillioner/pablo/releases/tag/v1.0.46
