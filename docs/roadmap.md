# Pablo Roadmap

Public view of what Pablo **ships today** and what is **planned next**. For current capabilities see [reference/capabilities.md](reference/capabilities.md). To contribute, start with [CONTRIBUTING.md](../CONTRIBUTING.md).

**How to maintain:** Mark completed items `[x]` in this file. User-facing changes belong in [CHANGELOG.md](../CHANGELOG.md). Test coverage updates go to [tests/TEST_SPEC.md](../tests/TEST_SPEC.md).

**Related docs:** [README.md](../README.md) · [docs/](README.md) · [SECURITY.md](../SECURITY.md) · [CONTRIBUTING.md](../CONTRIBUTING.md)

---

## At a glance

| Status | Count (approx.) |
|--------|-----------------|
| Shipped | 31 |
| Planned | 12 |

---

## Feature tree

```
Pablo
├── Manifest & CLI
│   ├── [x] Multi-profile pablo.yaml manifests
│   ├── [x] Profile-to-environment inheritance (variables, env_file, build)
│   ├── [x] Schema v2 — profiles-only, deploy.source, deploy.transfer
│   ├── [x] check, run, init (--template), uninstall, version, inspect, update
│   ├── [x] Manifest sequences (`pablo run sequence <name>`)
│   ├── [x] lsp (language server, stdio)
│   └── [ ] Streamlined build pipeline (unused builder cleanup)
├── Editor experience
│   ├── [x] VS Code extension (syntax, snippets, commands)
│   ├── [x] Visual Studio extension (LSP, tool window, toolbar)
│   ├── [x] Language server via pablo lsp
│   ├── [x] Completion, hover, live diagnostics
│   ├── [x] inspect protocol (CLI + LSP listProfiles)
│   ├── [x] Run command profile/environment picker
│   ├── [x] CodeLens Run on environment lines
│   └── [ ] Snippet versions synced with releases
├── Validation
│   ├── [x] Schema checks in CLI (path:line:col errors)
│   ├── [x] Same rules in the editor (pablo lsp)
│   ├── [x] Run blocked when manifest is invalid
│   ├── [x] Unknown YAML keys rejected
│   └── [ ] Richer validation (cross-field rules, richer positions)
├── Deployment
│   ├── [x] Types: static, binary, docker, git-sync
│   ├── [x] Strategies: overwrite, backup, recreate, rename-replace
│   ├── [x] Local deploy pipeline
│   ├── [x] Remote SSH deploy (tar streaming)
│   └── [x] Remote Docker Compose over SSH
├── Platform integration
│   ├── [x] PATH registration (Windows, macOS, Linux user)
│   ├── [x] Linux system-scope PATH
│   ├── [x] Windows PATH cleanup on uninstall
│   └── [x] Safe remote paths (Windows host to Linux target)
├── Quality & testing
│   ├── [x] Unit tests for core packages
│   ├── [x] Docker + SSH integration tests
│   ├── [ ] CI on every pull request
│   └── [ ] Broader automated test coverage
├── Documentation
│   ├── [x] Public docs (getting-started, guides, reference, examples)
│   ├── [x] Schema v2 docs rewrite + scenario catalog
│   ├── [x] Test strategy and catalog
│   └── [ ] Docs kept in sync with each release
└── Security
    └── [x] SSH host key verification
```

---

## Shipped

### Manifest & CLI

- [x] Multi-profile `pablo.yaml` with environment-specific deploy settings
- [x] Schema v2 — three nouns (Profile, Environment, Deploy); `deploy.source`, `deploy.transfer`, named credentials
- [x] `pablo check` — validate manifests before deploy
- [x] `pablo run` — full deployment pipeline
- [x] `pablo run sequence <name>` — run named manifest sequences in list order
- [x] `pablo init` (optional `--template` / `-t` wizard), `uninstall`, `version`, `inspect`, `update`
- [x] `pablo lsp` — embedded language server (single binary)

### Editor experience

- [x] VS Code extension with Pablo language ID and snippets
- [x] Visual Studio extension — LSP, **Run Deployment** tool window, **Pablo** toolbar (Manifest / Profile / Environment + Run)
- [x] Real-time diagnostics, completion, and hover via `pablo lsp`
- [x] Editor commands: Check YAML, Init Config, Run Deployment
- [x] Run Deployment picks profile and environment via `pablo inspect` / `pablo/listProfiles`
- [x] CodeLens **Run profile/env** on each environment line in `pablo.yaml`
- [x] Configurable CLI path (`pablo.path`) or PATH lookup

