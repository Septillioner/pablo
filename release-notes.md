## Pablo v1.4.0

### Added

- `pablo inspect` — list profiles and environments from a manifest (`--json` for machine-readable output).
- `pkg/target` — parse positional `profile/env` run targets shared by CLI and LSP.
- LSP CodeLens — `Run` on environment lines (`pablo.runWithArgs`).
- LSP custom request `pablo/listProfiles` for editor profile/environment pickers.
- VS Code extension: binary picker, inspect fallback, shell quoting helpers, and profile/environment gutter decorations.
- Public docs tree under `docs/` — getting started, guides, reference, development, FAQ, and troubleshooting.
- `docs/roadmap.md` (moved from `docs/goals.md`).

### Changed

- README slimmed down; install and usage detail live under `docs/`.
- `build.sh` accepts `BUILD_DIR` override for release artifact output.
- Local deploy prep errors for missing artifact / target directories are clearer.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | `pablo-darwin-amd64` |
| macOS (Apple Silicon) | `pablo-darwin-arm64` |
| Linux (amd64) | `pablo-linux-amd64` |
| Windows (amd64) | `pablo-windows-amd64.exe` |
| Windows (arm64) | `pablo-windows-arm64.exe` |
| VS Code extension | `pablo-1.4.0.vsix` |

Verify downloads with `checksums.txt` (SHA-256). Also on the Marketplace as `septillioner.pablo`.
