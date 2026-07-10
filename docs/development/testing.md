# Testing

How Pablo is tested — unit, integration, and manual fixtures.

Detailed catalog: [tests/TEST_SPEC.md](../../tests/TEST_SPEC.md) · Strategy (Turkish): [tests/TEST_PLAN.md](../../tests/TEST_PLAN.md)

---

## Layers

| Layer | Location | Purpose | Requirements |
|-------|----------|---------|--------------|
| **Unit** | `src/**/*_test.go` | Package logic, no network | Go 1.25.5+ |
| **E2E** | `tests/e2e/` | Real SSH + remote Docker in Ubuntu container | Docker, `ssh-keygen` |
| **Manual fixtures** | `tests/agnostic/`, `tests/windows/` | YAML scenarios, hand-run `check` / `run` | Built `pablo` or `go run` |

---

## Commands

### All layers (recommended)

```bash
./test.sh all          # macOS / Linux
./test.ps1 all         # Windows PowerShell
test.bat all           # Windows cmd
```

Modes: `unit`, `integration`, `e2e`, `all` (default).

### Unit tests (run after every change)

```bash
./test.sh unit
# or:
cd src
go test ./...
```

Twelve packages currently have tests: `filter`, `pathutil`, `config`, `template`, `deployer`, `health`, `hooks`, `system`, `ssh`, `pipeline`, `scm`, `docker`.

### E2E integration tests

```bash
./test.sh e2e
# or:
cd tests/e2e
go test -tags=integration -v -timeout 10m ./...
```

Scenarios: SSH static deploy, SSH remote Docker deploy.

### Manual fixture validation

```bash
# Validate only
go run ./src/main.go check -f tests/agnostic/local-deploy/pablo.yaml

# Deploy (writes real files)
cd tests/agnostic/local-deploy
go run ../../../src/main.go run -p local-test -e dev
```

---

## Coverage matrix

| System | Unit | E2E | Manual |
|--------|------|-----|--------|
| Config / inheritance | x | | x |
| Artifact filter | x | | |
| Remote paths | x | | |
| Template `{{VAR}}` | x | | |
| Deployer / strategies | x | | |
| Health check | x | | |
| Hooks | x | | |
| PATH (system adapter) | x | | x |
| SSH adapter | x | x | |
| Pipeline | x | | |
| SCM (git) | x | x | x |
| Docker adapter | x | x | x |
| Full pipeline `Run` | | x | x |

---

## Fixture layout

```
tests/
├── agnostic/          # Cross-platform YAML scenarios
│   ├── local-deploy/
│   ├── multi-profile/
│   ├── separate-apps/
│   └── self-deploy/
├── windows/           # Windows-specific PATH scenarios
└── e2e/               # Docker + SSH integration
```

---

## Rules for contributors

1. New unit tests live beside the package: `foo_test.go` in the same directory.
2. Do not write unit tests that require network — use E2E for remote behavior.
3. Update [TEST_SPEC.md](../../tests/TEST_SPEC.md) when adding tests.
4. Manual validation: note which fixture you used in your PR.

---

## Remaining priorities

See [roadmap — Go unit tests](../roadmap.md#go-unit-testleri):

1. Validation engine
2. Schema metadata
3. LSP adapter
4. Domain types, UI output

---

## Related

- [Contributing](contributing.md)
- [Architecture](architecture.md)
