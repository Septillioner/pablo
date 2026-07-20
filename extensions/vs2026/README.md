# Visual Studio 2026 Pablo Extension

Feature parity with `extensions/vscode-pablo` for Pablo YAML manifests.

## Requirements

- Visual Studio 2026 Insiders (or Visual Studio 2022 17.x)
- **Visual Studio extension development** workload
- Pablo CLI 1.3+ with `pablo lsp`

## Build

```bat
extensions\vs2026\build-vs2026.bat
```

Output: `extensions\vs2026\Pablo\bin\Release\Pablo.VisualStudio.vsix` (or `...\net472\` when built with SDK-style output)

## Debug

1. Open `extensions\vs2026\Pablo.sln` in **Visual Studio 2026 Insiders** (not Cursor)
2. Toolbar configuration: **Debug** (not Release)
3. Set **Pablo** as startup project
4. **F5** — a second VS window opens (`Experimental Instance`). Pablo loads **only in that window**
5. In the Exp window check:
   - **Extensions → Manage Extensions → Installed** → Pablo
   - **Tools → Pablo: Select Executable**
6. Pick `build\pablo.exe` (or any 1.3+ binary), open a `pablo.yaml`
7. Output: **View → Output → Pablo Language Server**

If Tools has no Pablo commands after F5: close both VS windows, rebuild Debug once, F5 again.

### Install into main VS (no F5)

```bat
extensions\vs2026\build-vs2026.bat
```

Double-click `Pablo\bin\Release\net472\Pablo.VisualStudio.vsix`, restart VS.

## Features

- LSP (`pablo lsp`): diagnostics, completion, hover
- CodeLens **Run** on environment lines (via LSP + `PabloMiddleLayer`)
- Profile/environment gutter stripes
- Commands: Check YAML, Init Config, Run Deployment, Select Executable
- On activation: one-shot CLI update check (`pablo update check --json`) with Yes/No update prompt
- **Pablo Run Deployment** tool window — profile + environment combos, **Run Deployment**, **Refresh**; opened from **Tools → Pablo: Run Deployment**
- **Pablo** toolbar (**View → Toolbars → Pablo**) — **Manifest** / **Profile** / **Environment** combos + **Run**; discovers manifests in the solution
- Snippets: `pablo-tpl-*`, `pablo-config`
