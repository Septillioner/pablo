# Changelog

All notable changes to Pablo are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.5.6] - 2026-07-10

### Fixed

- `install.sh` — resolving the latest release no longer fails with `curl: (23)` under `set -o pipefail` when `awk` exits early on a `curl | awk` pipe.
- `install.sh` — set `DOWNLOADED_BINARY` before checksum verification; expand temp dir paths in `EXIT`/`RETURN` traps so `set -u` does not trip on locals.
- `install.sh` — checksum verification failed on Linux/macOS when `checksums.txt` had Windows (CRLF) line endings; the release build now writes it with LF endings, and the installer strips a trailing `\r` defensively.
- `release-new-version.bat` — `checksums.txt` could include stale artifacts left over from a previous release's `build/` directory (e.g. an old `checksums.txt`, retired VSIX versions), producing hashes that did not match the files actually uploaded; the build directory is now recreated from scratch each release, and checksums are computed only for the exact files uploaded.

## [1.5.5] - 2026-07-10

### Fixed

- On Windows, `*` in include patterns could match across path segments (e.g. `*.exe` incorrectly included nested executables).
- Visual Studio extension — Pablo toolbar Profile/Environment combos did not refresh after saving `pablo.yaml`; toolbar now re-inspects on save and when opening profile/environment dropdowns.

### Added

- Global `--verbose` flag — after the artifact count, lists each relative path selected for deploy.
- `**` globstar support in `include` / `exclude` patterns (e.g. `**/*`, `**/*.exe`).
- `pablo update` — when other processes are using the Pablo binary, list them and prompt to close before replacing the executable (interactive terminals).
- `deploy.strategy: rename-replace` — per-file artifact replacement with timestamped rename, success cleanup, and full rollback on failure (local and remote SSH).
- `test.sh` / `test.ps1` / `test.bat` — unified test runner for `unit`, `integration`, `e2e`, and `all` modes.
- Progressive docs examples ([docs/examples](docs/examples/README.md)) — easy-to-hard manifests from local file copy (no build) through SSH, Docker, and multi-profile.

### Changed

- Include/exclude patterns use gitignore-style semantics: patterns without `/` match basenames at any depth; `/*.ext` or `./*.ext` limits to the artifact root.
- Quick start and configuration docs lead with no-build `static` copy; `build` documented as optional for `static`.
- `test.sh` / `test.ps1` — scenario-focused output with section headers, PASS/FAIL lines, and a summary block (`all` mode).

### Removed

- `publish-self.sh`, `publish-self.ps1`, and `publish-self.bat` — use `install.sh` / `install.ps1` instead.

## [1.5.4] - 2026-07-10

### Added

- `pablo init --template` / `-t` — interactive wizard to choose a sample template type (`static`, `binary`, `docker`, `git-sync`); requires an interactive terminal.
- Visual Studio extension — **Pablo Run Deployment** tool window with profile/environment combos and a **Run Deployment** button; **Tools → Pablo: Run Deployment** opens this panel.
- Visual Studio extension — **Pablo** toolbar (**View → Toolbars → Pablo**) with cascading **Manifest** / **Profile** / **Environment** combos plus **Run**; manifests are discovered from the solution and open documents.

### Fixed

