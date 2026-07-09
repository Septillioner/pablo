# Quick Start

Get from zero to a validated manifest in a few minutes.

**Prerequisites:** [Pablo installed](installation.md)

---

## 1. Generate a sample manifest

```bash
pablo init
```

This creates `pablo_sample.yaml` in the current directory. Rename it to `pablo.yaml` or use `-f` to point at another file.

---

## 2. Edit the manifest

Minimal `static` profile for a local deploy:

```yaml
name: my-app
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
    environments:
      production:
        deploy:
          target_path: ./deploy-output
          strategy: overwrite
```

Key fields:

- `output_dir` — where build artifacts live (not under `build:`).
- `deploy.target_path` — where files are copied on the target machine.
- `strategy` — `overwrite`, `backup`, or `recreate`.

Full field reference: [Configuration](../reference/configuration.md).

---

## 3. Validate

```bash
pablo check
```

On success, Pablo prints nothing and exits `0`. On failure, you get `path:line:col` diagnostics:

```
pablo.yaml:12:5: environments.production.deploy.target_path is required
```

---

## 4. Deploy

```bash
pablo run -p default -e production
```

Defaults: profile `default`, environment `production`, manifest `pablo.yaml`.

Pablo runs: hooks → build → filter → deploy → health check (if configured). Watch for log markers: `+` success, `-` error, `>` action.

---

## 5. Inspect profiles

```bash
pablo inspect
pablo inspect --json
```

Lists all profiles and environments — useful for scripting and the VS Code Run picker.

---

## What's next

| Goal | Guide |
|------|-------|
| Walk through a real fixture | [First deployment](first-deployment.md) |
| Deploy over SSH | [SSH guide](../guides/ssh.md) |
| Docker Compose | [Docker guide](../guides/docker.md) |
| Understand manifest layout | [Project structure](project-structure.md) |
