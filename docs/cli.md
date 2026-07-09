# Pablo CLI Reference

Command-line interface for running deployments, validating manifests, and managing installed artifacts.

**Defaults:** manifest = `pablo.yaml`, profile = `default`, environment = `production`

---

## Commands

| Command | Description |
|---------|-------------|
| `run` | Execute the full deployment pipeline |
| `check` | Validate a manifest file |
| `init` | Generate a sample manifest |
| `uninstall` | Remove deployed files and clean up PATH entries |
| `version` | Print Pablo version information |
| `lsp` | Start the language server (stdio; used by the VS Code extension) |

---

## `run`

Executes the deployment pipeline for the selected profile and environment.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--profile` | `-p` | `default` | Target profile |
| `--env` | `-e` | `production` | Target environment |
| `--file` | `-f` | `pablo.yaml` | Path to manifest |
| `--force` | | `false` | Allow deployment to protected system directories |

**Examples:**

```bash
pablo run -p frontend -e production
pablo run -p api -e production -f deploy.yaml
pablo run -e staging --force
```

---

## `check`

Loads and validates the manifest (semantic schema rules via `pkg/validate`, with `path:line:col` diagnostics). Optionally checks that a profile and environment exist.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | `pablo.yaml` | Path to manifest |
| `--profile` | `-p` | *(empty)* | Validate a specific profile |
| `--env` | `-e` | *(empty)* | Validate a specific environment |

**Examples:**

```bash
pablo check
pablo check -f pablo.yaml -p frontend -e production
```

---

## `init`

Creates a sample manifest file named `pablo_sample.yaml` in the current directory.

**Flags:** none

**Example:**

```bash
pablo init
```

---

## `uninstall`

Removes deployed files from the target path and cleans up PATH registrations for binary deployments.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--profile` | `-p` | `default` | Profile to uninstall |
| `--env` | `-e` | *(required)* | Environment to uninstall |
| `--file` | `-f` | `pablo.yaml` | Path to manifest |
| `--remove-backups` | | `false` | Also remove backup directories |

**Examples:**

```bash
pablo uninstall -p api -e production
pablo uninstall -p api -e production --remove-backups
```

---

## `version`

Prints the current Pablo version.

**Flags:** none

**Example:**

```bash
pablo version
```

---

## `lsp`

Starts the Pablo language server on stdio. Used by the VS Code extension (`pablo lsp`). Does not print the CLI header to stdout.

**Flags:** none

**Example:**

```bash
pablo lsp
```

