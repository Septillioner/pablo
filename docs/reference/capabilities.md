# Pablo Capabilities

Overview of supported deployment types, strategies, pipeline behavior, and current limitations.

See also: [Configuration](configuration.md) · [Roadmap](../roadmap.md)

---

## Deployment Types

| Type | Description | Local | Remote SSH | Status |
|------|-------------|-------|------------|--------|
| `static` | Filter and deploy files; `build` optional (omit to copy as-is) | Yes | Yes | Working |
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
| `rename-replace` | Rename existing artifact files, replace, cleanup on success | Working |
| `blue-green` | Zero-downtime swap | Not implemented |

---

## Pipeline Phases

1. Load and validate manifest
2. Pre-deploy hooks (`hooks.pre`)
3. Build (`build.command`) — skipped when `build` is omitted or empty
4. Pre-deployment commands (`deploy.pre_commands`)
5. Deployment (local copy or SSH tar stream)
6. Post-deployment commands (`deploy.post_commands`)
7. PATH registration (`register_path`, binary type)
8. Post-deploy hooks (`hooks.post`)
9. Health check (`pipeline.health_check`)
10. `on_success` / `on_failure` hooks

---

## Schema vs Runtime

Some fields are validated in the manifest but not yet executed at runtime:

| Field | Schema | Runtime |
|-------|--------|---------|
| `deploy.strategy: blue-green` | Allowed | Returns error |
| `deploy.service` (systemd / PM2) | Allowed | Not implemented — use `post_commands` |

---

## What Works

- Full local deploy pipeline for `static` and `binary` types (`static` works without `build` — copy/filter only).
- Remote SSH deploy with tar-streaming and SCP fallback (`deploy.remote: legacy`).
- Gitignore-style glob artifact filtering (`*.ext` at any depth, `/*.ext` root-only, `**` globstar).
- Template variable substitution (`{{VAR}}` in config files).
- Config inheritance — profile settings cascade into environments.
- Automatic PATH registration (Windows, macOS, Linux user and system scope).
- Backup, recreate, and rename-replace strategies with protected path detection (backup/recreate only).
- `docker` type with local and remote (SSH) Docker Compose orchestration.
- `git-sync` with local and remote (SSH) git clone/pull.
- Environment variable injection via `.env` file generation.
- LSP-powered VS Code extension with completion, hover, and YAML validation.
- Go unit tests for twelve packages: `filter`, `pathutil`, `config`, `template`, `deployer`, `health`, `hooks`, `system`, `ssh`, `pipeline`, `scm`, `docker` (`cd src && go test ./...`).
- Docker-based E2E tests for remote SSH static and docker deploy (`cd tests/e2e && go test -tags=integration ./...`).

---

## Known Limitations

- **Partial unit test coverage** — catalog in [tests/TEST_SPEC.md](../../tests/TEST_SPEC.md).
- `blue-green` **strategy** — declared but not implemented (returns error).
- **`deploy.service`** — schema exists; systemd/PM2 restart not implemented at runtime.
- **SSH host key verification** — currently disabled (`InsecureIgnoreHostKey`); see [SECURITY.md](../../SECURITY.md).
- **Schema validation coverage** — core rules in `pkg/validate`; advanced cross-field rules still expanding (see [roadmap](../roadmap.md)).
- `builder.Service` — exists as a standalone service but is currently unused; builds run inline.
- **Snippet versions** — hardcoded in the VS Code extension; not synced with the `VERSION` file.
