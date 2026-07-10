## Pablo v1.5.5

### Added

- Global `--verbose` flag — lists each relative path selected for deploy after the artifact count.
- `**` globstar support in `include` / `exclude` patterns (e.g. `**/*`, `**/*.exe`).
- `pablo update` — when other processes use the Pablo binary, list them and prompt to close before replacing (interactive terminals).
- `deploy.strategy: rename-replace` — per-file replacement with timestamped rename, success cleanup, and full rollback (local and remote SSH).
- `test.sh` / `test.ps1` / `test.bat` — unified test runner for `unit`, `integration`, `e2e`, and `all`.
- Progressive docs examples — easy-to-hard manifests from local copy through SSH, Docker, and multi-profile.

### Changed

- Include/exclude patterns use gitignore-style semantics: patterns without `/` match basenames at any depth; `/*.ext` or `./*.ext` limits to the artifact root.
- Quick start and configuration docs lead with no-build `static` copy; `build` is optional for `static`.
- Test runners — scenario-focused output with section headers, PASS/FAIL lines, and a summary block.

### Fixed

- On Windows, `*` in include patterns could match across path segments (e.g. `*.exe` incorrectly included nested executables).
- Visual Studio extension — Pablo toolbar Profile/Environment combos now refresh after saving `pablo.yaml` and when opening those dropdowns.

### Removed

- `publish-self.sh`, `publish-self.ps1`, and `publish-self.bat` — use `install.sh` / `install.ps1` instead.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | `pablo-darwin-amd64` |
| macOS (Apple Silicon) | `pablo-darwin-arm64` |
| Linux (amd64) | `pablo-linux-amd64` |
| Windows (amd64) | `pablo-windows-amd64.exe` |
| Windows (arm64) | `pablo-windows-arm64.exe` |
| VS Code extension | `pablo-1.5.5.vsix` |
| Visual Studio 2026 extension | `pablo-vs2026-1.5.5.vsix` |

Verify downloads with `checksums.txt` (SHA-256).
