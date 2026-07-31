## Pablo v2.4.0

Optional per-slot detect keys for blue-green when the target names a slot differently than Pablo writes it, plus clearer detect mismatch errors.

### Added

- **`deploy.blue_green.slots[].key`** — Optional detect match value when `detect_command` stdout differs from the write `path` (defaults to `path`). See [Blue-green guide](docs/guides/blue-green.md).

### Fixed

- **Blue-green detect mismatch error** — Expected slot keys are quoted individually so literal backslashes and other special characters are visible in the failure message.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
