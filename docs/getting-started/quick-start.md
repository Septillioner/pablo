# Quick Start

Get from zero to a validated deploy in a few minutes — starting with the simplest case.

**Prerequisites:** [Pablo installed](installation.md)

---

## 1. Generate a sample manifest

```bash
pablo init
```

This creates `pablo_sample.yaml` in the current directory. Rename it to `pablo.yaml` or use `-f` to point at another file.

To pick a deployment type interactively:

```bash
pablo init --template
```

Choose from `static`, `binary`, `docker`, or `git-sync`. The wizard requires an interactive terminal.

---

## 2. Start simple: copy files (no build)

The smallest useful `static` profile copies existing files. Omit `build` — Pablo skips that phase.

```yaml
name: my-app
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

Key fields:

- `output_dir` — which files to deploy (not under `build:`).
- `deploy.target_path` — where files are copied.
- `strategy` — `overwrite`, `backup`, `recreate`, or `rename-replace`.

Put something in `./src` (for example `index.html`), then continue.

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

Watch for log markers: `+` success, `-` error, `>` action.

---

## 5. Next: add a build (optional)

When you need compile/bundle before copy:

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

---

## 6. Inspect profiles

```bash
pablo inspect
pablo inspect --json
```

Lists all profiles and environments — useful for scripting and the VS Code Run picker.

---

## What's next

| Goal | Guide |
|------|-------|
| Step-by-step local walkthrough | [First deployment](first-deployment.md) |
| More examples (easy → hard) | [Examples](../examples/README.md) |
| Deploy over SSH | [SSH guide](../guides/ssh.md) |
| Docker Compose | [Docker guide](../guides/docker.md) |
| Manifest layout | [Project structure](project-structure.md) |
| Full field reference | [Configuration](../reference/configuration.md) |
