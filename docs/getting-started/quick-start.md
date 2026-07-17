# Quick Start

Get from zero to a validated deploy in a few minutes. This page stays with the simplest case — a local static copy — then points you at richer scenarios.

**Prerequisites:** [Pablo installed](installation.md)

---

## 1. Generate a sample manifest

```bash
pablo init
```

This creates `pablo_sample.yaml` in the current directory. Rename it to `pablo.yaml` or pass `-f` when you run commands.

To pick a deployment type interactively:

```bash
pablo init --template
```

Choose from `static`, `binary`, `docker`, or `git-sync`. The wizard needs an interactive terminal.

---

## 2. Start simple: copy files (no build)

The smallest useful `static` profile copies existing files. Omit `build` — Pablo skips that phase and deploys from `deploy.source`.

```yaml
name: my-app
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

`deploy.source` selects files (`dir`, `include`, `exclude`). `deploy.target_path` is where they land. `strategy` may be `overwrite`, `backup`, `recreate`, or `rename-replace`.

Put something in `./src` (for example `index.html`), then continue.

---

## 3. Validate

```bash
pablo check
```

On success Pablo prints nothing and exits `0`. On failure you get `path:line:col` diagnostics:

```
pablo.yaml:12:5: environments.production.deploy.target_path is required
```

---

## 4. Deploy

```bash
pablo run -p default -e production
```

Defaults when omitted: profile `default`, environment `production`, manifest `pablo.yaml`. Log markers are `+` success, `-` error, `>` action, `*` info, `!` warn.

---

## 5. Optional: add a build

When you need compile or bundle before copy, add a profile-level `build` and point `deploy.source.dir` at the output:

```yaml
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
          target_path: ./deploy-output
          strategy: overwrite
```

That is [Examples #2](../examples/README.md#2-local-static--build). For SSH, Docker, git-sync, sequences, and Windows scenarios, use the full [Examples](../examples/README.md) catalog.

---

## 6. Inspect profiles

```bash
pablo inspect
pablo inspect --json
```

Lists every profile and environment — useful for scripting and the editor Run picker.

---

## What's next

| Goal | Guide |
|------|-------|
| Step-by-step local walkthrough | [First deployment](first-deployment.md) |
| All scenarios (easy → hard) | [Examples](../examples/README.md) |
| Deploy over SSH | [SSH guide](../guides/ssh.md) |
| Docker Compose | [Docker guide](../guides/docker.md) |
| Manifest layout | [Project structure](project-structure.md) |
| Full field reference | [Configuration](../reference/configuration.md) |
