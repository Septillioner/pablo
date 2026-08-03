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
| `build` | Copied when env has no `build`; partial merge when env `build` exists (including `build.variables` / `build.env_file`) |

`deploy.source`, `deploy.target_path`, and `remote` are always set on the environment — they do not inherit.

---

## Variables and env files

Pablo **writes** dotenv-style files from YAML maps. It does **not** load a pre-existing `.env` from disk into the Pablo process.

**Canonical place for values:** environment (and inherited profile) `variables`. Put Vite/`VITE_*` maps and other app config there — not under `build.variables` unless you need a build-only override.

| Scope | Map field | File field | When written | Also injected into |
|-------|-----------|------------|--------------|--------------------|
| Build | Environment `variables` (optional `build.variables` overlay) | `build.env_file` | Before `build.command`, under `build.path` | Build command process environment |
| Deploy | `variables` (profile → env merge) | `env_file` | Into the deploy `target_path` (local or remote) | Template `{{VAR}}` substitution (non-docker) |

Rules:

1. **Write-only** — `env_file` / `build.env_file` are output paths. Committing a hand-edited `.env` does not feed Pablo; put values in YAML `variables` (or a gitignored override manifest). Optional `build.variables` only overlay build-time keys.
2. **Empty map skips write** — If the resolved variables map is empty, Pablo does not create or overwrite the file even when `env_file` / `build.env_file` is set.
3. **Path** — Relative `build.env_file` resolves under `build.path`; relative deploy `env_file` under `deploy.target_path`. Absolute paths are used as-is. On Windows, a leading `/` (e.g. `/test/.env.production`) is absolute (drive-root), not under `build.path` — prefer `.env.production`.
4. **Inheritance** — Profile `variables` / `env_file` merge into each environment; profile `build` (including its `variables` / `env_file`) merges as described above. Environment values win on key conflicts.
5. **Pre-build write is intentional** — Writing `build.env_file` before `build.command` is required for tools that read dotenv at compile time (e.g. Vite). It does not replace the post-deploy `env_file` write.

```yaml
name: frontend-app
version: 1.0.0

profiles:
  frontend:
    type: static
    build:
      command: npm run build
      path: ./frontend
      # Pre-build output under build.path (from environment variables below)
      env_file: .env.production
    environments:
      production:
        variables:
          VITE_API_BASE_URL: https://api.example.com
          VITE_ENV: production
        # Optional: also write the same map into the deploy target after deploy
        # env_file: .env
        deploy:
          source:
            dir: ./frontend/dist
            include: ["**/*"]
          target_path: /var/www/frontend
          strategy: overwrite
```

