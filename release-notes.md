## Pablo v1.5.1

### Added

- `pablo update` — download the latest CLI binary for your OS from GitHub Releases, verify the checksum, and replace the running executable.
- `pablo update --check` — report whether a newer release is available without downloading.
- Optional pin via `--version` / `PABLO_VERSION`.

### Fixed

- Windows one-liner installer — safer overwrite when `pablo.exe` is locked; PowerShell parse error in error messages.
- `install.cmd` — temp-file download instead of `iex (irm ...)`.

### Changed

- macOS/Linux installer uses temp + `mv` when replacing an existing binary.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | `pablo-darwin-amd64` |
| macOS (Apple Silicon) | `pablo-darwin-arm64` |
| Linux (amd64) | `pablo-linux-amd64` |
| Windows (amd64) | `pablo-windows-amd64.exe` |
| Windows (arm64) | `pablo-windows-arm64.exe` |
| VS Code extension | `pablo-1.5.1.vsix` |
| Visual Studio 2026 extension | `pablo-vs2026-1.5.1.vsix` |

Verify downloads with `checksums.txt` (SHA-256).
