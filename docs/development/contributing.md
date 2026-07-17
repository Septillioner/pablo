# Contributing

Summary of the contribution workflow. **Canonical document:** [CONTRIBUTING.md](../../CONTRIBUTING.md)

---

## Ground rules

- Open an issue before non-trivial work to align on direction.
- Keep PRs focused and reviewable.
- Match existing code style and package layout — [Architecture](architecture.md), [PMAP.md](../../PMAP.md).
- Justify new dependencies in the PR description.
- Check the [roadmap](../roadmap.md) for tracked backlog items.

---

## Setup

| Tool | When needed |
|------|-------------|
| Go 1.25.5+ | CLI |
| Git | Always |
| Docker + `ssh-keygen` | E2E |
| Node.js 20+ | VS Code extension |
| Visual Studio + VSSDK | VS2026 extension |

```bash
./scripts/build.sh              # → build/pablo[.exe]
./scripts/build.sh all          # all platforms

./scripts/test.sh unit          # or: cd src && go test ./...
```

Run from source:

```bash
cd src
go run main.go check -f ../pablo.yaml
go run main.go run -f ../pablo.yaml default/production
go run main.go run sequence <name> -f ../pablo.yaml
```

---

## Code conventions

- Concrete struct dependencies; DI wired in `src/main.go`.
- Validation: `pkg/validate` (CLI `check` / `run` + LSP).
- Logging via `pkg/ui`: `+` success, `-` error, `!` warn, `*` info, `>` action.
- Build commands: `sh -c` on Unix, `cmd /C` on Windows.
- Pre/post commands (`deploy.pre_commands` / `deploy.post_commands`): PowerShell on Windows, `sh` elsewhere (`services/hooks`).
- Template variables: `{{KEY}}` in config-like file extensions only.
- Remote paths: `pathutil.JoinRemote`; local: `filepath`.
- No emojis in code, comments, or commit messages.

---

## Extensions

| Area | Path | Docs |
|------|------|------|
| VS Code | `extensions/vscode-pablo/` | [guide](../guides/vscode.md) |
| Visual Studio | `extensions/vs2026/` | [guide](../guides/visual-studio.md) |

Both use `pablo lsp`. Rebuild/verify the CLI when changing LSP or editor integration.

---

## Pull requests

Include:

1. Summary and motivation
2. Manual validation steps (which fixture / environment / E2E scenario)
3. User-facing impact (CLI flags, schema changes)
4. Note if touching a known limitation from [capabilities](../reference/capabilities.md)
5. CHANGELOG update for user-facing changes (`## [Unreleased]`)

**Commit style:** short imperative subjects; conventional prefixes encouraged (`feat`, `fix`, `docs`, `chore`).

---

## Documentation changes

When changing behavior, update:

- [Configuration reference](../reference/configuration.md) for schema changes
- [CLI reference](../reference/cli.md) for new flags/commands
- [Capabilities](../reference/capabilities.md) for feature status
- [Examples](../examples/README.md) when adding or changing a user-facing scenario
- [PMAP.md](../../PMAP.md) for architecture / maintainer notes (do not commit)
- [CHANGELOG.md](../../CHANGELOG.md) for user-facing releases
- [roadmap](../roadmap.md) when completing planned items

---

## Security

Do not report security issues publicly. See [SECURITY.md](../../SECURITY.md).

---

## Full guide

[CONTRIBUTING.md](../../CONTRIBUTING.md) · [Testing](testing.md) · [Architecture](architecture.md)
