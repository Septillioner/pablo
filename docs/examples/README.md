# Examples

Copy-paste manifests ordered from simplest to more advanced. Each example is a complete `pablo.yaml` for one scenario — run `pablo check` then `pablo run` (or the command shown) from the directory that contains the file.

**Prerequisites:** [Installation](../getting-started/installation.md)

Hands-on walkthrough of example 1: [First deployment](../getting-started/first-deployment.md).

| # | Scenario | What you learn |
|---|----------|----------------|
| 1 | [Local static, no build](#1-local-static-no-build) | Copy files without a build step |
| 2 | [Local static + build](#2-local-static--build) | Optional `build.command` for `static` |
| 3 | [Local binary + PATH](#3-local-binary--path) | `binary` type and `register_path` |
| 4 | [SSH static](#4-ssh-static) | `remote` + named SSH credential |
| 5 | [SSH rename-replace](#5-ssh-rename-replace) | Per-file swap over SSH |
| 6 | [Docker Compose local](#6-docker-compose-local) | `docker` type on your machine |
| 7 | [Docker Compose over SSH](#7-docker-compose-over-ssh) | Compose on a remote host |
| 8 | [Git-sync + post commands](#8-git-sync--post-commands) | Pull source and run install/restart |
| 9 | [Multi-profile monorepo](#9-multi-profile-monorepo) | Several apps in one file |
| 10 | [Separate apps](#10-separate-apps) | One manifest per app |
| 11 | [Sequences](#11-sequences) | Ordered multi-target `pablo run sequence` |
| 12 | [Inheritance](#12-inheritance) | Profile `build` / `variables` / `env_file` |
| 13 | [Credentials](#13-credentials) | `ssh`, `token`, and `basic` side by side |
| 14 | [Transfer + checksum](#14-transfer--checksum) | `deploy.transfer` and `verify_checksum` |
| 15 | [Windows binary + service](#15-windows-binary--service) | Binary deploy with NSSM `post_commands` |
| 16 | [Windows rename-replace](#16-windows-rename-replace) | Local per-file swap on Windows |

---

## 1. Local static, no build

The smallest useful deploy: filter files from a directory and copy them to `target_path`. Omit `build` — Pablo skips that phase.

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

```bash
pablo check
pablo run
```

Fixture: [tests/agnostic/local-deploy](../../tests/agnostic/local-deploy/).

---

## 2. Local static + build

Same copy flow, plus a build command. Artifacts come from the build output directory named in `deploy.source.dir`.

```yaml
name: frontend
version: 0.1.0

profiles:
  default:
    type: static
    build:
      command: npm run build
      path: .
    environments:
      production:
        deploy:
          source:
            dir: ./dist
            include: ["**/*"]
            exclude: ["*.map"]
          target_path: ./deploy-output
          strategy: overwrite
```

```bash
pablo check
pablo run
```

`build` is optional for `static`. Use it when you need compile or bundle; skip it for plain HTML/CSS/assets.

---

## 3. Local binary + PATH

Build an executable, deploy the artifact, and optionally register its directory on PATH.

```yaml
name: mycli
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
          target_path: ./bin
          strategy: overwrite
        register_path:
          scope: user
```

```bash
pablo check
pablo run
```

Guide: [Binary and PATH](../guides/binary-and-path.md). Fixture: [tests/agnostic/self-deploy](../../tests/agnostic/self-deploy/).

---

## 4. SSH static

Same file copy as example 1, but a `remote` block sends files over SSH. `target_path` must be an absolute path on the remote host. Credentials are defined once at the root and referenced by name.

```yaml
name: site
version: 0.1.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  default:
    type: static
    environments:
      production:
        remote:
          host: web.example.com
          credential: prod-ssh
        deploy:
          source:
            dir: ./src
            include: ["**/*"]
          target_path: /var/www/html
          strategy: backup
```

```bash
pablo check -p default -e production
pablo run -p default -e production
```

Add `build` (as in example 2) if the site needs a compile step before transfer.

Guide: [SSH](../guides/ssh.md). E2E fixture: [tests/e2e/scenarios/static-site](../../tests/e2e/scenarios/static-site/).

---

## 5. SSH rename-replace

Use `rename-replace` when you want per-file atomic swaps without renaming the whole target directory. Sibling files that are not in the artifact set stay untouched. On failure, Pablo restores the renamed originals.

```yaml
name: site
version: 0.1.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  default:
    type: static
    environments:
      production:
        remote:
          host: web.example.com
          credential: prod-ssh
        deploy:
          source:
            dir: ./src
            include: ["*.html"]
          target_path: /var/www/html
          strategy: rename-replace
```

```bash
pablo check
pablo run
```

Guide: [Deploy strategies](../guides/deploy-strategies.md). E2E fixture: [tests/e2e/scenarios/static-site-hotfix](../../tests/e2e/scenarios/static-site-hotfix/).

For a clean slate that deletes the whole directory first, use `strategy: recreate` instead. For a timestamped copy of the previous tree, use `backup` (as in example 4).

---

## 6. Docker Compose local

Clone or pull a repo and run `docker compose` on your machine. Omit `remote`. Docker profiles require `git.repo` and `deploy.docker.compose_file`; they do not use `deploy.source`.

```yaml
name: stack
version: 0.1.0

profiles:
  default:
    type: docker
    git:
      repo: https://github.com/user/app.git
      branch: main
    environments:
      production:
        deploy:
          target_path: ./runtime
          docker:
            compose_file: docker-compose.yml
            build: true
```

```bash
pablo check
pablo run
```

Guide: [Docker](../guides/docker.md).

---

## 7. Docker Compose over SSH

Same Compose flow, but Pablo connects over SSH, syncs the git repo on the remote host, and runs `docker compose` there.

```yaml
name: stack
version: 0.1.0

credentials:
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
            stop_before_sync: true
```

```bash
pablo check
pablo run
```

Guide: [Docker](../guides/docker.md) · [SSH](../guides/ssh.md). E2E fixture: [tests/e2e/scenarios/compose-api](../../tests/e2e/scenarios/compose-api/).

---

## 8. Git-sync + post commands

Pull an interpreted app (PHP, Python, Node without a Pablo build) into `target_path`, write env files, then run install and restart commands. No `deploy.source` and no Compose block.

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
    variables:
      APP_NAME: my-api
    env_file: .env
    git:
      repo: https://github.com/user/my-api.git
      branch: main
    environments:
      production:
        variables:
          APP_ENV: production
        remote:
          host: api.example.com
          credential: prod-ssh
        deploy:
          target_path: /opt/my-api
          strategy: backup
          post_commands:
            - composer install --no-dev
            - php artisan migrate --force
            - systemctl restart my-api
```

```bash
pablo check
pablo run
```

Guide: [Git sync](../guides/git-sync.md). Fixture: [tests/agnostic/separate-apps/php-app](../../tests/agnostic/separate-apps/php-app/).

---

## 9. Multi-profile monorepo

One file, several apps — pick a profile with `-p`. Defaults are profile `default` and environment `production` when those names exist.

```yaml
name: monorepo
version: 1.0.0

profiles:
  web:
    type: static
    environments:
      production:
        deploy:
          source:
            dir: ./web/dist
            include: ["**/*"]
          target_path: ./out/web
          strategy: overwrite

  api:
    type: binary
    build:
      command: go build -o api .
      path: ./api
    environments:
      production:
        deploy:
          source:
            dir: ./api
            include: ["api"]
          target_path: ./out/api
          strategy: overwrite
```

```bash
pablo run -p web -e production
pablo run -p api -e production
```

Concepts: [Project structure](../getting-started/project-structure.md). Fixture: [tests/agnostic/multi-profile](../../tests/agnostic/multi-profile/).

---

## 10. Separate apps

Prefer one small manifest per app when repos or release cadences differ. Point Pablo at each file with `-f`, or run from that app’s directory.

```yaml
# frontend/pablo.yaml
name: frontend
version: 1.0.0

profiles:
  default:
    type: static
    environments:
      production:
        deploy:
          source:
            dir: ./dist
            include: ["**/*"]
          target_path: ./deploy-output
          strategy: overwrite
```

```yaml
# api/pablo.yaml
name: api
version: 1.0.0

profiles:
  default:
    type: docker
    git:
      repo: https://github.com/user/api.git
      branch: main
    environments:
      production:
        deploy:
          target_path: ./runtime
          docker:
            compose_file: docker-compose.yml
            build: true
```

```bash
pablo run -f frontend/pablo.yaml
pablo run -f api/pablo.yaml
```

Layout notes: [Project structure](../getting-started/project-structure.md). Fixtures: [tests/agnostic/separate-apps](../../tests/agnostic/separate-apps/).

---

## 11. Sequences

Define a root-level `sequences` map. Each step is `profile/env`. List order is execution order; the first failure stops the rest. You cannot combine `pablo run sequence` with `-p` / `-e`.

```yaml
name: release-bundle
version: 1.0.0

sequences:
  ship:
    - web/production
    - api/production

profiles:
  web:
    type: static
    environments:
      production:
        deploy:
          source:
            dir: ./web/dist
            include: ["**/*"]
          target_path: ./out/web
          strategy: overwrite

  api:
    type: static
    environments:
      production:
        deploy:
          source:
            dir: ./api/dist
            include: ["**/*"]
          target_path: ./out/api
          strategy: overwrite
```

```bash
pablo run sequence ship
```

Guide: [Sequences](../guides/sequences.md) · [CLI — run](../reference/cli.md#run).

---

## 12. Inheritance

Profile-level `build`, `variables`, and `env_file` cascade into every environment. Environment values win on conflict. Artifact paths (`deploy.source`, `target_path`) and `remote` never inherit — set them on each environment.

```yaml
name: inheritance-demo
version: 1.0.0

profiles:
  default:
    type: binary
    build:
      command: go build -o myservice ./cmd/server
      path: .
    variables:
      SERVICE_NAME: myservice
    environments:
      local:
        deploy:
          source:
            dir: .
            include: ["myservice"]
          target_path: ./out/local
          strategy: overwrite
      staging:
        variables:
          SERVICE_NAME: myservice-staging
        deploy:
          source:
            dir: .
            include: ["myservice"]
          target_path: ./out/staging
          strategy: overwrite
```

```bash
pablo run -e local
pablo run -e staging
```

Both environments reuse the profile `build`. Fixture: [tests/agnostic/inheritance-test](../../tests/agnostic/inheritance-test/). Env-file inheritance: [php-app](../../tests/agnostic/separate-apps/php-app/).

---

## 13. Credentials

Name credentials at the root, then reference them from `remote.credential` or `git.credential`. Pablo supports SSH keys/passwords, bearer tokens, and HTTP basic auth.

```yaml
name: creds-demo
version: 1.0.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519
  staging-ssh:
    type: ssh
    username: deploy
    password: "${STAGING_SSH_PASSWORD}"
  github:
    type: token
    value: "${GITHUB_TOKEN}"
  registry:
    type: basic
    username: dockeruser
    password: "${REGISTRY_PASSWORD}"

profiles:
  api:
    type: docker
    git:
      repo: https://github.com/company/api.git
      branch: main
      credential: github
    environments:
      production:
        remote:
          host: api.example.com
          credential: prod-ssh
        deploy:
          target_path: /opt/api
          docker:
            compose_file: docker-compose.yml
            build: true
```

```bash
pablo check -p api -e production
```

Guide: [Credentials](../guides/credentials.md). Fixture with all three types: [tests/agnostic/multi-profile](../../tests/agnostic/multi-profile/).

---

## 14. Transfer + checksum

Remote static and binary deploys stream a tar archive by default (`deploy.transfer: tar`). Use `legacy` for per-file SCP when debugging. Set `verify_checksum: true` to run a SHA-256 check on the remote after transfer.

```yaml
name: site
version: 0.1.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  default:
    type: static
    environments:
      production:
        remote:
          host: web.example.com
          credential: prod-ssh
        deploy:
          source:
            dir: ./dist
            include: ["**/*"]
          target_path: /var/www/html
          strategy: backup
          transfer: tar
          verify_checksum: true
```

```bash
pablo check
pablo run
```

Guide: [SSH — File transfer](../guides/ssh.md#file-transfer-modes).

---

## 15. Windows binary + service

Build a Windows executable, deploy it locally, then install or restart a service with `post_commands` (NSSM in this fixture). Adjust paths and commands for your service manager.

```yaml
name: win-service
version: 1.0.0

profiles:
  default:
    type: binary
    environments:
      local:
        build:
          path: ./src
          command: go build -o ../dist/myservice.exe .
        deploy:
          source:
            dir: ./dist
            include: ["myservice.exe"]
          target_path: "C:\\PabloTest\\Service"
          strategy: overwrite
          post_commands:
            - "nssm install PabloTestService C:\\PabloTest\\Service\\myservice.exe || echo already installed"
            - "nssm set PabloTestService AppDirectory C:\\PabloTest\\Service"
            - "nssm start PabloTestService || nssm restart PabloTestService"
```

```bash
pablo check -e local
pablo run -e local
```

Fixture: [tests/windows/nssm-service](../../tests/windows/nssm-service/).

---

## 16. Windows rename-replace

Local static deploy on Windows using per-file `rename-replace`. Same strategy semantics as example 5; paths use Windows conventions.

```yaml
name: win-rename-replace
version: 1.0.0

profiles:
  default:
    type: static
    environments:
      local:
        deploy:
          source:
            dir: ./src
            include: ["*.txt"]
          target_path: "C:\\PabloTest\\RenameReplace"
          strategy: rename-replace
```

```bash
pablo check -e local
pablo run -e local
```

Fixture: [tests/windows/rename-replace](../../tests/windows/rename-replace/).

---

## Related

| Topic | Page |
|-------|------|
| Field reference | [Configuration](../reference/configuration.md) |
| What works today | [Capabilities](../reference/capabilities.md) |
| Deploy strategies | [Deploy strategies](../guides/deploy-strategies.md) |
| SSH remote deploy | [SSH](../guides/ssh.md) |
| Git-based apps | [Git sync](../guides/git-sync.md) |
