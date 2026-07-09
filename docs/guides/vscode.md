# VS Code Extension

First-class editor support for `pablo.yaml` via the Pablo language server.

**Extension:** `septillioner.pablo` · **Requires:** Pablo CLI 1.3+ with `pablo lsp`

---

## Install

1. Install Pablo CLI — see [Installation](../getting-started/installation.md).
2. Install the extension from the VS Code marketplace or a release `.vsix`.
3. Open any `pablo.yaml` or `pablo*.yml` file.

---

## Features

| Feature | Source |
|---------|--------|
| Syntax highlighting | Extension grammar |
| Schema validation (squiggles) | `pablo lsp` → `pkg/validate` |
| Autocomplete (`Ctrl+Space`) | `pablo lsp` → `pkg/schema` |
| Hover docs | `pablo lsp` |
| Snippet templates (`pablo-tpl-*`) | Extension snippets |
| Check YAML / Init / Run commands | Extension → CLI |
| CodeLens **Run profile/env** | Extension → `pablo run` |
| Profile/env picker on Run | `pablo/listProfiles` or `inspect --json` |

---

## Binary resolution

The extension resolves the Pablo binary in this order:

1. **Selected executable** (via **Pablo: Select Executable**)
2. **`pablo.path`** setting
3. **PATH** — only when exactly one `pablo` is found

Workspace `build/pablo` is **not** auto-selected (may be stale). Point explicitly:

```json
{
  "pablo.path": "${workspaceFolder}/build/pablo.exe",
  "pablo.trace.server": "verbose"
}
```

Copy from `.vscode/settings.example.json` in the repo.

---

## Commands

| Command palette | Action |
|-----------------|--------|
| **Pablo: Check YAML** | `pablo check -f <file>` |
| **Pablo: Init Config** | `pablo init` |
| **Pablo: Run Deployment** | QuickPick profile + env → `pablo run` |
| **Pablo: Select Executable** | Choose CLI binary |

**CodeLens:** click **Run default/production** on an environment line to skip the picker.

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| "Pablo not found" on activate | **Select Pablo** or install from [Releases](https://github.com/septillioner/pablo/releases) |
| No autocomplete / hover | Binary must support `pablo lsp` (1.3+); old PATH install may lack LSP |
| Multiple `pablo` on PATH | Extension prompts to choose — pick the correct version |
| LSP crashes | Auto-restart disabled to prevent loops; check **Pablo Language Server** output |
| Snippets work but schema doesn't | Snippets ≠ LSP; schema needs running language server |

**Debug output:**

- **View → Output → Pablo Language Server** — `Using Pablo binary: ...`
- Set `"pablo.trace.server": "verbose"` for **Pablo LSP Trace**

**F5 development:**

1. `cd src && go build -o ../build/pablo.exe .`
2. Set `pablo.path` in `.vscode/settings.json`
3. **Run and Debug → Pablo: Run Extension**

More: [Troubleshooting](../troubleshooting.md) · [extension README](../../extensions/vscode-pablo/README.md)

---

## API surface

The extension uses:

- LSP stdio (`pablo lsp`)
- Custom request `pablo/listProfiles`
- CLI fallback: `pablo inspect --json`

Details: [API reference](../reference/api.md).