### Validation

- [x] Shared schema validation for CLI and editor
- [x] Line and column error locations in `pablo check`
- [x] Deploy blocked when validation fails

### Deployment

- [x] Deployment types: `static`, `binary`, `docker`, `git-sync`
- [x] Strategies: `overwrite`, `backup`, `recreate`, `rename-replace`
- [x] Artifact filtering (include / exclude globs)
- [x] Template variable substitution (`{{VAR}}`)
- [x] Pre/post deploy commands (`deploy.pre_commands`, `deploy.post_commands`)
- [x] Local and remote SSH deploy
- [x] Remote Docker Compose orchestration
- [x] SSH host key verification (`known_hosts`, optional TOFU / opt-out)

### Platform integration

- [x] Automatic PATH registration for binary deployments
- [x] Linux system-scope PATH via `/etc/profile.d/`
- [x] Windows PATH removal on `pablo uninstall`
- [x] Protected system directory detection

### Quality & testing

- [x] Unit tests across twelve core packages (`cd src && go test ./...`)
- [x] Integration tests: Ubuntu SSH target in Docker ([tests/e2e](../tests/e2e/))
  - Real-world stories: static site, hotfix (`rename-replace`), Go binary, Compose API, PHP git-sync
  - Strategies: overwrite, backup, recreate, rename-replace; plus sequence, legacy transfer, checksum
  - Run: `cd tests/e2e && go test -tags=integration -v -timeout 15m ./...`

### Documentation

- [x] [docs/](README.md) — getting-started, guides, reference, development (schema v2 rewrite)
- [x] [Examples](examples/README.md) — sixteen scenarios aligned with test fixtures
- [x] [tests/TEST_PLAN.md](../tests/TEST_PLAN.md) and [tests/TEST_SPEC.md](../tests/TEST_SPEC.md)

---

## Planned

### Validation (P2 — quality)

- [x] **Unknown YAML keys** — rejected by shared validate (CLI + LSP)
- [ ] **Expanded schema rules** — more cross-field checks, richer error positions
- [ ] **Schema automation (phase 2)** — generate editor metadata and rules from a single source

### Quality & testing (P2)

<a id="go-unit-testleri"></a>

- [ ] **Broader unit test coverage** — validation, schema, language server adapter, remaining packages
- [ ] **CI pipeline**
  - [ ] `go test ./...` on every pull request
  - [ ] Integration tests (`-tags=integration`) in a separate job or scheduled run
  - [ ] Manual fixture spot-checks documented in CI

#### Remaining test priorities

| Order | Area | Why |
|-------|------|-----|
| 1 | Validation engine | Shared by CLI and editor; high regression risk |
| 2 | Schema metadata | Powers completion and hover |
| 3 | Language server adapter | Protocol mapping |
| 4 | Domain types, UI output | Lower risk; mostly declarative |

Catalog: [tests/TEST_SPEC.md](../tests/TEST_SPEC.md)

### Platform & editor (P3)

- [ ] **Snippet version sync** — align VS Code snippets with release version
- [ ] **Build service cleanup** — integrate or remove unused builder abstraction
- [ ] **Documentation sync** — keep reference docs current with each release

### Security (P0)

- [x] **SSH host key verification** — `known_hosts` by default; `remote.host_key_verification` / `remote.trust_on_first_use`; see [SECURITY.md](../SECURITY.md)

### Automation (P4)

- [ ] **GitHub Actions** — unit tests, optional integration job, release workflow

---

## Priority guide

| Priority | Focus | Examples |
|----------|-------|----------|
| P0 | Security | SSH host key pinning |
| P1 | Missing features | — |
| P2 | Quality | Validation depth, tests, CI |
| P3 | Platform & DX | Snippets, docs sync |
| P4 | Automation | GitHub Actions, release pipeline |
| P5 | Future | Schema reflection, unknown-key diagnostics |

---

## How to use this file

1. Before starting work, find the matching branch in the feature tree or **Planned** section.
2. Open a PR that marks the relevant checkbox `[x]` when done.
3. Add user-facing notes to [CHANGELOG.md](../CHANGELOG.md).
4. Update [tests/TEST_SPEC.md](../tests/TEST_SPEC.md) when test goals are completed ([TEST_PLAN.md](../tests/TEST_PLAN.md) workflow).
