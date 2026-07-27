## Pablo v2.2.3

VS Code Activity Bar Deploy view for picking a manifest, profile, and environment, then running deployment.

### Added

- **VS Code — Pablo Activity Bar** — Dedicated Pablo view container (not nested under Explorer) with Manifest / Profile / Environment pickers and **Run Deployment**. Discovers `pablo.yaml` and `pablo*.yaml` per workspace folder (exact `pablo.yaml` sorted first). **Pablo: Run Deployment** auto-selects the file when exactly one manifest exists; with multiple, uses the active editor (if a discovered manifest), then the Pablo view selection, then QuickPick.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
