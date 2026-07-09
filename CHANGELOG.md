# Changelog

All notable changes to Pablo are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.5.1] - 2026-07-09

### Added

- `pablo update` — download the latest CLI binary for the current OS/arch from GitHub Releases, verify SHA-256, and replace the running executable (`--check`, `--version` / `PABLO_VERSION`).

### Fixed

- Windows installer (`install.ps1`) — rename-replace when overwriting a locked `pablo.exe` instead of failing on `Copy-Item`.
- PowerShell parse error in installer error strings (`$TargetPath:` → `${TargetPath}:`).
- `install.cmd` — download `install.ps1` to a temp file and run with `-File` instead of `iex (irm ...)`.

### Changed

- `install.sh` — install via temp file + `mv` for safer overwrite of an existing binary.

## [1.5.0] - 2026-07-09

### Added

- One-liner CLI installers — `install.sh` (macOS/Linux), `install.ps1` (Windows PowerShell), and `install.cmd` (Windows bootstrap).
- Installers download release binaries from GitHub, verify SHA-256 checksums, and install to system or user PATH (`PABLO_VERSION` pins a release tag).
- Visual Studio 2026 extension (`extensions/vs2026/`) — LSP via `pablo lsp`, CodeLens **Run**, profile/environment gutter stripes, manifest commands, YAML snippets, and executable picker.
- `extensions/vs2026/build-vs2026.bat` — MSBuild script that produces `Pablo.VisualStudio.vsix`.
- `docs/guides/visual-studio.md` — install, build, debug, and feature overview for the VS extension.
- `release-new-version.bat` — builds and publishes `pablo-vs2026-<version>.vsix` alongside CLI binaries and the VS Code VSIX.

### Changed

- README and `docs/getting-started/installation.md` — one-liner install as the recommended path; PowerShell uses temp-file execution instead of `irm | iex`.
- PowerShell installer — PATH shadowing warnings and post-install `pablo` command resolution checks.
- `.gitignore` — Visual Studio extension build outputs (`extensions/vs2026/`).

### Removed

- Bundled `pablo-lsp` binary from `extensions/vscode-pablo/` (extension relies on `pablo lsp` from the CLI on PATH).

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

[Unreleased]: https://github.com/septillioner/pablo/compare/v1.5.1...HEAD
[1.5.1]: https://github.com/septillioner/pablo/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/septillioner/pablo/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/septillioner/pablo/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/septillioner/pablo/compare/v1.0.46...v1.3.0
[1.0.46]: https://github.com/septillioner/pablo/releases/tag/v1.0.46