- Visual Studio extension — incorrect `AdornmentLayerDefinition` MEF export broke the text editor factory (`ITextEditorFactoryService2` missing) when the extension loaded.
- Visual Studio extension — LSP probe quoted `help` as a PowerShell string (`'help' lsp`), so valid 1.3+ binaries were reported as lacking LSP; Select Executable now probes the chosen binary and shows its version + LSP status.
- Visual Studio extension — Debug F5 did not deploy into the Experimental Instance (SDK-style build skipped VSIX deploy); Debug builds now package and deploy to Exp.
- Visual Studio extension — **Run Deployment** and related commands failed to detect an open `pablo.yaml` when focus left the editor; active manifest resolution now uses DTE, document frame, and last-known path tracking.
- Visual Studio extension — Run Deployment tool window used default WPF colors (unreadable on dark theme); labels and combos now use VS `EnvironmentColors`.
- Visual Studio extension — deploy tool window showed misleading "No profiles" when Pablo CLI was not configured; inspect errors and missing executable are shown inline.
- Visual Studio extension — terminal commands used PowerShell syntax inside `cmd.exe`; shell detection now launches the matching host with correct quoting.
- Visual Studio extension — `cmd /k` mangled multi-quoted paths (`ERROR_INVALID_NAME` / "filename, directory name, or volume label syntax is incorrect"); Run now uses `cmd /s /k` with a single wrapped command line.
- Visual Studio extension — `pablo` content type was exported as a custom class instead of `ContentTypeDefinition`, so manifests never got the Pablo content type and LSP/CodeLens did not activate.
- Visual Studio extension — manifest path resolution falls back to any open `pablo*.yaml` document when editor focus is lost.
- Visual Studio extension — deploy tool window ComboBoxes inherited ToolWindow text color on white chrome (white-on-white in dark theme); combos and buttons now use VS `ThemedDialog*` styles.
- Visual Studio extension — LSP `ActivateAsync` returned null when MEF loaded the language client before package init (no retry); binary resolution now uses settings/PATH without requiring `PabloPackage.Instance`, and LSP restarts after package initialization.
- Visual Studio extension — LSP client bound only to `pablo` content type; built-in YAML editor keeps `yaml`/`YAML`, so the client never activated for many manifests. Client now also binds to inbox YAML content types.
- Visual Studio extension — Pablo toolbar DropDownCombos used wrong `InValue`/`OutValue` protocol (index vs label); combos appeared empty. Handlers now return current value + `string[]` lists and select by label.
- Visual Studio extension — toolbar Profile/Environment combos raced async inspect; lists stayed empty until reopen. Inspect now completes synchronously before get-list returns.
- Visual Studio extension — manifest discovery missed repo-root YAML in Open Folder; solution directory filesystem scan added.
- Visual Studio extension — Pablo toolbar Manifest combo showed duplicate `pablo.yaml` labels; items now use solution-relative paths with automatic disambiguation.
- Visual Studio extension — LSP `ActivateAsync` could throw on background thread during logging; activation wrapped in try/catch with PID/exit logging.
- LSP completion — items now include `insertText` for reliable insertion in Visual Studio (and VS Code).

## [1.5.1] - 2026-07-09

### Added

- `pablo update` — download the latest CLI binary for the current OS/arch from GitHub Releases, verify SHA-256, and replace the running executable (`--check`, `--version` / `PABLO_VERSION`).

### Fixed

- Windows installer (`install.ps1`) — rename-replace when overwriting a locked `pablo.exe` instead of failing on `Copy-Item`.
- PowerShell parse error in installer error strings (`$TargetPath:` → `${TargetPath}:`).
- `install.cmd` — download `install.ps1` to a temp file and run with `-File` instead of `iex (irm ...)`.

### Changed

- `install.sh` — install via temp file + `mv` for safer overwrite of an existing binary.
- `pablo version` — prints only the version string (removed the "Architecture: Modular Monolith" line).

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

[Unreleased]: https://github.com/septillioner/pablo/compare/v1.5.6...HEAD
[1.5.6]: https://github.com/septillioner/pablo/compare/v1.5.5...v1.5.6
[1.5.5]: https://github.com/septillioner/pablo/compare/v1.5.4...v1.5.5
[1.5.4]: https://github.com/septillioner/pablo/compare/v1.5.1...v1.5.4
[1.5.1]: https://github.com/septillioner/pablo/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/septillioner/pablo/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/septillioner/pablo/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/septillioner/pablo/compare/v1.0.46...v1.3.0
[1.0.46]: https://github.com/septillioner/pablo/releases/tag/v1.0.46
