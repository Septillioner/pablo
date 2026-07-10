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
| `rename-replace` | Per-file rename of existing artifacts, replace, cleanup on success | Binary or single-file updates without touching the whole directory |
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

## Rename-replace

For each deployed artifact file, if a file with the same name already exists at `target_path`, Pablo renames it to `{filename}.{YYYYMMDD_HHMMSS_mmm}` (for example `app.exe.20260710_121530_042`), copies the new file to the original name, then deletes the renamed copies when the deploy succeeds.

On failure, Pablo performs a full rollback: newly written files are removed and renamed originals are restored.

Best for: binary or selective file updates where you want atomic per-file replacement without renaming the entire target directory.

Unlike `backup`, sibling files in `target_path` that are not part of the artifact set are left untouched.

---

## Protected paths

Pablo blocks `backup` and `recreate` against known system directories (e.g. `/`, `/usr`, `C:\Windows`) unless you pass `--force`:

```bash
pablo run -e production --force
```

Protected path detection is shallow — only top-level system paths are checked. Treat `--force` as a last resort.

---

## Rollback

Pablo does not auto-rollback on deploy failure for `overwrite`. For `backup`/`recreate`/`rename-replace`, the deployer rolls back when the copy or transfer step fails (`rename-replace` restores renamed files and removes partially written artifacts).

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
| Single binary or per-file swap | `rename-replace` |
| Zero-downtime swap | Not available yet |
