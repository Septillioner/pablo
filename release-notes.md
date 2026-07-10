## Pablo v1.5.4

### Added

- `pablo init --template` / `-t` — interactive wizard for sample template type (`static`, `binary`, `docker`, `git-sync`).
- Visual Studio extension — **Pablo Run Deployment** tool window (profile/environment combos + Run).
- Visual Studio extension — **Pablo** toolbar with Manifest / Profile / Environment combos + Run.

### Fixed

- Visual Studio extension — MEF adornment/content-type exports, LSP activation (YAML content types, package-init race), toolbar combo protocol, manifest discovery, dark-theme tool window, terminal quoting (`cmd /s /k`), and related Run/LSP reliability fixes.
- LSP completion — `insertText` for reliable insertion in Visual Studio and VS Code.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | `pablo-darwin-amd64` |
| macOS (Apple Silicon) | `pablo-darwin-arm64` |
| Linux (amd64) | `pablo-linux-amd64` |
| Windows (amd64) | `pablo-windows-amd64.exe` |
| Windows (arm64) | `pablo-windows-arm64.exe` |
| VS Code extension | `pablo-1.5.4.vsix` |
| Visual Studio 2026 extension | `pablo-vs2026-1.5.4.vsix` |

Verify downloads with `checksums.txt` (SHA-256).
