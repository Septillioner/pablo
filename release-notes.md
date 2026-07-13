## Pablo v1.5.63

### Added

- SSH host key verification against OpenSSH `known_hosts` (enabled by default); optional `remote.trust_on_first_use: on` to record unknown keys on first connect.
- README "Why Pablo?" section and demo GIF at the top of the project README.

### Changed

- SSH connections no longer skip host key checks by default. Existing manifests keep working when the host is already in `known_hosts`; otherwise add the key or set `remote.host_key_verification: off` (emits a warning).

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |
| VS Code extension | pablo-1.5.63.vsix |
| Visual Studio 2026 extension | pablo-vs2026-1.5.63.vsix |

Verify downloads with checksums.txt (SHA-256).