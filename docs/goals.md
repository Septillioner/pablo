# Pablo Roadmap

Public view of what Pablo **ships today** and what is **planned next**. For current capabilities see [capabilities.md](capabilities.md). To contribute, start with [CONTRIBUTING.md](../CONTRIBUTING.md).

**How to maintain:** Mark completed items `[x]` in this file. User-facing changes belong in [CHANGELOG.md](../CHANGELOG.md). Test coverage updates go to [tests/TEST_SPEC.md](../tests/TEST_SPEC.md).

**Related docs:** [README.md](../README.md) · [cli.md](cli.md) · [capabilities.md](capabilities.md) · [schema/schema.md](schema/schema.md) · [SECURITY.md](../SECURITY.md) · [CONTRIBUTING.md](../CONTRIBUTING.md)

---

## At a glance

| Status | Count (approx.) |
|--------|-----------------|
| Shipped | 28 |
| Planned | 14 |

---

## Feature tree

```
Pablo
├── Manifest & CLI
│   ├── [x] Multi-profile pablo.yaml manifests
│   ├── [x] Profile-to-environment inheritance
│   ├── [x] check, run, init, uninstall, version
│   ├── [x] lsp (language server, stdio)
│   └── [ ] Streamlined build pipeline (unused builder cleanup)
├── Editor experience
│   ├── [x] VS Code extension (syntax, snippets, commands)
│   ├── [x] Language server via pablo lsp
│   ├── [x] Completion, hover, live diagnostics
│   └── [ ] Snippet versions synced with releases
├── Validation
│   ├── [x] Schema checks in CLI (path:line:col errors)
│   ├── [x] Same rules in the editor (pablo lsp)
│   ├── [x] Run blocked when manifest is invalid
│   └── [ ] Richer validation (cross-field rules, unknown keys)
├── Deployment
│   ├── [x] Types: static, binary, docker, git-sync
│   ├── [x] Strategies: overwrite, backup, recreate
│   ├── [x] Local deploy pipeline
│   ├── [x] Remote SSH deploy (tar streaming)
│   ├── [x] Remote Docker Compose over SSH
│   ├── [ ] Service management (systemd / PM2)
│   └── [ ] Blue-green deployments
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
│   ├── [x] Public docs (cli, capabilities, schema)
│   ├── [x] Test strategy and catalog
│   └── [ ] Docs kept in sync with each release
└── Security
    └── [ ] SSH host key verification
```

---

## Shipped

### Manifest & CLI

- [x] Multi-profile `pablo.yaml` with environment-specific deploy settings
- [x] `pablo check` — validate manifests before deploy
- [x] `pablo run` — full deployment pipeline
- [x] `pablo init`, `uninstall`, `version`
- [x] `pablo lsp` — embedded language server (single binary)

### Editor experience

- [x] VS Code extension with Pablo language ID and snippets
- [x] Real-time diagnostics, completion, and hover via `pablo lsp`
- [x] Editor commands: Check YAML, Init Config, Run Deployment
- [x] Configurable CLI path (`pablo.path`) or PATH lookup

### Validation

- [x] Shared schema validation for CLI and editor
- [x] Line and column error locations in `pablo check`
- [x] Deploy blocked when validation fails

### Deployment

- [x] Deployment types: `static`, `binary`, `docker`, `git-sync`
- [x] Strategies: `overwrite`, `backup`, `recreate`
- [x] Artifact filtering (include / exclude globs)
- [x] Template variable substitution (`{{VAR}}`)
- [x] Hooks, health checks, pre/post deploy commands
- [x] Local and remote SSH deploy
- [x] Remote Docker Compose orchestration

### Platform integration

- [x] Automatic PATH registration for binary deployments
- [x] Linux system-scope PATH via `/etc/profile.d/`
- [x] Windows PATH removal on `pablo uninstall`
- [x] Protected system directory detection

### Quality & testing

- [x] Unit tests across twelve core packages (`cd src && go test ./...`)
- [x] Integration tests: Ubuntu SSH target in Docker ([tests/e2e](../tests/e2e/))
  - SSH static deploy scenario
  - SSH remote docker deploy scenario
  - Run: `cd tests/e2e && go test -tags=integration -v -timeout 10m ./...`

### Documentation

- [x] [cli.md](cli.md), [capabilities.md](capabilities.md), [schema/schema.md](schema/schema.md)
- [x] [tests/TEST_PLAN.md](../tests/TEST_PLAN.md) and [tests/TEST_SPEC.md](../tests/TEST_SPEC.md)

---

## Planned

### Deployment (P1 — missing features)

- [ ] **Service management** — `deploy.service` for systemd and PM2 (schema exists; runtime not implemented)
- [ ] **Blue-green strategy** — declared in schema; returns error at runtime

### Validation (P2 — quality)

- [ ] **Expanded schema rules** — cross-field checks, unknown YAML keys, richer error positions
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
- [ ] **Documentation sync** — keep [cli.md](cli.md) and [capabilities.md](capabilities.md) current

### Security (P0)

- [ ] **SSH host key verification** — replace `InsecureIgnoreHostKey`; see [SECURITY.md](../SECURITY.md)

### Automation (P4)

- [ ] **GitHub Actions** — unit tests, optional integration job, release workflow

---

## Priority guide

| Priority | Focus | Examples |
|----------|-------|----------|
| P0 | Security | SSH host key pinning |
| P1 | Missing features | systemd/PM2, blue-green |
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
