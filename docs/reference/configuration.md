# Configuration Reference

Complete field reference for Pablo’s deployment manifest (`pablo.yaml`). Progressive copy-paste scenarios live in [Examples](../examples/README.md). Capabilities and limits: [Capabilities](capabilities.md). Credentials guide: [Credentials](../guides/credentials.md).

---

## Three nouns

| Noun | YAML location | Meaning |
|------|---------------|---------|
| **Profile** | `profiles.<name>` | **What** you deploy — type, build, git, shared variables |
| **Environment** | `profiles.<name>.environments.<name>` | **Where** it runs — local or remote (`remote`), deploy settings |
| **Deploy** | `environments.<name>.deploy` | **How** artifacts land — source, strategy, commands, docker |

Everything deployable lives under `profiles`. Root-level fields are project metadata (`name`, `version`), shared `credentials`, and optional `sequences`.

---

## Core rules

1. **Profiles only** — All deploy config is under `profiles`.
2. **Remote means SSH** — If `remote` is present, Pablo deploys over SSH. If absent, deploy is local. Fields: `host`, `credential`, `host_key_verification`, `trust_on_first_use`.
3. **Artifacts via `deploy.source`** — Static and binary types use `deploy.source` (`dir`, `include`, `exclude`) on each environment.
4. **Shell around deploy** — Use `deploy.pre_commands` and `deploy.post_commands`.
5. **Named credentials** — Define credentials at the root; reference by name (`remote.credential`, `git.credential`).

---

## Inheritance

Only these profile fields cascade into each environment:

| Profile field | Inherited as |
|---------------|--------------|
| `variables` | Merged into environment `variables` (env wins on conflict) |
| `env_file` | Default env file name when environment omits `env_file` |
| `build` | Copied when env has no `build`; partial merge when env `build` exists |

`deploy.source`, `deploy.target_path`, and `remote` are always set on the environment — they do not inherit.

---

## Type matrix

| Type | Required | Forbidden on environment |
|------|----------|--------------------------|
| `static` | `deploy.source`, `deploy.target_path` | `git`, `deploy.docker` |
| `binary` | `build.command` (profile or env), `deploy.source`, `deploy.target_path` | `git`, `deploy.docker` |
| `docker` | `git.repo`, `deploy.target_path`, `deploy.docker.compose_file` | `deploy.source`, `register_path` |
| `git-sync` | `git.repo`, `deploy.target_path` | `deploy.source`, `deploy.docker`, `register_path` |

`build` is optional for `static` (omit to copy files as-is). `build.command` is required for `binary`.

---

## Root fields

