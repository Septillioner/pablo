# Examples (easy → hard)

Short, copy-paste manifests ordered from simplest to more advanced. Each step adds one concept.

**Prerequisites:** [Installation](../getting-started/installation.md)

| # | Example | What you learn |
|---|---------|----------------|
| 1 | [Copy files locally](#1-copy-files-locally-no-build) | `static` without `build` |
| 2 | [Build then copy](#2-build-then-copy-local) | Optional `build.command` |
| 3 | [Binary + PATH](#3-binary--path-local) | `binary` type |
| 4 | [SSH static](#4-ssh-static-remote) | `remote` + credentials |
| 5 | [Docker Compose](#5-docker-compose) | `docker` type |
| 6 | [Multi-profile](#6-multi-profile-one-manifest) | Several apps in one file |
| 7 | [Sequences](#7-run-targets-in-order-sequences) | Ordered multi-target `pablo run sequence` |

Hands-on walkthrough of #1: [First deployment](../getting-started/first-deployment.md).

---

## 1. Copy files locally (no build)

Point `output_dir` at existing files. Omit `build` — Pablo skips the build phase and copies filtered files to `target_path`.

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

```bash
pablo check
pablo run
```

Fixture: [tests/agnostic/local-deploy](../../tests/agnostic/local-deploy/).

---

## 2. Build then copy (local)

Same as #1, plus a build command. Artifacts come from the build output directory.

```yaml
name: frontend
version: 0.1.0

profiles:
  default:
    type: static
    build:
      command: npm run build
      path: .
    output_dir:
      dir: ./dist
      include: ["**/*"]
      exclude: ["*.map"]
    environments:
      production:
        deploy:
          target_path: ./deploy-output
          strategy: overwrite
```

`build` is optional for `static`. Use it when you need compile/bundle; skip it for plain HTML/CSS/assets.

---

## 3. Binary + PATH (local)

Build an executable, deploy it, optionally register the directory on PATH.

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

Guide: [Binary and PATH](../guides/binary-and-path.md).

---

## 4. SSH static (remote)

Same file copy as #1, but `remote` sends files over SSH. `target_path` must be an absolute path on the remote host.

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
    output_dir:
      dir: ./src
      include: ["**/*"]
    environments:
      production:
        remote:
          method: ssh
          host: web.example.com
          credential: prod-ssh
        deploy:
          target_path: /var/www/html
          strategy: backup
```

Add `build` (as in #2) if the site needs a compile step before transfer.

Guide: [SSH](../guides/ssh.md). E2E fixture: [tests/e2e/scenarios/ssh-static](../../tests/e2e/scenarios/ssh-static/).

---

## 5. Docker Compose

Clone/pull a repo and run `docker compose` (local or remote).

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

Guide: [Docker](../guides/docker.md).

---

## 6. Multi-profile (one manifest)

One file, several apps — pick with `-p`.

```yaml
name: monorepo
version: 1.0.0

profiles:
  web:
    type: static
    output_dir:
      dir: ./web/dist
      include: ["**/*"]
    environments:
      production:
        deploy:
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

## 7. Run targets in order (sequences)

Define a root-level `sequences` list. Each step is `profile/env`. List order is execution order; the first failure stops the rest.

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
    output_dir:
      dir: ./web/dist
      include: ["**/*"]
    environments:
      production:
        deploy:
          target_path: ./out/web
          strategy: overwrite

  api:
    type: static
    output_dir:
      dir: ./api/dist
      include: ["**/*"]
    environments:
      production:
        deploy:
          target_path: ./out/api
          strategy: overwrite
```

```bash
pablo run sequence ship
```

Reference: [Sequences guide](../guides/sequences.md) · [Configuration — Sequences](../reference/configuration.md#sequences) · [CLI — run](../reference/cli.md#run).

---

## Related

| Topic | Page |
|-------|------|
| Field reference | [Configuration](../reference/configuration.md) |
| What works today | [Capabilities](../reference/capabilities.md) |
| Deploy strategies | [Deploy strategies](../guides/deploy-strategies.md) |
| Git-based apps | [Git sync](../guides/git-sync.md) |
