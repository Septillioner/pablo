## Pablo v2.0.2

Patch release: Adds `pablo update check` (standalone subcommand with optional `--json` output), deprecates `pablo update --check`, and enables CLI update checks in VS Code and Visual Studio extensions on activation with easy update actions.


### Added

- `pablo update check` — check-only subcommand with optional `--json` (`current_version`, `latest_version`, `release_tag`, `update_available`); `pablo update --check` remains as a deprecated alias.
- VS Code and Visual Studio extensions check for CLI updates once on activation (`pablo update check --json`) and offer an Update action that runs `pablo update` after stopping the language server.


### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
