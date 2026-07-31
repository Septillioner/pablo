## Pablo v2.3.0

Blue-green slot deploy for static and binary profiles: detect the idle slot, write artifacts there, then run your switch command.

### Added

- **`deploy.blue_green`** — Slot-based deploy for `static` / `binary` (local + SSH): `detect_command` selects the idle slot, artifacts write there, then `switch_command` (global or per-slot) cuts traffic. See [Blue-green guide](docs/guides/blue-green.md).

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
