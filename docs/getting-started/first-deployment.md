# First Deployment

Walk through a local static deploy end to end — HTML from `./src` into `./dist`, with no build step and no remote server. The same shape lives in the repository fixture [tests/agnostic/local-deploy](../../tests/agnostic/local-deploy/).

**Prerequisites:** [Installation](installation.md) · [Quick start](quick-start.md)

---

## Step 1 — Create the project

```bash
mkdir my-first-deploy && cd my-first-deploy
```

Create `src/index.html`:

```html
<!DOCTYPE html>
<html>
<head><title>Hello Pablo</title></head>
<body><h1>Deployed with Pablo</h1></body>
</html>
```

---

## Step 2 — Write `pablo.yaml`

```yaml
name: local-test-app
version: 0.0.1

profiles:
  local-test:
    type: static
    environments:
      dev:
        deploy:
          source:
            dir: ./src
            include: ["*.html"]
          target_path: ./dist
          strategy: overwrite
```

For `static` profiles, every environment needs its own `deploy.source`. The glob `*.html` matches HTML files by basename at any depth; use `/*.html` for the artifact root only. `target_path: ./dist` is relative to the manifest directory.

---

## Step 3 — Validate

```bash
pablo check -f pablo.yaml -p local-test -e dev
```

Fix any reported errors before continuing.

---

## Step 4 — Run

```bash
pablo run -p local-test -e dev -f pablo.yaml
```

Pablo creates or overwrites `./dist/` and places `index.html` there. Confirm with:

```bash
# macOS / Linux
ls ./dist
cat ./dist/index.html

# Windows
dir dist
type dist\index.html
```

---

## Step 5 — Re-deploy with backup

Change `strategy` to `backup` and run again. Pablo renames the existing `./dist` to a timestamped directory (for example `./dist.20260709-143022`) before writing fresh files — a local rollback copy you can rename back by hand if needed.

```yaml
deploy:
  target_path: ./dist
  strategy: backup
```

---

## Step 6 — Clean up

```bash
pablo uninstall -p local-test -e dev -f pablo.yaml
```

Uninstall removes the deployed directory (and PATH entries for `binary` profiles). You can also delete `./dist` manually.

---

## Try the repo fixture

```bash
cd tests/agnostic/local-deploy
pablo check -f pablo.yaml -p local-test -e dev
pablo run -p local-test -e dev -f pablo.yaml
```

---

## Next steps

| Topic | Guide |
|-------|-------|
| Progressive scenarios | [Examples](../examples/README.md) |
| Remote server deploy | [SSH](../guides/ssh.md) |
| Compiled binary + PATH | [Binary and PATH](../guides/binary-and-path.md) |
| Docker Compose | [Docker](../guides/docker.md) |
| Strategies in depth | [Deploy strategies](../guides/deploy-strategies.md) |
