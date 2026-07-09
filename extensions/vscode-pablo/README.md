# Pablo for Visual Studio Code

Pablo is a powerful deployment helper for managing multi-environment builds and deployments. This extension provides first-class support for `pablo.yaml` configuration files.

## Features

- **LSP Powered Intelligence**: Real-time validation, diagnostics, and autocompletion via `pablo lsp` (same binary as the CLI).
- **Smart Autocompletion**: Context-aware suggestions for all Pablo configuration fields and enum values.
- **Hover Documentation**: Quick access to documentation and field descriptions directly within the editor.
- **Scaffold Templates**: Ready-to-use templates for common scenarios:
    - `pablo-tpl-static`: Static Website
    - `pablo-tpl-node-pm2`: Node.js with PM2
    - `pablo-tpl-go-systemd`: Go with Systemd
    - `pablo-tpl-docker`: Docker Compose
- **Custom File Icon**: Beautiful logo integration for your `pablo.yaml` files.
- **CLI Integration**: Run Pablo commands (`check`, `init`, `run`) directly from the command palette. **Run** prompts for profile and environment.
- **CodeLens Run**: Each environment line in `pablo.yaml` shows a **Run profile/env** link (e.g. `Run default/windows-local`) that runs deployment without QuickPick.

## Usage

1. Open any `pablo.yaml` or `pablo.yml` file.
2. Start typing `pablo-` to see available templates.
3. Use `Ctrl+Space` for smart completions.
4. Hover over any key to see its description.

## Requirements

- Pablo CLI **1.3+** with `pablo lsp` support.
- Auto-resolve order: **selected executable** → `pablo.path` → **PATH** (only when exactly one `pablo` is found).
- Workspace `build/pablo(.exe)` is **not** auto-selected (relative paths can point to old copies); use **Pablo: Select Executable** or set `pablo.path`.
- Multiple `pablo` binaries on PATH → extension prompts you to choose.

## Troubleshooting

### Autocomplete / hover not working

Schema completion and hover come from **`pablo lsp`**, not from snippets.

| Check | Action |
|-------|--------|
| **Pablo not found dialog** | On activate, if Pablo is missing or too old, choose **Select Pablo** or **Install Pablo**. |
| **Select executable** | Command Palette → **Pablo: Select Executable** — selected, `pablo.path`, workspace build, PATH, or browse. |
| Binary version | Listed in the picker; must support `pablo lsp` (1.3+). Entries marked `(no lsp)` will not start the language server. |
| Old PATH install | An older `pablo` on PATH without `lsp` triggers the dialog — pick `build/pablo.exe` or upgrade from [Releases](https://github.com/septillioner/pablo/releases). |
| Binary path | Or set `pablo.path` manually in settings. |
| Output logs | **View → Output** → **Pablo Language Server** — look for `Using Pablo binary: ...` |
| LSP trace | Set `"pablo.trace.server": "verbose"` and open **Pablo LSP Trace** output. |
| Language mode | Status bar should show **Pablo Configuration** or **YAML** for `pablo*.yaml` files. |
| Snippets vs LSP | `pablo-tpl-*` = snippets; `Ctrl+Space` on YAML keys = LSP schema completion. |

Example workspace settings (copy from repo `.vscode/settings.example.json`):

```json
{
  "pablo.path": "${workspaceFolder}/build/pablo.exe",
  "pablo.trace.server": "verbose"
}
```

### Debugging the extension (F5)

1. Build CLI: `cd src && go build -o ../build/pablo.exe .`
2. Copy `.vscode/settings.example.json` → `.vscode/settings.json` (or merge into user settings).
3. Open repo root in VS Code → **Run and Debug** → **Pablo: Run Extension** (F5).
4. In the **Extension Development Host** window, open `pablo.yaml` or `pablo-sepy.yaml`.
5. Set breakpoints in `extensions/vscode-pablo/src/extension.ts` if needed.
6. Watch **Pablo Language Server** and **Pablo LSP Trace** output channels.

## Release Notes

### 1.3.0

- LSP via `pablo lsp` with shared schema validation and diagnostics.
- **Pablo: Select Executable** — choose CLI binary (selected, setting, workspace build, PATH, browse).
- **Pablo not found** dialog: Cancel / Select Pablo / Install Pablo (GitHub Releases).
- Deterministic binary resolution; no silent fallback to unsupported PATH binaries.
- LSP auto-restart disabled on crash to prevent EPIPE restart loops.
- `pablo.path` setting for custom binary location.
- Snippet templates updated to version `1.3.0`.

### 0.0.1
- Initial release with full LSP support.
- Cross-platform support for Windows, macOS (Intel/Silicon), and Linux.
- Comprehensive set of snippets and templates.

---
**Enjoy efficient deployments with Pablo!**