With that manifest, Pablo writes `./frontend/.env.production` from environment `variables`, injects those keys into the `npm run build` process environment, then deploys artifacts. Optional `build.variables` keys override the same name for the build write and process env only. Full copy-paste: [Examples #2](../examples/README.md#2-local-static--build).

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
| `variables` | Map<String, String> | Deploy/runtime variables inherited by all environments (see [Variables and env files](#variables-and-env-files)) |
| `env_file` | String | Default deploy env file name inherited when an environment omits `env_file` |
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
| `variables` | Map<String, String> | Optional build-only overlay — merged on top of environment `variables` for the pre-build file and build process env |
| `env_file` | String | Dotenv file to **write** under `build.path` before building (from environment `variables`, plus optional `build.variables` overlay) |

Prefer environment `variables` as the source of truth; use `build.env_file` only to name the pre-build output. See [Variables and env files](#variables-and-env-files).

```yaml
build:
  command: npm run build
  path: ./frontend
  env_file: .env.production
  # Optional overlay only — put VITE_* under environments.<name>.variables
  # variables:
  #   NODE_ENV: production
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
| `variables` | Map<String, String> | Canonical variable map (merged with profile) — feeds deploy `env_file`, and when `build.env_file` is set also the pre-build file + build process env |
| `env_file` | String | Dotenv file **written** into the deploy target from merged `variables` (skipped if the map is empty) |
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
| `target_path` | String | **Required.** Path on the target machine (absolute for remote Linux). With `blue_green`: stable live reference (not written to; Pablo does not create it) |
| `source` | [Source](#source) | Artifact location (`static`, `binary` only) |
| `strategy` | String | `overwrite` (default), `backup`, `recreate`, `rename-replace`. With `blue_green`, default is `recreate` (`backup` is rejected) |
| `transfer` | String | Remote transfer method: `tar` (default) or `legacy` (SCP one-by-one) |
| `verify_checksum` | Boolean | After remote static/binary deploy, verify SHA-256 (default: `false`) |
| `pre_commands` | List<String> | Commands before artifacts are transferred |
| `post_commands` | List<String> | Commands after artifacts are transferred |
| `docker` | [Docker](#docker) | Docker Compose settings (`docker` type only) |
| `blue_green` | [Blue-Green](#blue-green) | Slot-based deploy (`static` / `binary` only) |

---

## Blue-Green

Slot-based deploy for `static` and `binary`. Pablo detects the active slot, writes artifacts into the idle slot, then runs your switch command. Pablo does **not** own the traffic-switching mechanism (symlink, systemd, reverse proxy, etc.) — that stays in your commands.

| Field | Type | Description |
|---|---|---|
| `slots` | List\<Object\> | **Required.** Exactly two entries. Each has `path` (required), optional `key`, optional `switch_command` |
| `detect_command` | String | **Required.** Command whose stdout matches a slot key (`key` if set, else `path`) |
| `switch_command` | String | Default switch command when a slot omits `switch_command` |

### Slot fields

| Field | Type | Description |
|---|---|---|
| `path` | String | **Required.** Directory Pablo writes this slot into |
| `key` | String | Value `detect_command` returns for this slot. Defaults to `path`. Use when the target names the slot differently than Pablo writes it |
| `switch_command` | String | Switch traffic to this slot; overrides `blue_green.switch_command` |

### Detect contract

- Runs before `pre_commands` on the target machine (SSH when `remote` is set).
- Local commands use the manifest directory as cwd so relative paths resolve against the project.
- Local `switch_command` uses the same cwd (manifest directory).
- Remote: stdout only (stderr ignored). Exit code must be 0.
- Trimmed stdout must **exactly** equal one slot's effective key (`key` if set, else `path`), or be empty.

| `detect_command` stdout | Behavior |
|---|---|
| Empty / whitespace | No active slot — deploy to `slots[0]` |
| Exact match of a slot key | Deploy to the other slot |
| Unmatched value or multiple lines | Hard error |
| Non-zero exit | Hard error |

### Path roles when `blue_green` is set

| Role | Path |
|---|---|
| Artifact write, `env_file`, templates, checksum | Selected (idle) slot |
| `pre_commands` / `post_commands` cwd | Selected slot |
| `detect_command` / `switch_command` cwd (local) | Manifest directory |
| `switch_command` cwd (remote) | `target_path` if it exists as a directory; otherwise no `cd` |
| `register_path` | `target_path` |
| `uninstall` | Removes `target_path` and both slots |

Pablo never creates `target_path` under blue-green (avoids turning a planned symlink path into a real directory).

### Command environment

Injected only into command execution (not into `vars` / `env_file` / `{{VAR}}`):

| Variable | Value |
|---|---|
| `PABLO_TARGET_SLOT` | Slot path written this run |
| `PABLO_PREVIOUS_SLOT` | Detected active slot path; empty on first deploy |

Slot-level `switch_command` overrides `blue_green.switch_command`. Every slot must resolve a switch command (own or global).

```yaml
deploy:
  source:
    dir: ./dist
    include: ["**/*"]
  target_path: /var/www/app
  strategy: recreate
  blue_green:
    slots:
      - path: /var/www/app-blue
      - path: /var/www/app-green
    detect_command: cat /var/www/app/.active 2>/dev/null || true
    switch_command: >
      echo "$PABLO_TARGET_SLOT" > /var/www/app/.active &&
      ln -sfn "$PABLO_TARGET_SLOT" /var/www/app/current &&
      systemctl reload nginx
```

Guide: [Blue-green](../guides/blue-green.md).

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
