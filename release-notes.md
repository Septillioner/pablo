## Pablo v2.2.0

Environment-first variables and pre-build `build.env_file` for Schema v2 manifests.

### Changed

- **Env-first variables + pre-build `build.env_file`** — Environment `variables` (profile→env merge) are the canonical map. When `build.env_file` is set, Pablo writes that map under `build.path` before `build.command` and injects it into the build process env. Optional `build.variables` overlay build-only keys. Deploy `env_file` under `target_path` is unchanged.
- **Docs — variables / env files** — Recommend Vite/`VITE_*` maps under environment `variables` with `build.env_file: .env.production`; clarified write-only dotenv generation, empty-map skip, relative vs absolute paths; Schema v2 examples in configuration.md and Examples #2.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
