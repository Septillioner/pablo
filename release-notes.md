## Pablo v2.2.1

Dynamic shell completion for `run` targets and common manifest flags.

### Added

- **Dynamic shell completion** — `pablo run` completes `profile/env` targets and `sequence <name>`; `-p` / `--profile`, `-e` / `--env`, and `-f` / `--file` complete from the selected (or default `pablo.yaml`) manifest on `run`, `check`, `uninstall`, and `inspect`. Enable via `pablo completion bash|zsh|fish|powershell` (see CLI reference).

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
