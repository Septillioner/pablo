## Pablo v1.7.2

### Added

- `deploy.verify_checksum` — optional post-transfer SHA-256 verification for remote static/binary deploys (default `false`).

### Fixed

- Remote tar/SCP paths from a Windows host no longer embed backslashes in Linux filenames (nested directories extract correctly).

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |
| VS Code extension | pablo-1.7.2.vsix |
| Visual Studio 2026 extension | pablo-vs2026-1.7.2.vsix |

Verify downloads with checksums.txt (SHA-256).
