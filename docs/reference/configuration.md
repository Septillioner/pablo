# Configuration Reference

Complete reference for Pablo's deployment configuration file (`pablo.yaml`).

See also: [Capabilities](capabilities.md) · [Credentials guide](../guides/credentials.md)

---

## Inheritance

Profile-level settings cascade into each environment unless overridden:

| Profile field | Inherited as |
|---------------|--------------|
| `variables` | Merged into environment variables |
| `env_file` | Default env file name for environments |
| `build` | Copied to environment when env has no `build`; partial field merge when env `build` exists |
| `output_dir` | Becomes `env.deploy.source` when environment has no `source` |

Environment `variables` are merged into `deploy.variables`.

**Legacy format:** If `profiles` is omitted and a top-level `type` field exists, Pablo auto-wraps the config into `profiles.default`.

---

## Root Fields

| Field | Type | Description |
|---|---|---|
| `name` | String | Project name |
| `version` | String | Project version |
| `credentials` | Map<String, [Credential](#credential)> | Global reusable credentials (optional) |
| `sequences` | Map<String, String[]> | Named ordered deployment sequences (optional); see [Sequences](#sequences) |
| `profiles` | Map<String, [Profile](#profile)> | Application profiles |

---

## Sequences

Named lists of `profile/environment` targets to run in order. **List order is execution order** — Pablo runs each step sequentially and stops on the first failure.

| Field | Type | Description |
|---|---|---|
| `<name>` | String[] | Ordered steps; each item is `profile/env` (e.g. `extension/vsix`) |

**Example:**

```yaml
sequences:
  extension:
    - extension/vsix
    - extension/marketplace
```

Run with:

```bash
pablo run sequence extension
```

Cannot combine `pablo run sequence` with `-p` / `-e`. Global flags (`-f`, `--force`, `--verbose`) apply to every step.

Guide: [Sequences](../guides/sequences.md).

---

## Credential

Reusable credentials for SSH, Git, Docker registries, etc.

| Field | Type | Description |
|---|---|---|
| `type` | String | **Required.** Credential type: `ssh`, `token`, `basic` |
| `username` | String | Username (for `ssh`, `basic`) |
| `password` | String | Password (for `basic`, `ssh` password auth) |
| `key` | String | SSH private key path (for `ssh`) |
| `passphrase` | String | SSH key passphrase (optional) |
| `value` | String | Token value (for `token`) |

**Example:**
```yaml
credentials:
  prod_server:
    type: ssh
    username: deploy
    key: ~/.ssh/id_rsa
  github:
    type: token
    value: ghp_xxxxx
```

---

## Profile

A complete application configuration.

| Field | Type | Description |
|---|---|---|
| `type` | String | **Required.** Deployment type: `static`, `binary`, `docker`, `git-sync` |
| `build` | [Build](#build) | Build configuration (inherited by environments unless overridden) |
| `git` | [Git](#git) | Git repository config (for `docker`, `git-sync`) |
| `output_dir` | [OutputDir](#outputdir) | Artifact location and filtering rules |
| `environments` | Map<String, [Environment](#environment)> | **Required.** Deployment targets |
| `hooks` | [Hooks](#hooks) | Lifecycle hooks |
| `pipeline` | [Pipeline](#pipeline) | Pipeline settings |
| `variables` | Map<String, String> | Variables inherited by all environments |
| `env_file` | String | Env file name inherited by all environments |

---

## Build

Build configuration. Can be defined at profile level (inherited) or overridden per environment.

For `static` profiles, `build` is **optional**. Omit it to copy files from `output_dir` / `deploy.source` without a compile step.

| Field | Type | Description |
|---|---|---|
| `command` | String | **Required when `build` is set.** Build command (e.g., `npm run build`, `go build -o app .`) |
| `path` | String | Working directory for the build command (relative to manifest) |
| `variables` | Map<String, String> | Environment variables injected during build |
| `env_file` | String | File to write variables to before building |

**Example:**
```yaml
build:
  command: npm run build
  path: ./frontend
  variables:
    NODE_ENV: production
```

---

## Git

Git repository configuration for `docker` and `git-sync` types.

| Field | Type | Description |
|---|---|---|
| `repo` | String | **Required.** Git repository URL |
| `branch` | String | Branch name (default: `main`) |
| `credential` | String | Credential reference (optional) |

**Example:**
```yaml
git:
  repo: https://github.com/user/project.git
  branch: main
  credential: github
```

---

## OutputDir

Artifact location and file filtering configuration. Can be a simple string (directory path) or an object.

| Field | Type | Description |
|---|---|---|
| `dir` | String | Directory containing build artifacts (relative to manifest) |
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

**Simple form:**
```yaml
output_dir: ./dist
```

**Object form:**
```yaml
output_dir:
  dir: ./dist
  include: ["**/*"]
  exclude: ["*.map", "*.log"]
```

---

## Environment

Deployment target configuration.

| Field | Type | Description |
|---|---|---|
| `deploy` | [Deploy](#deploy) | **Required.** Deployment settings |
| `remote` | [Remote](#remote) | Remote server connection (enables SSH deployment) |
| `build` | [Build](#build) | Override profile-level build settings |
| `variables` | Map<String, String> | Runtime variables for this environment (merged with profile variables) |
| `env_file` | String | Env file name (inherited from profile if not set) |
| `register_path` | [RegisterPath](#registerpath) | PATH registration (binary type only) |

---

## Remote

Remote server connection configuration. When present, deployment targets the remote host via SSH.

| Field | Type | Description |
|---|---|---|
| `method` | String | **Required.** Connection method: `ssh` |
| `host` | String | **Required.** Remote host address (port defaults to 22) |
| `credential` | String | **Required.** Credential reference name |
| `host_key_verification` | String | Host key check against OpenSSH `known_hosts`. `on` (default) or `off` |
| `trust_on_first_use` | String | When `on`, record an unknown host key on first connect. `on` or `off` (default). Only applies when verification is on |

**Example:**
```yaml
remote:
  method: ssh
  host: 192.168.1.100
  credential: prod_server
```

**Host key verification (opt-out / TOFU):**
```yaml
remote:
  method: ssh
  host: 192.168.1.100
  credential: prod_server
  host_key_verification: off   # not recommended; emits a warning
  # trust_on_first_use: on     # optional; default is off
```

By default Pablo verifies the remote host key using `~/.ssh/known_hosts` (Windows: `%USERPROFILE%\.ssh\known_hosts`). See [SSH guide](../guides/ssh.md) and [SECURITY.md](../../SECURITY.md).

---

## Deploy

Deployment method and settings.

| Field | Type | Description |
|---|---|---|
| `target_path` | String | **Required.** Absolute path on the target machine |
| `strategy` | String | Strategy: `overwrite` (default), `backup`, `recreate`, `rename-replace` |
| `remote` | String | Transfer method: `tar` (default, high performance) or `legacy` (SCP one-by-one) |
| `source` | [Source](#source) | Override profile-level artifact settings for this environment |
| `docker` | [Docker](#docker) | Docker config (for `docker` type) |
| `service` | [Service](#service) | Service management (schema only — not implemented at runtime) |
| `pre_commands` | List<String> | Commands to run before artifacts are deployed |
| `post_commands` | List<String> | Commands to run after artifacts are deployed |
| `variables` | Map<String, String> | Deploy-level variables (merged from environment) |
| `env_file` | String | Generate an env file at this relative path inside `target_path` |

---

## Source

Override artifact settings at the deploy level (takes precedence over profile `output_dir`).

| Field | Type | Description |
|---|---|---|
| `dir` | String | Artifact directory |
| `include` | List<String> | Glob patterns to include (see [OutputDir](#outputdir) semantics) |
| `exclude` | List<String> | Glob patterns to exclude (see [OutputDir](#outputdir) semantics) |

**Example:**
```yaml
deploy:
  source:
    dir: ./build
    include: ["pablo"]
    exclude: ["*.tmp"]
  target_path: /opt/app
```

---

## Docker

Docker deployment configuration.

| Field | Type | Description |
|---|---|---|
| `compose_file` | String | **Required.** Path to docker-compose file |
| `build` | Boolean | Build images before up |
| `command` | String | Docker compose command (default: `up -d`) |

**Example:**
```yaml
docker:
  compose_file: docker-compose.yml
  build: true
  command: up -d --build
```

---

## Service

Service management configuration. **Schema only** — not executed at runtime. Use `post_commands` (e.g. `systemctl restart myapp`) until service management ships.

| Field | Type | Description |
|---|---|---|
| `type` | String | **Required.** Service type: `systemd`, `pm2` |
| `name` | String | **Required.** Service name |
| `restart` | Boolean | Restart after deployment |

---

## RegisterPath

PATH registration (binary type only).

| Field | Type | Description |
|---|---|---|
| `scope` | String | Scope: `user` (default), `system` |

---

## Hooks

Lifecycle hooks executed at the profile level (before/after the entire deployment).

| Field | Type | Description |
|---|---|---|
| `pre` | String | Command before deployment |
| `post` | String | Command after deployment |

---

## Pipeline

Pipeline-wide settings.

| Field | Type | Description |
|---|---|---|
| `on_success` | String | Command on success |
| `on_failure` | String | Command on failure |
| `health_check` | String | Health check URL (HTTP GET, retries for 30s) |

---

## Deployment Types

### `static` — Files / frontend / SPA

Optional build → Filter artifacts → Deploy files

**Required:** `output_dir` or `deploy.source`, `environments.deploy.target_path`

`build` is optional. Omit it to copy existing files (HTML, assets, pre-built `dist`, etc.).

### `binary` — Compiled Executables
Build → Deploy binary → Register PATH

**Required:** `build`, `environments.deploy.target_path`

### `docker` — Containerized Services
Git clone/pull → Generate env file → Docker compose up

**Required:** `git`, `environments.deploy.docker`

### `git-sync` — Interpreted Languages
Git pull → Generate env file → Run post commands

**Required:** `git`, `environments.deploy.target_path`

---

## Examples (easy → hard)

Progressive copy-paste samples live in [Examples](../examples/README.md). Minimal patterns below.

### Local copy (no build)

```yaml
name: site
version: 0.1.0

profiles:
  default:
    type: static
    output_dir:
      dir: ./src
      include: ["**/*"]
    environments:
      production:
        deploy:
          target_path: ./deploy-output
          strategy: overwrite
```

### Local build then copy

```yaml
profiles:
  default:
    type: static
    build:
      command: npm run build
      path: .
    output_dir:
      dir: ./dist
      include: ["**/*"]
    environments:
      production:
        deploy:
          target_path: ./deploy-output
          strategy: overwrite
```

### SSH static

```yaml
credentials:
  server-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_rsa

profiles:
  frontend:
    type: static
    output_dir:
      dir: ./dist
      include: ["**/*"]
    environments:
      production:
        remote:
          method: ssh
          host: web.example.com
          credential: server-ssh
        deploy:
          target_path: /var/www/html
          strategy: backup
```

### Full multi-profile (kitchen sink)

```yaml
name: my-app
version: 1.0.46
credentials:
  server-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_rsa

profiles:
  frontend:
    type: static
    build:
      command: npm run build
      path: ./frontend
    output_dir:
      dir: ./frontend/dist
      include: ["**/*"]
      exclude: ["*.map"]
    hooks:
      pre: echo "Starting frontend deploy"
    environments:
      production:
        remote:
          method: ssh
          host: web.example.com
          credential: server-ssh
        deploy:
          target_path: /var/www/html
          strategy: backup

  api:
    type: binary
    build:
      command: go build -o api-server .
      path: ./backend
    environments:
      production:
        remote:
          method: ssh
          host: api.example.com
          credential: server-ssh
        deploy:
          source:
            dir: ./backend
            include: ["api-server"]
          target_path: /opt/api
          strategy: backup
          post_commands:
            - systemctl daemon-reload
            - systemctl restart api-server
        variables:
          APP_ENV: production
          DB_HOST: db.internal
        register_path:
          scope: system
    pipeline:
      health_check: https://api.example.com/health
      on_failure: "echo 'Deploy failed!' | mail admin@example.com"
```
