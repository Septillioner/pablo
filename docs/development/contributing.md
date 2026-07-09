# Contributing

Summary of the contribution workflow. **Canonical document:** [CONTRIBUTING.md](../../CONTRIBUTING.md)

---

## Ground rules

- Open an issue before non-trivial work to align on direction.
- Keep PRs focused and reviewable.
- Match existing code style and package layout.
- Justify new dependencies in the PR description.
- Check the [roadmap](../roadmap.md) for tracked backlog items.

---

## Setup

**Prerequisites:** Go 1.25.5+, Git. Node.js 20+ for extension work.

```bash
./build.sh              # → build/pablo[.exe]
./build.sh all          # all platforms

cd src
go test ./...           # unit tests
```

Run from source:

```bash
cd src
go run main.go run -e production -p default -f ../pablo.yaml
```

---

## Code conventions

- Concrete struct dependencies; DI wired in `src/main.go`.
- Logging via `pkg/ui`: `+` success, `-` error, `!` warn, `*` info, `>` action.
- Build commands: `sh -c` on Unix, `cmd /C` on Windows.
- Hooks: PowerShell on Windows, `sh` elsewhere.
- Template variables: `{{KEY}}` in config-like file extensions only.
- No emojis in code, comments, or commit messages.

---

## Pull requests

Include:

1. Summary and motivation
2. Manual validation steps (which fixture / environment)
3. User-facing impact (CLI flags, schema changes)
4. Note if touching a known limitation from [capabilities](../reference/capabilities.md)

**Commit style:** short imperative subjects; conventional prefixes encouraged (`feat`, `fix`, `docs`, `chore`).

---

## Documentation changes

When changing behavior, update:

- [Configuration reference](../reference/configuration.md) for schema changes
- [CLI reference](../reference/cli.md) for new flags/commands
- [Capabilities](../reference/capabilities.md) for feature status
- [CHANGELOG.md](../../CHANGELOG.md) for user-facing releases
- [roadmap](../roadmap.md) when completing planned items

---

## Security

Do not report security issues publicly. See [SECURITY.md](../../SECURITY.md).

---

## Full guide

[CONTRIBUTING.md](../../CONTRIBUTING.md)
