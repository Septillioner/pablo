# Deploy Strategies

`deploy.strategy` controls how Pablo treats an existing deployment directory before writing new artifacts.

**See also:** [Configuration — Deploy](../reference/configuration.md#deploy) · [Examples #5](../examples/README.md#5-ssh-rename-replace) · [Capabilities](../reference/capabilities.md)

---

## Strategies at a glance

| Strategy | Behavior | Use when |
|----------|----------|----------|
| `overwrite` | Copy into the existing directory (default) | Idempotent updates, partial overwrites |
| `backup` | Rename the existing dir with a timestamp, then deploy fresh | You want a local rollback copy on disk |
| `recreate` | Delete the target dir, recreate it, deploy | Clean slate with no leftover files |
| `rename-replace` | Per-file rename of existing artifacts, replace, cleanup on success | Binary or selective file updates without touching the whole tree |

---

## Overwrite (default)

Files land in `target_path`. Matching names are replaced; other files remain.

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

Best for local dev drops and incremental updates where leftover files are harmless. See [Examples #1](../examples/README.md#1-local-static-no-build).

---

## Backup

Before deploying, Pablo renames the existing `target_path` to a timestamped sibling (for example `target_path.YYYYMMDD-HHMMSS`), then writes a fresh tree.

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
```

Remove old backups with:

```bash
pablo uninstall -p default -e production --remove-backups
```

Backups are not encrypted. See [SECURITY.md](../../SECURITY.md). Sample: [Examples #4](../examples/README.md#4-ssh-static).

---

## Recreate

Pablo deletes `target_path` entirely, creates an empty directory, then deploys. Use this when stale files must not remain. It is destructive — everything in the target directory is removed.

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
            dir: ./dist
            include: ["**/*"]
          target_path: ./deploy-output
          strategy: recreate
```

---

## Rename-replace

For each deployed artifact, if a file with the same name already exists, Pablo renames it to `{filename}.{YYYYMMDD_HHMMSS_mmm}`, copies the new file to the original name, then deletes the renamed copies when the deploy succeeds. On failure it rolls back: newly written files are removed and renamed originals are restored.

Unlike `backup`, sibling files that are not in the artifact set stay untouched. Best for binary or selective file updates.

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

Samples: [Examples #5](../examples/README.md#5-ssh-rename-replace) (SSH) · [Examples #16](../examples/README.md#16-windows-rename-replace) (Windows local).

---

## Protected paths

Pablo blocks `backup` and `recreate` against known system directories (for example `/`, `/usr`, `C:\Windows`) unless you pass `--force`:

```bash
pablo run -e production --force
```

Detection is shallow — only top-level system paths are checked. Treat `--force` as a last resort.

---

## Rollback

For `overwrite`, Pablo does not auto-rollback on deploy failure. For `backup`, `recreate`, and `rename-replace`, the deployer rolls back when the copy or transfer step fails (`rename-replace` restores renamed files and removes partially written artifacts).

Manual rollback after a successful `backup` deploy: remove the new `target_path`, then rename the timestamped backup directory back to `target_path`.

---

## Choosing a strategy

| Scenario | Recommended |
|----------|-------------|
| Local dev / CI artifact drop | `overwrite` |
| Production static site | `backup` |
| Tree that must not keep stale files | `recreate` |
| Single binary or per-file swap | `rename-replace` |

For zero-downtime style cutover with two directories and a user-owned switch, see [Blue-green](blue-green.md) (`deploy.blue_green`) instead of a strategy value.
