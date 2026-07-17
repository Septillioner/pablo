# Sequences

Run several `profile/env` targets in a fixed order from one command. A normal `pablo run` deploys one profile and one environment; sequences chain targets when order matters — for example promote staging then production, or package then publish.

**See also:** [Configuration — Sequences](../reference/configuration.md#sequences) · [CLI — run](../reference/cli.md#run) · [Examples #11](../examples/README.md#11-sequences)

---

## Define a sequence

Add a root-level `sequences` map. Each value is an ordered list of `profile/env` steps (cross-profile steps are allowed). List order is execution order — Pablo does not sort alphabetically.

```yaml
name: release-bundle
version: 1.0.0

sequences:
  ship:
    - web/production
    - api/production

profiles:
  web:
    type: static
    environments:
      production:
        deploy:
          source:
            dir: ./web/dist
            include: ["**/*"]
          target_path: ./out/web
          strategy: overwrite

  api:
    type: static
    environments:
      production:
        deploy:
          source:
            dir: ./api/dist
            include: ["**/*"]
          target_path: ./out/api
          strategy: overwrite
```

---

## Run a sequence

```bash
pablo run sequence ship
pablo run sequence ship -f pablo.yaml --verbose
```

| Flag | Applies to sequences? |
|------|------------------------|
| `-f` / `--file` | Yes — all steps use that manifest |
| `--force` | Yes — every step |
| `--verbose` | Yes — every step |
| `-p` / `--profile` | No — cannot combine with `sequence` |
| `-e` / `--env` | No — cannot combine with `sequence` |

`pablo run sequence/foo` (one argument containing `/`) is still a normal profile/env target whose profile name is `sequence`. Use two arguments: `pablo run sequence <name>`.

---

## Behavior

1. Load and validate the manifest (including every sequence step).
2. Look up `sequences.<name>`.
3. For each step in list order, run the full single-target pipeline (build → pre/post commands → deploy → PATH registration).
4. On the first failure, abort — later steps do not run.

There is no `--continue-on-error` in the current release.

---

## Validation and discovery

```bash
pablo check -f pablo.yaml
pablo inspect -f pablo.yaml
pablo inspect -f pablo.yaml --json
```

`check` rejects empty sequences, invalid `profile/env` strings, and missing profiles or environments. `inspect` lists sequence names and their steps with order preserved.

---

## Limitations

- No nested sequences (a step cannot be another sequence name).
- Steps must be full `profile/env` — bare environment names are not allowed.
- Editor CodeLens / Run pickers still target a single profile/env; use the CLI for sequences.

---

## Related

| Topic | Page |
|-------|------|
| Field reference | [Configuration — Sequences](../reference/configuration.md#sequences) |
| Manifest layout | [Project structure](../getting-started/project-structure.md#sequences) |
| Copy-paste sample | [Examples #11](../examples/README.md#11-sequences) |
| Pipeline overview | [Capabilities](../reference/capabilities.md#sequences) |
