# Pablo for Visual Studio Code

Pablo is a powerful deployment helper for managing multi-environment builds and deployments. This extension provides first-class support for `pablo.yaml` configuration files.

## Features

- **LSP Powered Intelligence**: Real-time validation, diagnostics, and autocompletion powered by a high-performance Go-based Language Server.
- **Smart Autocompletion**: Context-aware suggestions for all Pablo configuration fields and enum values.
- **Hover Documentation**: Quick access to documentation and field descriptions directly within the editor.
- **Scaffold Templates**: Ready-to-use templates for common scenarios:
    - `pablo-tpl-static`: Static Website
    - `pablo-tpl-node-pm2`: Node.js with PM2
    - `pablo-tpl-go-systemd`: Go with Systemd
    - `pablo-tpl-docker`: Docker Compose
- **Custom File Icon**: Beautiful logo integration for your `pablo.yaml` files.
- **CLI Integration**: Run Pablo commands (`check`, `init`, `run`) directly from the command palette.

## Usage

1. Open any `pablo.yaml` or `pablo.yml` file.
2. Start typing `pablo-` to see available templates.
3. Use `Ctrl+Space` for smart completions.
4. Hover over any key to see its description.

## Requirements

- This extension bundles pre-compiled Go binaries for the LSP server. No additional dependencies are required for basic editor features.
- To run deployments, the `pablo` CLI must be installed on your system.

## Release Notes

### 0.0.1
- Initial release with full LSP support.
- Cross-platform support for Windows, macOS (Intel/Silicon), and Linux.
- Comprehensive set of snippets and templates.

---
**Enjoy efficient deployments with Pablo!**