| Field | Type | Description |
|---|---|---|
| `name` | String | Project name |
| `version` | String | Project version |
| `credentials` | Map<String, [Credential](#credential)> | Named reusable credentials (optional) |
| `sequences` | Map<String, String[]> | Named ordered deployment sequences (optional); see [Sequences](#sequences) |
| `profiles` | Map<String, [Profile](#profile)> | **Required.** Application profiles |

---

## Sequences

Named lists of `profile/environment` targets. List order is execution order — Pablo runs each step sequentially and stops on the first failure.

| Field | Type | Description |
|---|---|---|
| `<name>` | String[] | Ordered steps; each item is `profile/env` (e.g. `api/staging`) |

```yaml
sequences:
  release:
    - api/staging
    - api/production
```

```bash
pablo run sequence release
```

Cannot combine `pablo run sequence` with `-p` / `-e`. Global flags (`-f`, `--force`, `--verbose`) apply to every step.

Guide: [Sequences](../guides/sequences.md) · [Examples #11](../examples/README.md#11-sequences).

---

## Credential

| Field | Type | Description |
|---|---|---|
| `type` | String | **Required.** `ssh`, `token`, or `basic` |
| `username` | String | Username (`ssh`, `basic`) |
| `password` | String | Password (`basic`, or `ssh` password auth) |
| `key` | String | SSH private key path (`ssh`) |
| `passphrase` | String | SSH key passphrase (optional) |
| `value` | String | Token value (`token`) |

```yaml
credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519
  github:
    type: token
    value: ghp_xxxxx
  registry:
    type: basic
    username: dockeruser
    password: "${REGISTRY_PASSWORD}"
```

---

## Profile

| Field | Type | Description |
|---|---|---|
| `type` | String | **Required.** `static`, `binary`, `docker`, or `git-sync` |
| `variables` | Map<String, String> | Variables inherited by all environments |
| `env_file` | String | Env file name inherited by all environments |
| `build` | [Build](#build) | Build config (required for `binary`; optional for `static`) |
| `git` | [Git](#git) | Git repo config (`docker`, `git-sync` only) |
| `environments` | Map<String, [Environment](#environment)> | **Required.** Deployment targets |

---

## Build

At profile level (inherited) or overridden per environment.

| Field | Type | Description |
|---|---|---|
| `command` | String | **Required when `build` is set.** Shell command (e.g. `npm run build`, `go build -o app .`) |
| `path` | String | Working directory for the build command (relative to manifest) |
| `variables` | Map<String, String> | Environment variables for the build process only |
| `env_file` | String | Write variables to this file before building |

```yaml
build:
  command: npm run build
  path: ./frontend
  variables:
    NODE_ENV: production
```

---

## Git

For `docker` and `git-sync` profiles.

| Field | Type | Description |
|---|---|---|
| `repo` | String | **Required.** Git repository URL |
| `branch` | String | Branch name (default: `main`) |
| `credential` | String | Credential name for private HTTPS repos |

```yaml
git:
  repo: https://github.com/user/project.git
  branch: main
  credential: github
```

---

## Environment

| Field | Type | Description |
|---|---|---|
| `deploy` | [Deploy](#deploy) | **Required.** Deployment settings |
| `remote` | [Remote](#remote) | SSH connection — present means remote deploy |
| `build` | [Build](#build) | Override profile-level build |
| `variables` | Map<String, String> | Runtime variables (merged with profile) |
| `env_file` | String | Env file written into the deploy target |
| `register_path` | [RegisterPath](#registerpath) | PATH registration (`binary` only) |

---

## Remote

When present, deployment targets the remote host via SSH. Omit for local deploy.

| Field | Type | Description |
|---|---|---|
| `host` | String | **Required.** Remote host (`host` or `host:port`; port defaults to 22) |
| `credential` | String | **Required.** Credential name from root `credentials` |
| `host_key_verification` | String | Verify host key against `known_hosts`. `on` (default) or `off` |
| `trust_on_first_use` | String | Record unknown host key on first connect. `on` or `off` (default) |

```yaml
remote:
  host: web.example.com
  credential: prod-ssh
```

Host key opt-out / TOFU:

```yaml
remote:
  host: web.example.com
  credential: prod-ssh
  host_key_verification: off
  # trust_on_first_use: on
```

By default Pablo verifies the remote host key using `~/.ssh/known_hosts` (Windows: `%USERPROFILE%\.ssh\known_hosts`). See [SSH guide](../guides/ssh.md) and [SECURITY.md](../../SECURITY.md).

---

## Deploy

| Field | Type | Description |
|---|---|---|
| `target_path` | String | **Required.** Path on the target machine (absolute for remote Linux) |
| `source` | [Source](#source) | Artifact location (`static`, `binary` only) |
| `strategy` | String | `overwrite` (default), `backup`, `recreate`, `rename-replace` |
| `transfer` | String | Remote transfer method: `tar` (default) or `legacy` (SCP one-by-one) |
| `verify_checksum` | Boolean | After remote static/binary deploy, verify SHA-256 (default: `false`) |
| `pre_commands` | List<String> | Commands before artifacts are transferred |
| `post_commands` | List<String> | Commands after artifacts are transferred |
| `docker` | [Docker](#docker) | Docker Compose settings (`docker` type only) |

---

## Source

Artifact filtering for `static` and `binary` types. Set on each environment — not inherited.

| Field | Type | Description |
|---|---|---|
| `dir` | String | **Required.** Directory containing artifacts (relative to manifest) |
| `include` | List<String> | Glob patterns to include |
| `exclude` | List<String> | Glob patterns to exclude |

Patterns use gitignore-style semantics (relative to `dir`):

| Pattern | Matches |
|---------|---------|
| `*.exe` | Any file named `*.exe` at any depth |
| `/*.exe` or `./*.exe` | `.exe` files in the artifact root only |
| `dist/*.exe` | One directory level under `dist/` |
| `**/*.exe` | `.exe` files at any depth (explicit recursive) |
| `**/*` | All files (same as omitting `include`) |

```yaml
deploy:
  source:
    dir: ./dist
    include: ["**/*"]
    exclude: ["*.map", "*.log"]
  target_path: /var/www/html
  strategy: backup
```

---

## Docker

For the `docker` profile type.

| Field | Type | Description |
|---|---|---|
| `compose_file` | String | **Required.** Path to docker-compose file (relative to cloned repo) |
| `build` | Boolean | Pass `--build` to `docker compose up` |
| `stop_before_sync` | Boolean | Stop running Compose stack before git sync (default: `true`) |

```yaml
deploy:
  target_path: /opt/app
  docker:
    compose_file: docker-compose.yml
    build: true
    stop_before_sync: true
```

When `stop_before_sync` is omitted or `true`, Pablo runs `docker compose ps -q` before git clone/pull. If containers are present, it runs `docker compose down` (without `-v`) then syncs and brings the stack back up.

---

## RegisterPath

PATH registration for `binary` profiles only.

| Field | Type | Description |
|---|---|---|
| `scope` | String | `user` (default) or `system` |

```yaml
register_path:
  scope: user
```

---

## Examples by type

One complete sample per type. For the full easy→hard ladder (SSH, strategies, sequences, Windows, and more), see [Examples](../examples/README.md).

### `static` — local copy (no build)

```yaml
name: site
version: 0.1.0

profiles:
  default:
    type: static
    environments:
      production:
        deploy:
          source:
            dir: ./src
            include: ["**/*"]
          target_path: ./deploy-output
          strategy: overwrite
```

### `binary` — build, deploy, PATH

```yaml
name: cli
version: 0.1.0

profiles:
  default:
    type: binary
    build:
      command: go build -o mycli .
      path: .
    environments:
      production:
        deploy:
          source:
            dir: .
            include: ["mycli"]
          target_path: /usr/local/bin
          strategy: overwrite
        register_path:
          scope: user
```

### `docker` — git sync + Compose

```yaml
name: stack
version: 0.1.0

credentials:
  github:
    type: token
    value: ghp_xxxxx
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  default:
    type: docker
    git:
      repo: https://github.com/user/app.git
      branch: main
      credential: github
    environments:
      production:
        remote:
          host: docker.example.com
          credential: prod-ssh
        deploy:
          target_path: /opt/app
          docker:
            compose_file: docker-compose.yml
            build: true
```

### `git-sync` — pull + post commands

```yaml
name: api
version: 0.1.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  default:
    type: git-sync
    git:
      repo: https://github.com/user/api.git
      branch: main
    environments:
      production:
        remote:
          host: api.example.com
          credential: prod-ssh
        deploy:
          target_path: /opt/api
          strategy: backup
          post_commands:
            - composer install --no-dev
            - systemctl restart my-api
        variables:
          APP_ENV: production
        env_file: .env
```
