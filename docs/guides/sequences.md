# Sequences

Run multiple `profile/env` targets in a fixed order from one command.

**See also:** [Configuration — Sequences](../reference/configuration.md#sequences) · [CLI — run](../reference/cli.md#run) · [Examples #7](../examples/README.md#7-run-targets-in-order-sequences)

---

## Why use sequences

A single `pablo run` deploys one profile and one environment. Sequences chain several targets when order matters — for example package a VSIX, then publish to a marketplace, or promote staging then production.

---

## Define a sequence

Add a root-level `sequences` map. Each value is an ordered list of `profile/env` steps (cross-profile allowed).

```yaml
name: my-app
version: 1.0.0

sequences:
  extension:
    - extension/vsix
    - extension/marketplace

profiles:
  extension:
    type: static
    # ...
    environments:
      vsix:
        deploy:
          target_path: ./dist/vsix
          strategy: overwrite
      marketplace:
        deploy:
          target_path: ./dist/marketplace
          strategy: overwrite
```

**List order is execution order.** Pablo does not sort steps alphabetically.

---

## Run a sequence

```bash
pablo run sequence extension
pablo run sequence extension -f pablo-sepy.yaml --verbose
```

| Flag | Applies to sequences? |
|------|------------------------|
| `-f` / `--file` | Yes — all steps use that manifest |
| `--force` | Yes — every step |
| `--verbose` | Yes — every step |
| `-p` / `--profile` | No — cannot combine with `sequence` |
| `-e` / `--env` | No — cannot combine with `sequence` |

`pablo run sequence/foo` (one argument with `/`) is still a normal profile/env target whose profile name is `sequence`. Use two arguments: `pablo run sequence <name>`.

---

## Behavior

1. Load and validate the manifest (including every sequence step).
2. Look up `sequences.<name>`.
3. For each step in list order, run the full single-target pipeline (`build` → deploy → hooks → health check).
4. On the first failure, abort — later steps do not run.

There is no `--continue-on-error` in the current release.

---

## Validation and discovery

```bash
pablo check -f pablo.yaml
pablo inspect -f pablo.yaml
pablo inspect -f pablo.yaml --json
```

`check` rejects empty sequences, invalid `profile/env` strings, and missing profiles or environments. `inspect` lists sequence names and their steps (order preserved).

---

## Limitations (v1)

- No nested sequences (a step cannot be another sequence name).
- Steps must be full `profile/env` — bare environment names are not allowed.
- Editor CodeLens / Run pickers still target a single profile/env (use the CLI for sequences).

---

## Related

| Topic | Page |
|-------|------|
| Field reference | [Configuration — Sequences](../reference/configuration.md#sequences) |
| Manifest layout | [Project structure](../getting-started/project-structure.md#sequences) |
| Copy-paste sample | [Examples #7](../examples/README.md#7-run-targets-in-order-sequences) |
| Pipeline overview | [Capabilities](../reference/capabilities.md#sequences) |
