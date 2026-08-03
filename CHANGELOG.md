# Changelog

All notable changes to Pablo are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.6.0] - 2026-08-03

### Added

- **`deploy.pre_commands` / `post_commands` `cwd`** — Optional per-command working directory: `project` (manifest dir, local only) or `target` (`target_path` / blue-green idle slot). Entries may still be plain strings. See [Pre/Post commands](docs/reference/configuration.md#prepost-commands).

### Changed

- **Local pre/post default cwd** — Without blue-green, omitted `cwd` uses the manifest directory (was the process cwd). Remote and blue-green defaults remain the deploy/slot path.

## [2.5.0] - 2026-08-03

### Added

- **`--quiet`** — Global flag (also `PABLO_QUIET`) that skips brand header and section chrome; keeps fail/warn lines and the final Result. Cannot combine with `--verbose`.
- **`--json-summary`** — Optional `run` flag: after Result, print one JSON object to stdout (`project`, `version`, `profile`/`env`/`type`/`mode`, optional `paths`, `duration_ms`, `ok`; sequences use `sequence` instead of per-step fields).
- **Deployment Info** — Shows `Type`, `Mode` (`local`|`remote`), and `Strategy` (static/binary) alongside project/profile/target.

### Changed

- **`--verbose`** — Beyond artifact paths: also logs build/hook command cwd (and detect `Capture` command). Help text updated. Env `PABLO_VERBOSE` is respected when the flag is unset.
- **Pipeline section titles** — Named sections (`Build`, `Pre-Deployment`, `Deployment`, `Post-Deployment`, `Slot Switch`) instead of `Phase 2`…`Phase 5` (no Phase 1 gap).
- **Blue-green slot logging** — Detect and switch print active/target paths (and `key` when it differs from `path`), plus cutover direction before running `switch_command`.
- **`run` / `uninstall` SilenceUsage** — Validation and deploy failures no longer dump Cobra USAGE; pipeline failures already shown via `ui.Log` are not reprinted on stderr.

### Fixed

- **Windows hooks UTF-8** — Local PowerShell hooks (`Execute` / `Capture`) set console input/output encoding to UTF-8 so Turkish and other non-ASCII native tool output (e.g. IIS `appcmd`) is less likely to mojibake in UTF-8 terminals.
- **Hook / backup chatter** — Removed raw `Executing hook…` printf (commands already use `ui.Log`); backup messages use `ui.Log`. Verbose mode can still show hook cwd.

## [2.4.1] - 2026-08-03

### Fixed

- **Blue-green local `switch_command` cwd** — Runs with cwd = manifest directory (same as `detect_command`), so relative script paths resolve against the project.

## [2.4.0] - 2026-07-31

### Added

- **`deploy.blue_green.slots[].key`** — Optional detect match value when `detect_command` stdout differs from the write `path` (defaults to `path`). See [Blue-green guide](docs/guides/blue-green.md).

### Fixed

- **Blue-green detect mismatch error** — Expected slot keys are quoted individually so literal backslashes and other special characters are visible in the failure message.

## [2.3.0] - 2026-07-31

### Added

- **`deploy.blue_green`** — Slot-based deploy for `static` / `binary` (local + SSH): `detect_command` selects the idle slot, artifacts write there, then `switch_command` (global or per-slot) cuts traffic. See [Blue-green guide](docs/guides/blue-green.md).

## [2.2.3] - 2026-07-27

### Added

- **VS Code — Pablo Activity Bar** — Dedicated Pablo view container (not nested under Explorer) with Manifest / Profile / Environment pickers and **Run Deployment**. Discovers `pablo.yaml` and `pablo*.yaml` per workspace folder (exact `pablo.yaml` sorted first). **Pablo: Run Deployment** auto-selects the file when exactly one manifest exists; with multiple, uses the active editor (if a discovered manifest), then the Pablo view selection, then QuickPick.

## [2.2.2] - 2026-07-24

### Fixed

- **LSP / validation — tab indentation** — Schema path resolution (`GetYAMLPath`) now treats leading tabs as indent (previously only spaces), so completion and hover work while editing tab-indented YAML. `pablo lsp` also triggers completion on Tab. Validation reports a clear error when indentation uses tabs (`YAML indentation cannot use tab characters; use spaces`) instead of only the generic YAML parse failure.

## [2.2.1] - 2026-07-23

### Added

- **Dynamic shell completion** — `pablo run` completes `profile/env` targets and `sequence <name>`; `-p` / `--profile`, `-e` / `--env`, and `-f` / `--file` complete from the selected (or default `pablo.yaml`) manifest on `run`, `check`, `uninstall`, and `inspect`. Enable via `pablo completion bash|zsh|fish|powershell` (see [CLI reference](docs/reference/cli.md#shell-completion)).

## [2.2.0] - 2026-07-22

### Changed

- **Env-first variables + pre-build `build.env_file`** — Environment `variables` (profile→env merge) are the canonical map. When `build.env_file` is set, Pablo writes that map under `build.path` before `build.command` and injects it into the build process env. Optional `build.variables` overlay build-only keys. Deploy `env_file` under `target_path` is unchanged.
- **Docs — variables / env files** — Recommend Vite/`VITE_*` maps under environment `variables` with `build.env_file: .env.production`; clarified write-only dotenv generation, empty-map skip, relative vs absolute paths; Schema v2 examples in [configuration.md](docs/reference/configuration.md) and [Examples #2](docs/examples/README.md#2-local-static--build).

## [2.1.3] - 2026-07-22

### Removed

- **StepRail** — Sticky/footer step rail (`Validate → Build → Deploy → Post` and sequence labels), pulse animation, cursor-up footer erase/reprint, and `ui.WithExternalOutput` scaffolding are gone.

### Changed

- **CLI UX** — Interactive runs use the prior standard chrome only: `Header` / `Section`, marked `Log` lines, `Spinner`, `ProgressBar` / `FileProgress`, and `Result`. Build, hooks, git, and compose subprocesses stream stdout/stderr directly again.

## [2.1.2] - 2026-07-20

### Changed

- **Step rail footer** — Interactive StepRail is bottom-anchored: logs scroll above a single footer line (erase + reprint), instead of a sticky header that cursor-ups through build output. Spinner/ProgressBar still own the live line (footer hidden while they run). Build, hooks, git, and compose subprocesses use `ui.WithExternalOutput` so streaming stdout cannot wipe the rail.

## [2.1.1]

### Fixed

- **Step rail vs spinner** — Sticky StepRail pulse no longer races Spinner/ProgressBar `\r` redraws (common garble in Windows PowerShell). Live-line chrome is serialized; rail pulse pauses while a spinner or incomplete progress bar owns the line, and phase updates repaint the spinner after the rail redraw.

## [2.1.0]

### Added

- **Pablo: Update** — VS Code (`pablo.update`) and Visual Studio (`Pablo.Update`) command to manually check for CLI updates and install when available (same stop-LSP → `pablo update` → restart flow as the activation check).
- **CLI step rail** — Interactive TTY runs show a sticky progress rail (`Validate → Build → Deploy → Post`, skipping phases that do not apply; sequences use target/env labels). Animations stay off for non-TTY, `NO_COLOR`, `CI`, and `PABLO_PLAIN`.

## [2.0.2]

### Added

- `pablo update check` — check-only subcommand with optional `--json` (`current_version`, `latest_version`, `release_tag`, `update_available`); `pablo update --check` remains as a deprecated alias.
- VS Code and Visual Studio extensions check for CLI updates once on activation (`pablo update check --json`) and offer an Update action that runs `pablo update` after stopping the language server.

## [2.0.1] - 2026-07-17

### Changed

- **CLI theme + motion** — Compact cyan brand header (wordmark + version + accent rule), aligned status marks (`ok` / `fail` / `warn` / `info` / `run`), lighter section rules, and a single-line result footer. Long-running steps now use braille spinners and pulsing progress bars (SSH connect, file filter, local copy, remote tar/legacy transfer, checksum verify, remote git/compose). Animations stay off for non-TTY, `NO_COLOR`, `CI`, and `PABLO_PLAIN`.

### Fixed

- Include/exclude globs with Windows-style backslashes (e.g. `src\*.go`) now match on Linux/macOS hosts.

## [2.0.0]

### Changed

- **Repo scripts under `scripts/`** — Install, build, and test runners moved from the repo root to `scripts/` (`install.sh` / `install.ps1` / `install.cmd`, `build.sh`, `test.sh` / `test.ps1` / `test.bat`). One-liner install URLs now use `.../master/scripts/install.*`.
- **Docs rewrite (schema v2)** — Public documentation under `docs/` rewritten for the current manifest model only, with a sixteen-scenario [Examples](docs/examples/README.md) catalog and fluent guides/reference.
- **Schema v2 (breaking)** — Manifest shape simplified around three nouns: Profile (what), Environment (where), Deploy (how). All deploy config lives under `profiles`; profile inheritance is limited to `variables`, `env_file`, and `build`.
- **`deploy.remote` renamed to `deploy.transfer`** — Values unchanged: `tar` (default) or `legacy`.
- **`remote` block** — Presence implies SSH; removed `remote.method`. Required fields: `host`, `credential`.
- **Artifacts** — Static/binary types require explicit `deploy.source` (`dir`, `include`, `exclude`) on each environment; no profile-level `output_dir` or source inheritance.

### Added

- Unknown YAML keys are validation errors (`pablo check` / LSP).
- Stricter type gates: required/forbidden fields by profile `type`; empty `environments` is an error; `remote.credential` is required and must resolve.

### Removed

- **Schema v2 (breaking)** — `output_dir`, `hooks`, `pipeline`, `deploy.service`, `deploy.strategy: blue-green`, `remote.method`, `deploy.ssh` / flat env `target_path` / `strategy`, `docker.command`, sequence `profile.env` form (only `profile/env`), silent default SSH credential invent, and legacy top-level `type` / `environments` auto-wrap. Use `deploy.pre_commands` / `deploy.post_commands` instead of hooks; reference credentials by name at the root.

## [1.7.3] - 2026-07-17

### Fixed

- Remote `docker compose` failures now show the remote command stdout/stderr instead of only exit status 1.

## [1.7.2] - 2026-07-17

### Added

- `deploy.verify_checksum` — optional post-transfer SHA-256 verification for remote static/binary deploys (default `false`; uses remote `sha256sum -c` over stdin).

### Fixed

- Remote tar/SCP paths from a Windows host no longer embed backslashes in Linux filenames (nested dirs like `assets/` extract correctly).

## [1.7.1] - 2026-07-17

### Added

- Docker redeploy precheck: when a Compose stack is already running, Pablo stops it (`compose down`, no `-v`) before git sync, then brings it back up. Controlled by `deploy.docker.stop_before_sync` (default `true`).

## [1.5.63] - 2026-07-13

### Added

- SSH host key verification against OpenSSH `known_hosts` (enabled by default); optional `remote.trust_on_first_use: on` to record unknown keys on first connect.
- README "Why Pablo?" section and VHS demo tape (`docs/assets/demo.tape`) for terminal GIF recording.

### Changed

- SSH connections no longer skip host key checks by default. Existing manifests keep working when the host is already in `known_hosts`; otherwise add the key or set `remote.host_key_verification: off` (emits a warning).

## [1.5.62] - 2026-07-10

### Added

- `sequences` in manifest root — named ordered lists of `profile/env` targets; run with `pablo run sequence <name>` (list order is execution order; stops on first failure).

## [1.5.61] - 2026-07-10

### Fixed

- fix install.sh checksum CRLF matching

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
- `builder.Service` is unused; pipeline runs builds inline.
- VS Code snippets hardcode an older version string instead of reading `src/VERSION`.

[Unreleased]: https://github.com/septillioner/pablo/compare/v2.3.0...HEAD
[2.3.0]: https://github.com/septillioner/pablo/compare/v2.2.3...v2.3.0
[2.2.3]: https://github.com/septillioner/pablo/compare/v2.2.2...v2.2.3
[2.2.2]: https://github.com/septillioner/pablo/compare/v2.2.1...v2.2.2
[2.2.1]: https://github.com/septillioner/pablo/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/septillioner/pablo/compare/v2.1.3...v2.2.0
[1.7.3]: https://github.com/septillioner/pablo/compare/v1.7.2...v1.7.3
[1.7.2]: https://github.com/septillioner/pablo/compare/v1.7.1...v1.7.2
[1.7.1]: https://github.com/septillioner/pablo/compare/v1.5.63...v1.7.1
[1.5.63]: https://github.com/septillioner/pablo/compare/v1.5.62...v1.5.63
[1.5.62]: https://github.com/septillioner/pablo/compare/v1.5.61...v1.5.62
[1.5.61]: https://github.com/septillioner/pablo/compare/v1.5.60...v1.5.61
[1.5.7]: https://github.com/septillioner/pablo/compare/v1.5.6...v1.5.7
[1.5.6]: https://github.com/septillioner/pablo/compare/v1.5.5...v1.5.6
[1.5.5]: https://github.com/septillioner/pablo/compare/v1.5.4...v1.5.5
[1.5.4]: https://github.com/septillioner/pablo/compare/v1.5.1...v1.5.4
[1.5.1]: https://github.com/septillioner/pablo/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/septillioner/pablo/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/septillioner/pablo/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/septillioner/pablo/compare/v1.0.46...v1.3.0
[1.0.46]: https://github.com/septillioner/pablo/releases/tag/v1.0.46
