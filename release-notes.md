## Pablo v1.5.6

### Fixed

- `install.sh` — no longer fails with `curl: (23)` under `set -o pipefail` when resolving the latest release (`curl | awk` early exit).
- `install.sh` — set `DOWNLOADED_BINARY` before checksum verification; expand temp dir paths in `EXIT`/`RETURN` traps so `set -u` does not trip on locals.
- `install.sh` — checksum verification works when `checksums.txt` has Windows (CRLF) line endings; release build writes LF endings, installer strips trailing `\r`.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | `pablo-darwin-amd64` |
| macOS (Apple Silicon) | `pablo-darwin-arm64` |
| Linux (amd64) | `pablo-linux-amd64` |
| Windows (amd64) | `pablo-windows-amd64.exe` |
| Windows (arm64) | `pablo-windows-arm64.exe` |
| VS Code extension | `pablo-1.5.6.vsix` |
| Visual Studio 2026 extension | `pablo-vs2026-1.5.6.vsix` |

Verify downloads with `checksums.txt` (SHA-256).
