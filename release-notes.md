## Pablo v2.1.0

Patch release: Adds `pablo update check` (standalone subcommand with optional `--json` output), deprecates `pablo update --check`, enables CLI update checks in VS Code and Visual Studio extensions on activation, and adds a manual **Pablo: Update** command in both extensions.


### Added

- `pablo update check` — check-only subcommand with optional `--json` (`current_version`, `latest_version`, `release_tag`, `update_available`); `pablo update --check` remains as a deprecated alias.
- VS Code and Visual Studio extensions check for CLI updates once on activation (`pablo update check --json`) and offer an Update action that runs `pablo update` after stopping the language server.
- **Pablo: Update** — VS Code (`pablo.update`) and Visual Studio (`Pablo.Update`) command to manually check for CLI updates and install when available (same stop-LSP → `pablo update` → restart flow as the activation check).
- **CLI step rail** — Interactive TTY runs show a sticky progress rail (`Validate → Build → Deploy → Post`, skipping phases that do not apply; sequences use target/env labels). Animations stay off for non-TTY, `NO_COLOR`, `CI`, and `PABLO_PLAIN`.


### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
