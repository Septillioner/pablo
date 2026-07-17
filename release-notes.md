## Pablo v1.7.1

### Added

- Docker redeploy precheck: when a Compose stack is already running, Pablo stops it (`compose down`, no `-v`) before git sync, then brings it back up. Controlled by `deploy.docker.stop_before_sync` (default `true`).

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |
| VS Code extension | pablo-1.7.1.vsix |
| Visual Studio 2026 extension | pablo-vs2026-1.7.1.vsix |

Verify downloads with checksums.txt (SHA-256).
