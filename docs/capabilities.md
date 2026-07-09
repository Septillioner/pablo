# Pablo Capabilities

Overview of supported deployment types, strategies, pipeline behavior, and current limitations.

---

## Deployment Types

| Type | Description | Local | Remote SSH | Status |
|------|-------------|-------|------------|--------|
| `static` | Frontend / SPA — build, filter artifacts, deploy files | Yes | Yes | Working |
| `binary` | Compiled executables — build, deploy, PATH register | Yes | Yes | Working |
| `docker` | Docker Compose — git clone/pull, compose up | Yes | Yes | Working |
| `git-sync` | Interpreted languages — git pull, post commands | Yes | Yes | Working |

---

## Deploy Strategies

| Strategy | Description | Status |
|----------|-------------|--------|
| `overwrite` | Copy files over existing (default) | Working |
| `backup` | Rename existing dir with timestamp, then deploy | Working |
| `recreate` | Delete target dir, create fresh, deploy | Working |
| `blue-green` | Zero-downtime swap | Not implemented |

---

## Pipeline Phases

1. Pre-deploy hooks (`hooks.pre`)
2. Build (`build.command`)
3. Pre-deployment commands (`deploy.pre_commands`)
4. Deployment (local copy or SSH tar stream)
5. Post-deployment commands (`deploy.post_commands`)
6. Post-deploy hooks (`hooks.post`)
7. Health check (`pipeline.health_check`)

---

## What Works

- Full local deploy pipeline for `static` and `binary` types.
- Remote SSH deploy with tar-streaming and SCP fallback.
- Glob-based artifact filtering (include/exclude patterns).
- Template variable substitution (`{{VAR}}` in config files).
- Config inheritance — profile settings cascade into environments.
- Automatic PATH registration (Windows, macOS, Linux user scope).
- Backup and recreate strategies with protected path detection.
- `docker` type with local and remote (SSH) Docker Compose orchestration.
- `git-sync` with local and remote (SSH) git clone/pull.
- Environment variable injection via `.env` file generation.
- LSP-powered VS Code extension with completion, hover, and YAML validation.
- Go unit tests for twelve packages: `filter`, `pathutil`, `config`, `template`, `deployer`, `health`, `hooks`, `system`, `ssh`, `pipeline`, `scm`, `docker` (`cd src && go test ./...`).
- Test documentation: [tests/TEST_PLAN.md](../tests/TEST_PLAN.md) and [tests/TEST_SPEC.md](../tests/TEST_SPEC.md).
- Docker-based E2E tests for remote SSH static and docker deploy (`cd tests/e2e && go test -tags=integration ./...`).

---

## Known Limitations

- **Partial unit test coverage** — twelve packages have `*_test.go`; catalog in [tests/TEST_SPEC.md](../tests/TEST_SPEC.md). Optional SSH command mocks and `domain`/`ui` remain in [goals.md](goals.md).
- `blue-green` **strategy** — declared but not implemented (returns error).
- **SSH host key verification** — currently disabled (`InsecureIgnoreHostKey`); see [SECURITY.md](../SECURITY.md).
- **Schema validation coverage** — core rules in `pkg/validate`; advanced cross-field rules still expanding (see [goals.md](goals.md)).
- `builder.Service` — exists as a standalone service but is currently unused; builds run inline.
- **Snippet versions** — hardcoded; not synced with the `VERSION` file.
