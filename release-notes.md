## Pablo v2.1.3

Patch release: remove StepRail and restore the standard Section/Log/Spinner/ProgressBar CLI UX.

### Removed

- **StepRail** — Sticky/footer step rail (`Validate → Build → Deploy → Post` and sequence labels), pulse animation, cursor-up footer erase/reprint, and `ui.WithExternalOutput` scaffolding are gone.

### Changed

- **CLI UX** — Interactive runs use the prior standard chrome only: `Header` / `Section`, marked `Log` lines, `Spinner`, `ProgressBar` / `FileProgress`, and `Result`. Build, hooks, git, and compose subprocesses stream stdout/stderr directly again.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
