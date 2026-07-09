## Pablo v1.5.0

### Added

- One-liner CLI installers — `install.sh` (macOS/Linux), `install.ps1` (Windows PowerShell), and `install.cmd` (Windows bootstrap).
- Installers download release binaries from GitHub, verify SHA-256 checksums, and install to system or user PATH (`PABLO_VERSION` pins a release tag).
- Visual Studio 2026 extension — LSP via `pablo lsp`, CodeLens **Run**, profile/environment gutter stripes, manifest commands, YAML snippets, and executable picker.
- `docs/guides/visual-studio.md` — install, build, debug, and feature overview for the VS extension.

### Changed

- README and installation docs — one-liner install as the recommended path; PowerShell uses temp-file execution instead of `irm | iex`.
- PowerShell installer — PATH shadowing warnings and post-install `pablo` command resolution checks.

### Removed

- Bundled `pablo-lsp` binary from the VS Code extension (uses `pablo lsp` from the CLI on PATH).

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | `pablo-darwin-amd64` |
| macOS (Apple Silicon) | `pablo-darwin-arm64` |
| Linux (amd64) | `pablo-linux-amd64` |
| Windows (amd64) | `pablo-windows-amd64.exe` |
| Windows (arm64) | `pablo-windows-arm64.exe` |
| VS Code extension | `pablo-1.5.0.vsix` |
| Visual Studio 2026 extension | `pablo-vs2026-1.5.0.vsix` |

Verify downloads with `checksums.txt` (SHA-256).
