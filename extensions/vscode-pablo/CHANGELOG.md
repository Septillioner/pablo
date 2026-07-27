# Change Log

All notable changes to the "pablo" extension will be documented in this file.

Check [Keep a Changelog](http://keepachangelog.com/) for recommendations on how to structure this file.

## [Unreleased]

## [2.2.3] - 2026-07-27

### Added

- Pablo Activity Bar view container with Deploy webview: Manifest / Profile / Environment pickers and Run Deployment.
- Workspace manifest discovery (`pablo.yaml` and `pablo*.yaml`); **Pablo: Run Deployment** auto-selects a single manifest.
- On activation, check once for a newer Pablo CLI via `pablo update check --json` and offer an **Update** action.

## [1.3.0] - 2026-07-09

### Added

- LSP powered by `pablo lsp` (same CLI binary as deployment commands).
- `pablo.path` setting for a custom Pablo binary location.
- Shared schema validation and diagnostics aligned with `pablo check`.
- CLI commands from the palette: Check, Init, Run.

### Changed

- Snippet template version examples updated to `1.3.0`.
- Extension release aligned with Pablo CLI `1.3.0`.

## [0.0.1] - 2025

- Initial release with syntax highlighting, snippets, and templates.
