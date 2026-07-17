# Pablo Capabilities

What Pablo supports today: deployment types, strategies, pipeline phases, and known limits.

See also: [Configuration](configuration.md) · [Examples](../examples/README.md) · [Roadmap](../roadmap.md)

---

## Deployment types

| Type | Description | Local | Remote SSH | Status |
|------|-------------|-------|------------|--------|
| `static` | Filter and deploy files; `build` optional (omit to copy as-is) | Yes | Yes | Working |
| `binary` | Compiled executables — build, deploy, PATH register | Yes | Yes | Working |
| `docker` | Docker Compose — stop if running (default), git clone/pull, compose up | Yes | Yes | Working |
| `git-sync` | Interpreted languages — git pull, post commands | Yes | Yes | Working |

---

## Deploy strategies

| Strategy | Description | Status |
|----------|-------------|--------|
| `overwrite` | Copy files over existing (default) | Working |
| `backup` | Rename existing dir with timestamp, then deploy | Working |
| `recreate` | Delete target dir, create fresh, deploy | Working |
| `rename-replace` | Rename existing artifact files, replace, cleanup on success | Working |

---

## Pipeline phases

Single-target `pablo run` (one profile + environment):

1. Load and validate manifest
2. Build (`build.command`) — skipped when `build` is omitted or empty
3. Pre-deployment commands (`deploy.pre_commands`)
4. Deployment (local copy or SSH tar stream)
5. Post-deployment commands (`deploy.post_commands`)
6. PATH registration (`register_path`, binary type)

### Sequences

Root-level `sequences` name ordered lists of `profile/env` targets. `pablo run sequence <name>` runs each step with the full pipeline above, in list order, and stops on the first failure. See [Sequences](../guides/sequences.md) · [Configuration — Sequences](configuration.md#sequences).

---

## What works

- Full local deploy pipeline for `static` and `binary` (`static` works without `build` — copy/filter only).
- Remote SSH deploy with tar-streaming and SCP fallback (`deploy.transfer: legacy`); optional `deploy.verify_checksum` for post-transfer SHA-256 checks.
- Gitignore-style glob artifact filtering (`*.ext` at any depth, `/*.ext` root-only, `**` globstar).
- Template variable substitution (`{{VAR}}` in config files).
- Config inheritance — profile `variables`, `env_file`, and `build` cascade into environments.
- Named `sequences` — ordered multi-target runs via `pablo run sequence <name>` (stops on first failure).
- Automatic PATH registration (Windows, macOS, Linux user and system scope).
- Backup, recreate, and rename-replace strategies with protected path detection (backup/recreate only).
- `docker` type with local and remote (SSH) Docker Compose orchestration.
- `git-sync` with local and remote (SSH) git clone/pull.
- Environment variable injection via env file generation.
- LSP-powered VS Code and Visual Studio extensions with completion, hover, and YAML validation.
- Go unit tests for core packages (`cd src && go test ./...`).
- Docker-based E2E tests for remote SSH static, rename-replace, and docker deploy (`cd tests/e2e && go test -tags=integration ./...`).
- Windows fixtures for rename-replace and NSSM service install via `post_commands`.

---

## Known limitations

- Partial unit test coverage — catalog in [tests/TEST_SPEC.md](../../tests/TEST_SPEC.md).
- SSH host key verification is enabled by default via `known_hosts`; opt out with `remote.host_key_verification: off`; optional `remote.trust_on_first_use` — see [SECURITY.md](../../SECURITY.md).
- Schema validation coverage — core rules in `pkg/validate`; advanced cross-field rules still expanding (see [roadmap](../roadmap.md)).
- `builder.Service` exists as a standalone service but is currently unused; builds run inline.
- Snippet versions are hardcoded in the VS Code extension and not synced with the `VERSION` file.
