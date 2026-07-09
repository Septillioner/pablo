# Deploy Strategies

Control how Pablo handles an existing deployment directory.

**See also:** [Configuration — Deploy](../reference/configuration.md#deploy) · [Capabilities](../reference/capabilities.md)

---

## Strategies

| Strategy | Behavior | Use when |
|----------|----------|----------|
| `overwrite` | Copy new files into existing directory (default) | Idempotent updates, partial overwrites |
| `backup` | Rename existing dir with timestamp, deploy fresh | Safe rollback to previous version on disk |
| `recreate` | Delete target dir, recreate, deploy | Clean slate needed |
| `blue-green` | *(not implemented)* | — |

Set on `deploy.strategy`:

```yaml
deploy:
  target_path: /var/www/html
  strategy: backup
```

---

## Overwrite (default)

Files are copied into `target_path`. Existing files with the same name are replaced; other files remain.

Best for: static sites, incremental binary updates where old files are harmless.

---

## Backup

Before deploying, Pablo renames the existing `target_path` to `target_path.YYYYMMDD-HHMMSS` (or similar timestamp suffix).

Best for: production deploys where you want a local rollback copy on the target.

Remove old backups with:

```bash
pablo uninstall -p myprofile -e production --remove-backups
```

Backups are **not encrypted**. See [SECURITY.md](../../SECURITY.md).

---

## Recreate

Pablo deletes `target_path` entirely, creates an empty directory, then deploys.

Best for: environments that must not retain stale files.

**Caution:** destructive — all files in the target directory are removed.

---

## Protected paths

Pablo blocks `backup` and `recreate` against known system directories (e.g. `/`, `/usr`, `C:\Windows`) unless you pass `--force`:

```bash
pablo run -e production --force
```

Protected path detection is shallow — only top-level system paths are checked. Treat `--force` as a last resort.

---

## Rollback

Pablo does not auto-rollback on deploy failure for `overwrite`. For `backup`/`recreate`, the deployer attempts rollback when the copy step fails.

Manual rollback after a successful `backup` deploy:

1. Remove the new `target_path`
2. Rename the timestamped backup directory back to `target_path`

---

## Blue-green

Declared in the schema but **not implemented**. Using `strategy: blue-green` returns a runtime error. Track progress on the [roadmap](../roadmap.md).

---

## Choosing a strategy

| Scenario | Recommended |
|----------|-------------|
| Local dev / CI artifact drop | `overwrite` |
| Production static site | `backup` |
| Container-less binary with stale libs | `recreate` |
| Zero-downtime swap | Not available yet |
