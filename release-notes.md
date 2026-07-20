## Pablo v2.1.2

Patch release: bottom-safe StepRail footer so build and other subprocess logs cannot erase the progress rail.

### Changed

- **Step rail footer** — Interactive StepRail is bottom-anchored: logs scroll above a single footer line (erase + reprint), instead of a sticky header that cursor-ups through build output. Spinner/ProgressBar still own the live line (footer hidden while they run). Build, hooks, git, and compose subprocesses use `ui.WithExternalOutput` so streaming stdout cannot wipe the rail.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
