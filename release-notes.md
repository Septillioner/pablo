## Pablo v2.5.0

CLI UX polish: quieter defaults, richer verbose detail, named pipeline sections, blue-green slot clarity, and optional JSON run summaries.

### Added

- **`--quiet`** — Global flag (also `PABLO_QUIET`) that skips brand header and section chrome; keeps fail/warn lines and the final Result. Cannot combine with `--verbose`.
- **`--json-summary`** — Optional `run` flag: after Result, print one JSON object to stdout (`project`, `version`, `profile`/`env`/`type`/`mode`, optional `paths`, `duration_ms`, `ok`; sequences use `sequence` instead of per-step fields).
- **Deployment Info** — Shows `Type`, `Mode` (`local`|`remote`), and `Strategy` (static/binary) alongside project/profile/target.

### Changed

- **`--verbose`** — Beyond artifact paths: also logs build/hook command cwd (and detect `Capture` command). Help text updated. Env `PABLO_VERBOSE` is respected when the flag is unset.
- **Pipeline section titles** — Named sections (`Build`, `Pre-Deployment`, `Deployment`, `Post-Deployment`, `Slot Switch`) instead of `Phase 2`…`Phase 5` (no Phase 1 gap).
- **Blue-green slot logging** — Detect and switch print active/target paths (and `key` when it differs from `path`), plus cutover direction before running `switch_command`.
- **`run` / `uninstall` SilenceUsage** — Validation and deploy failures no longer dump Cobra USAGE; pipeline failures already shown via `ui.Log` are not reprinted on stderr.

### Fixed

- **Windows hooks UTF-8** — Local PowerShell hooks (`Execute` / `Capture`) set console input/output encoding to UTF-8 so Turkish and other non-ASCII native tool output (e.g. IIS `appcmd`) is less likely to mojibake in UTF-8 terminals.
- **Hook / backup chatter** — Removed raw `Executing hook…` printf (commands already use `ui.Log`); backup messages use `ui.Log`. Verbose mode can still show hook cwd.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
