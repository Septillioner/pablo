# First Deployment

A hands-on local static deploy using the repository test fixture.

**Prerequisites:** [Installation](installation.md) · [Quick start](quick-start.md)

---

## Scenario

Deploy HTML files from `./src` to `./dist` on your machine — no build step, no remote server.

This matches [tests/agnostic/local-deploy/pablo.yaml](../../tests/agnostic/local-deploy/pablo.yaml).

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

    output_dir:
      dir: ./src
      include: ["*.html"]

    environments:
      dev:
        deploy:
          target_path: ./dist
          strategy: overwrite
```

Notes:

- `output_dir` at profile level becomes the deploy source (inheritance).
- `include: ["*.html"]` filters to HTML files only.
- `target_path: ./dist` is relative to the manifest directory.

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

Expected outcome:

- `./dist/` is created (or overwritten).
- `index.html` appears in `./dist/`.

Verify:

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

Change `strategy` to `backup` and run again:

```yaml
deploy:
  target_path: ./dist
  strategy: backup
```

Pablo renames the existing `./dist` to something like `./dist.20260709-143022` before deploying fresh files.

---

## Step 6 — Clean up

`pablo uninstall` removes the deployed directory and PATH entries (for binary type). For this static example, delete `./dist` manually or use uninstall:

```bash
pablo uninstall -p local-test -e dev -f pablo.yaml
```

---

## Try from the repo fixture

```bash
cd tests/agnostic/local-deploy
pablo check -f pablo.yaml -p local-test -e dev
pablo run -p local-test -e dev -f pablo.yaml
```

---

## Next steps

| Topic | Guide |
|-------|-------|
| Remote server deploy | [SSH](../guides/ssh.md) |
| Compiled binary + PATH | [Binary and PATH](../guides/binary-and-path.md) |
| Docker Compose | [Docker](../guides/docker.md) |
| Deploy strategies in depth | [Deploy strategies](../guides/deploy-strategies.md) |
