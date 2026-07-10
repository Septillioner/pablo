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
| Syntax highlighting | YAML editor + `pablo` / inbox YAML content types |
| Schema validation (squiggles) | `pablo lsp` → `pkg/validate` |
| Autocomplete | `pablo lsp` → `pkg/schema` |
| Hover docs | `pablo lsp` |
| Snippet templates (`pablo-tpl-*`) | Extension snippets |
| Check YAML / Init / Select Executable | Extension → CLI |
| **Run Deployment** tool window | Dockable panel: manifest path, profile/env combos, Run + Refresh |
| **Pablo** toolbar | **View → Toolbars → Pablo** — Manifest / Profile / Environment + Run |
| CodeLens **Run profile/env** | LSP + extension → `pablo run` |
| Profile/env gutter stripes | Extension adornments |
| Profile/env picker | `pablo/listProfiles` or `inspect --json` |

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
| **Tools → Pablo: Run Deployment** | Opens **Pablo Run Deployment** tool window (profile + env combos → `pablo run`) |
| **Pablo** toolbar (**View → Toolbars → Pablo**) | **Manifest** / **Profile** / **Environment** combos + **Run** — discovers `pablo*.yaml` in the solution and runs `pablo run -f … profile/env` |
| **Tools → Pablo: Select Executable** | Choose CLI binary |

**CodeLens:** requires a working LSP connection and **Select Executable** (Pablo 1.3+). Click **Run** on an environment line to deploy without the toolbar or tool window.

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| No LSP features | Binary must support `pablo lsp` (1.3+); use **Select Executable**; check **View → Output → Pablo Language Server** for `Starting Pablo LSP` |
| No CodeLens | CodeLens comes from `pablo lsp`; fix LSP first (executable + YAML file open) |
| Run / tool window cannot find manifest | Keep a `pablo*.yaml` open, or pick one from the **Pablo** toolbar Manifest combo |
| Tool window shows inspect / executable errors | Configure CLI via **Select Executable**, then **Refresh** |
| Dark theme: unreadable combos | Update to a build that uses VS `ThemedDialog*` / `EnvironmentColors` styles |
| Multiple `pablo` on PATH | Use **Select Executable** |
| LSP stopped | Check **View → Output → Pablo Language Server** |

---

## Build from source

See [extensions/vs2026/README.md](../../extensions/vs2026/README.md).

```bat
extensions\vs2026\build-vs2026.bat
```

F5 debug: open `extensions\vs2026\Pablo.sln` in Visual Studio 2026 Insiders.
