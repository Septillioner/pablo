# Visual Studio Extension

First-class editor support for `pablo.yaml` in Visual Studio 2026 Insiders and Visual Studio 2022 via the Pablo language server.

**Extension:** `Pablo.VisualStudio` VSIX · **Requires:** Pablo CLI 1.3+ with `pablo lsp` · **Targets:** VS 2022 (17.x) and VS 2026 Insiders (18.x/19.x manifest range)

---

## Install

1. Install Pablo CLI — see [Installation](../getting-started/installation.md).
2. Install the VSIX from [GitHub Releases](https://github.com/septillioner/pablo/releases) (`pablo-vs2026-<version>.vsix`).
3. Open any `pablo.yaml` or `pablo*.yml` file.

---

## Features

| Feature | Source |
|---------|--------|
| Syntax highlighting | YAML editor + `pablo` content type |
| Schema validation (squiggles) | `pablo lsp` → `pkg/validate` |
| Autocomplete | `pablo lsp` → `pkg/schema` |
| Hover docs | `pablo lsp` |
| Snippet templates (`pablo-tpl-*`) | Extension snippets |
| Check YAML / Init / Run commands | Extension → CLI |
| CodeLens **Run profile/env** | Extension → `pablo run` |
| Profile/env gutter stripes | Extension adornments |
| Profile/env picker on Run | `pablo/listProfiles` or `inspect --json` |

---

## Binary resolution

The extension resolves the Pablo binary in this order:

1. **Selected executable** (via **Pablo: Select Executable**)
2. **Tools → Options → Pablo → Executable path**
3. **PATH** — only when exactly one `pablo` is found

Workspace `build/pablo.exe` is listed in the executable picker but not auto-selected.

---

## Commands

| Menu | Action |
|------|--------|
| **Tools → Pablo: Check YAML** | `pablo check -f <file>` |
| **Tools → Pablo: Init Config** | `pablo init` |
| **Tools → Pablo: Run Deployment** | Profile + env picker → `pablo run` |
| **Tools → Pablo: Select Executable** | Choose CLI binary |

**CodeLens:** click **Run** on an environment line to skip the picker.

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| No LSP features | Binary must support `pablo lsp` (1.3+); use **Select Executable** |
| Multiple `pablo` on PATH | Use **Select Executable** |
| LSP stopped | Check **View → Output → Pablo Language Server** |

---

## Build from source

See [extensions/vs2026/README.md](../../extensions/vs2026/README.md).

```bat
extensions\vs2026\build-vs2026.bat
```

F5 debug: open `extensions\vs2026\Pablo.sln` in Visual Studio 2026 Insiders.
