## Pablo v2.0.1

Patch release: richer CLI motion on top of the v2.0.0 theme refresh.

### Changed

- **CLI theme + motion** — Compact cyan brand header (wordmark + version + accent rule), aligned status marks (`ok` / `fail` / `warn` / `info` / `run`), lighter section rules, and a single-line result footer.
- **Spinners & progress bars** — Long-running steps use braille spinners and pulsing progress bars (SSH connect, file filter, local copy, remote tar/legacy transfer, checksum verify, remote git/compose). Animations stay off for non-TTY, `NO_COLOR`, `CI`, and `PABLO_PLAIN`.

### Fixed

- Include/exclude globs with Windows-style backslashes (e.g. `src\*.go`) now match on Linux/macOS hosts.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
