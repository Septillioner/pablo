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

1. Open `extensions\vs2026\Pablo.sln` in Visual Studio 2026 Insiders
2. Set **Pablo** as startup project
3. F5 (experimental instance)
4. Tools > Options > Pablo > set executable path or use **Pablo: Select Executable**

## Features

- LSP (`pablo lsp`): diagnostics, completion, hover
- CodeLens **Run** on environment lines
- Profile/environment gutter stripes
- Commands: Check YAML, Init Config, Run Deployment, Select Executable
- Snippets: `pablo-tpl-*`, `pablo-config`
