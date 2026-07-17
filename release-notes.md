## Pablo v2.0.0

Schema v2 is a breaking cut: one path per concept, no legacy aliases.

### Changed

- **Schema v2** — Manifests use Profile (what), Environment (where), and Deploy (how). All deploy config lives under `profiles`; inheritance is limited to `variables`, `env_file`, and `build`.
- **`deploy.remote` → `deploy.transfer`** — Values unchanged: `tar` (default) or `legacy`.
- **`remote` block** — Presence implies SSH (`remote.method` removed). Required: `host`, `credential`.
- **Artifacts** — Static/binary types require explicit `deploy.source` (`dir`, `include`, `exclude`) per environment.
- **Docs & scripts** — Public docs rewritten for schema v2; install/build/test runners live under `scripts/` (one-liner URLs use `.../master/scripts/install.*`).

### Added

- Unknown YAML keys are validation errors (`pablo check` / LSP).
- Stricter type gates: required/forbidden fields by profile `type`; empty `environments` is an error; `remote.credential` must resolve.

### Removed

- Legacy fields and layouts: `output_dir`, `hooks`, `pipeline`, `deploy.service`, `deploy.strategy: blue-green`, `remote.method`, `deploy.ssh` / flat env `target_path` / `strategy`, `docker.command`, sequence `profile.env` form, silent default SSH credential, and top-level `type` / `environments` auto-wrap. Use `deploy.pre_commands` / `deploy.post_commands` instead of hooks; reference credentials by name at the root.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |
| VS Code extension | pablo-2.0.0.vsix |
| Visual Studio 2026 extension | pablo-vs2026-2.0.0.vsix |

Verify downloads with checksums.txt (SHA-256).
