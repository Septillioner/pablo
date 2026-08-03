## Pablo v2.6.0

Per-command working directory for deploy pre/post hooks: run scripts from the project root or the deploy target without hard-coding absolute paths.

### Added

- **`deploy.pre_commands` / `post_commands` `cwd`** — Optional per-command working directory: `project` (manifest dir, local only) or `target` (`target_path` / blue-green idle slot). Entries may still be plain strings. See [Pre/Post commands](docs/reference/configuration.md#prepost-commands).

### Changed

- **Local pre/post default cwd** — Without blue-green, omitted `cwd` uses the manifest directory (was the process cwd). Remote and blue-green defaults remain the deploy/slot path.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
