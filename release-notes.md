## Pablo v2.2.2

LSP and validation fixes for tab-indented YAML editing.

### Fixed

- **LSP / validation — tab indentation** — Schema path resolution now treats leading tabs as indent (previously only spaces), so completion and hover work while editing tab-indented YAML. `pablo lsp` also triggers completion on Tab. Validation reports a clear error when indentation uses tabs instead of only the generic YAML parse failure.

### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
