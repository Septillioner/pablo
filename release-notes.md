## Pablo v1.3.0

### Added

- Docker-based E2E integration tests under `tests/e2e/` (Ubuntu SSH target, `go test -tags=integration`).
- Go unit tests for priority packages: `filter`, `pathutil`, `config/loader`, `template`, `deployer`, `health`, `hooks`, and `system` (`cd src && go test ./...`).
- `pkg/pathutil` — `JoinRemote` / `DirRemote` for POSIX remote paths (Windows host → Linux target).
- Remote SSH `docker` deploy — git sync, env file, and `docker compose up -d` over SSH.
- Linux system-scope PATH registration via `/etc/profile.d/pablo.sh`.
- `goals.md` — public roadmap and feature backlog.
- Public-facing project metadata: `LICENSE` (MIT), `CONTRIBUTING.md`, `SECURITY.md`, `RELEASING.md`, and changelog.
- README sections for prerequisites, install from release, install from source, and self-deploy.
- `.gitignore` coverage for LSP build outputs, VS Code extension `dist/` / `out/` / `*.vsix`, and Go coverage files.
- `pablo lsp` subcommand — single-binary Language Server Protocol for VS Code and other editors.
- Shared manifest validation (`pkg/validate`) with line/column diagnostics in `pablo check`, `pablo run`, and LSP.

### Changed

- Windows `RemovePath` during `pablo uninstall` — removes Pablo PATH entries via PowerShell (User/Machine scope).
- Pipeline remote path building uses `pathutil` instead of `filepath.Join` / `filepath.Dir`.
- `README.md` restructured for first-time external users; release-binary install path documented.
- LSP version reports the same value as `pablo version` (from `src/VERSION`).
- VS Code extension spawns `pablo lsp` instead of a separate `pablo-lsp` binary.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | `pablo-darwin-amd64` |
| macOS (Apple Silicon) | `pablo-darwin-arm64` |
| Linux (amd64) | `pablo-linux-amd64` |
| Windows (amd64) | `pablo-windows-amd64.exe` |
| Windows (arm64) | `pablo-windows-arm64.exe` |

Verify downloads with `checksums.txt` (SHA-256).
