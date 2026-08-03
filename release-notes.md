## Pablo v2.4.1

Local blue-green `switch_command` now uses the same working directory as `detect_command`.

### Fixed

- **Blue-green local `switch_command` cwd** — Runs with cwd = manifest directory (same as `detect_command`), so relative script paths resolve against the project.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
