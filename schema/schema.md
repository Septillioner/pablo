# pablo.yaml Schema Documentation

Complete reference for Pablo's deployment configuration file.

## Root Fields

| Field | Type | Description |
|---|---|---|
| `name` | String | Project name |
| `version` | String | Project version |
| `credentials` | Map<String, [Credential](#credential)> | Global reusable credentials (optional) |
| `profiles` | Map<String, [Profile](#profile)> | Application profiles |

> **Backward compatibility:** If `profiles` is omitted and a top-level `type` field exists, Pablo auto-wraps the config into a `profiles.default` profile.

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

| Field | Type | Description |
|---|---|---|
| `command` | String | **Required.** Build command (e.g., `npm run build`, `go build -o app .`) |
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

**Example:**
```yaml
remote:
  method: ssh
  host: 192.168.1.100
  credential: prod_server
```

---

## Deploy

Deployment method and settings.

| Field | Type | Description |
|---|---|---|
| `target_path` | String | **Required.** Absolute path on the target machine |
| `strategy` | String | Strategy: `overwrite` (default), `backup`, `recreate` |
| `remote` | String | Transfer method: `tar` (default, high performance) or `legacy` (SCP one-by-one) |
| `source` | [Source](#source) | Override profile-level artifact settings for this environment |
| `docker` | [Docker](#docker) | Docker config (for `docker` type) |
| `service` | [Service](#service) | Service management (for `binary` type) |
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
| `include` | List<String> | Glob patterns to include |
| `exclude` | List<String> | Glob patterns to exclude |

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

Service management configuration.

| Field | Type | Description |
|---|---|---|
| `type` | String | **Required.** Service type: `systemd`, `pm2` |
| `name` | String | **Required.** Service name |
| `restart` | Boolean | Restart after deployment |

**Example:**
```yaml
service:
  type: systemd
  name: myapp
  restart: true
```

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

### `static` — Frontend / SPA
Build -> Filter artifacts -> Deploy files

**Required:** `output_dir` or `deploy.source`, `environments.deploy.target_path`

### `binary` — Compiled Executables
Build -> Deploy binary -> Register PATH -> Restart service

**Required:** `build`, `environments.deploy.target_path`

### `docker` — Containerized Services
Git clone/pull -> Generate env file -> Docker compose up

**Required:** `git`, `environments.deploy.docker`

### `git-sync` — Interpreted Languages
Git pull -> Generate env file -> Run post commands

**Required:** `git`, `environments.deploy.target_path`

---

## Full Example

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
          service:
            type: systemd
            name: api-server
            restart: true
          post_commands:
            - systemctl daemon-reload
        variables:
          APP_ENV: production
          DB_HOST: db.internal
        register_path:
          scope: system
    pipeline:
      health_check: https://api.example.com/health
      on_failure: "echo 'Deploy failed!' | mail admin@example.com"
```

---

## CLI Usage

```bash
# Deploy frontend to production
pablo run -p frontend -e production

# Deploy with custom manifest
pablo run -p api -e production -f deploy.yaml

# Validate configuration
pablo check -f pablo.yaml -p frontend -e production

# Generate sample config
pablo init

# Remove deployed files
pablo uninstall -p api -e production --remove-backups
```
