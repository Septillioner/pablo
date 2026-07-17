# Pablo CLI Reference

Command-line interface for running deployments, validating manifests, and managing installed artifacts.

**Defaults:** manifest = `pablo.yaml`, profile = `default`, environment = `production`

See also: [Exit codes](exit-codes.md) · [Configuration](configuration.md)

---

## Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--verbose` | `false` | During `run`, list each artifact path after the deploy count |

---

## Commands

| Command | Description |
|---------|-------------|
| `run` | Execute the full deployment pipeline |
| `check` | Validate a manifest file |
| `inspect` | List profiles, environments, and sequences from a manifest |
| `init` | Generate a sample manifest (`--template` / `-t` for type wizard) |
| `uninstall` | Remove deployed files and clean up PATH entries |
| `version` | Print Pablo version information |
| `update` | Update the Pablo CLI binary from GitHub Releases |
| `lsp` | Start the language server (stdio; used by editor extensions) |

---

## `run`

Executes the deployment pipeline for a single profile/environment, or runs a named sequence from the manifest.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--profile` | `-p` | `default` | Target profile (not used with `sequence`) |
| `--env` | `-e` | `production` | Target environment (not used with `sequence`) |
| `--file` | `-f` | `pablo.yaml` | Path to manifest |
| `--force` | | `false` | Allow deployment to protected system directories |

Single target:

```bash
pablo run -p frontend -e production
pablo run -p api -e production -f deploy.yaml
pablo run -e staging --force
pablo run -p local-test -e dev --verbose
pablo run default/windows-local
```

Sequence:

```bash
pablo run sequence extension
pablo run sequence extension -f pablo-sepy.yaml --verbose
```

Runs each step in manifest list order; stops on the first failure. `-p` / `-e` cannot be combined with `sequence`. Guide: [Sequences](../guides/sequences.md).

---

## `check`

Loads and validates the manifest (semantic schema rules via `pkg/validate`, with `path:line:col` diagnostics). Optionally checks that a profile and environment exist.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | `pablo.yaml` | Path to manifest |
| `--profile` | `-p` | *(empty)* | Validate a specific profile |
| `--env` | `-e` | *(empty)* | Validate a specific environment |

```bash
pablo check
pablo check -f pablo.yaml -p frontend -e production
```

---

## `inspect`

Lists profiles, environments, and named sequences from a manifest. Editor extensions use this to populate profile/environment pickers.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | `pablo.yaml` | Path to manifest |
| `--json` | | `false` | Output machine-readable JSON (no CLI header) |

```bash
pablo inspect
pablo inspect -f pablo.yaml --json
```

JSON shape: [API](api.md#inspect-json).

---

## `init`

Creates a sample manifest file named `pablo_sample.yaml` in the current directory.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--template` | `-t` | `false` | Interactive template type wizard (`static`, `binary`, `docker`, `git-sync`) |

Without `--template`, writes a local `static` sample. With `--template`, shows a numbered menu in the terminal (requires an interactive TTY).

```bash
pablo init
pablo init --template
pablo init -t
```

---

## `uninstall`

Removes deployed files from the target path and cleans up PATH registrations for binary deployments. Local targets only.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--profile` | `-p` | `default` | Profile to uninstall |
| `--env` | `-e` | *(required)* | Environment to uninstall |
| `--file` | `-f` | `pablo.yaml` | Path to manifest |
| `--remove-backups` | | `false` | Also remove backup directories |

```bash
pablo uninstall -p api -e production
pablo uninstall -p api -e production --remove-backups
```

---

## `version`

Prints the current Pablo version and architecture label.

```bash
pablo version
```

---

## `update`

Downloads the latest Pablo CLI binary for your OS/architecture from GitHub Releases, verifies `checksums.txt`, and replaces the running executable. Editor extensions (VSIX) are not updated.

If other processes are running the same Pablo binary (for example `pablo lsp` from an editor), Pablo lists them and asks whether to close them before continuing. In a non-interactive terminal, close those processes manually if the replace step fails.

| Flag | Default | Description |
|------|---------|-------------|
| `--check` | `false` | Report whether a newer release exists; exit code `1` when an update is available |
| `--version` | *(empty)* | Pin a release tag (e.g. `v1.5.0`); also reads `PABLO_VERSION` |

```bash
pablo update
pablo update --check
PABLO_VERSION=v1.5.0 pablo update
```

After a successful update, open a new terminal (or invoke `pablo version` again) so your shell picks up the new binary.

---

## `lsp`

Starts the Pablo language server on stdio. Used by editor extensions. Does not print the CLI header to stdout.

```bash
pablo lsp
```

See [API](api.md) for LSP capabilities.
