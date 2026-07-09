# API Reference

Pablo does not expose an HTTP REST API. The programmatic surfaces are the CLI `inspect --json` output and the embedded Language Server Protocol (LSP) server.

---

## `inspect --json`

```bash
pablo inspect -f pablo.yaml --json
```

Outputs JSON to stdout with no CLI header. Used by scripts and the VS Code extension as a fallback when LSP is unavailable.

**Response shape:**

```json
{
  "name": "my-app",
  "version": "1.3.0",
  "profiles": [
    {
      "name": "default",
      "type": "static",
      "environments": ["production", "staging"]
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Manifest `name` |
| `version` | string | Manifest `version` |
| `profiles` | array | List of profiles |
| `profiles[].name` | string | Profile key |
| `profiles[].type` | string | Deployment type |
| `profiles[].environments` | string[] | Environment keys under this profile |

---

## Language Server (`pablo lsp`)

Transport: stdio. Started by the VS Code extension or manually for debugging.

### Standard LSP capabilities

| Capability | Description |
|------------|-------------|
| `textDocument/publishDiagnostics` | Live YAML validation (same rules as `pablo check`) |
| `textDocument/completion` | Field and enum suggestions from `pkg/schema` |
| `textDocument/hover` | Field documentation |
| `textDocument/codeLens` | **Run profile/env** links on environment lines |

Document sync: full document sync for `pablo.yaml` and `pablo*.yml` files.

### Custom request: `pablo/listProfiles`

Returns the same structure as `inspect --json` for a manifest.

**Params:**

```json
{
  "uri": "file:///path/to/pablo.yaml"
}
```

If the document is open in the editor, the server reads the in-memory buffer; otherwise it reads from disk.

**Implementation:** `src/internal/lsp/profiles.go` · shared logic in `pkg/inspect`.

---

## VS Code extension commands

| Command | CLI equivalent |
|---------|----------------|
| Pablo: Check YAML | `pablo check -f <file>` |
| Pablo: Init Config | `pablo init` |
| Pablo: Run Deployment | `pablo run -f <file> -p <profile> -e <env>` |
| Pablo: Select Executable | *(no CLI equivalent)* |

CodeLens **Run profile/env** invokes `pablo run` with the manifest path, profile, and environment from the clicked line.

See [VS Code guide](../guides/vscode.md) for setup and troubleshooting.
