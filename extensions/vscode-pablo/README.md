# Pablo for Visual Studio Code

Pablo is a powerful deployment helper for managing multi-environment builds and deployments. This extension provides first-class support for `pablo.yaml` configuration files.

## Features

- **LSP Powered Intelligence**: Real-time validation, diagnostics, and autocompletion via `pablo lsp` (same binary as the CLI).
- **Pablo Activity Bar**: Dedicated Pablo view (not under Explorer) with Manifest / Profile / Environment pickers and **Run Deployment**. Discovers `pablo.yaml` and `pablo*.yaml` across workspace folders.
- **CLI update check**: On activation, runs `pablo update check --json` once and offers an **Update** action when a newer CLI release is available.
- **Smart Autocompletion**: Context-aware suggestions for all Pablo configuration fields and enum values.
- **Hover Documentation**: Quick access to documentation and field descriptions directly within the editor.
- **Scaffold Templates**: Ready-to-use Schema v2 templates (type `pablo-` to browse):
    - `pablo-config`: Empty skeleton
    - `pablo-tpl-static-site`: Static site (overwrite)
    - `pablo-tpl-static-hotfix`: Static hotfix (rename-replace)
    - `pablo-tpl-go-service`: Go/binary service with post_commands
    - `pablo-tpl-compose`: Docker Compose (git + deploy.docker)
    - `pablo-tpl-php-app`: PHP git-sync with env_file + post_commands
    - `pablo-tpl-sequence`: Staging → production release sequence
    - `pablo-tpl-site-backup`: Static site with backup strategy
    - `pablo-tpl-clean-redeploy`: Clean redeploy (recreate)
    - `pablo-tpl-legacy-transfer`: Legacy transfer mode
    - `pablo-tpl-verified-transfer`: Checksum-verified transfer
- **Custom File Icon**: Beautiful logo integration for your `pablo.yaml` files.
- **CLI Integration**: Run Pablo commands (`check`, `init`, `run`) directly from the command palette. **Pablo: Run Deployment** auto-selects the manifest when only one exists; otherwise uses the active editor, Deploy view selection, or QuickPick, then prompts for profile and environment.
- **CodeLens Run**: Each environment line shows **$(play) Run** (hover for `profile/env`); runs `pablo run -f ... profile/env` without QuickPick.
- **Environment gutters**: Colored left borders at **profile indent** (one color per profile) and **environment indent** (separate hue-spread palette per env block, including nested lines).

## Usage

1. Open the **Pablo** icon in the Activity Bar → pick Manifest / Profile / Environment → **Run Deployment**.
2. Or open any `pablo.yaml` / `pablo*.yml` file and use Command Palette **Pablo: Run Deployment**, CodeLens, or the editor title Run action.
3. Start typing `pablo-` to see available templates.
4. Use `Ctrl+Space` for smart completions.
5. Hover over any key to see its description.

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
