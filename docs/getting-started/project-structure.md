# Project Structure

How Pablo organizes configuration — profiles, environments, and multiple manifest files.

This page covers **manifest concepts**, not the Pablo source repository layout. For repository architecture see [Architecture](../development/architecture.md).

---

## One manifest, many apps

A single `pablo.yaml` can describe multiple application profiles:

```yaml
name: my-monorepo
version: 1.0.0

profiles:
  frontend:
    type: static
    # ...
  api:
    type: binary
    # ...
  worker:
    type: docker
    # ...
```

Each profile has its own `type`, build settings, and `environments` map.

Run a specific profile and environment:

```bash
pablo run -p frontend -e production
pablo run -p api -e staging
```

---

## Profiles and environments

```
pablo.yaml
├── credentials   (optional)
├── sequences     (optional — ordered profile/env lists)
└── profiles
    └── <profileName>
        ├── type          (static | binary | docker | git-sync)
        ├── build         (optional, inherited)
        ├── output_dir    (optional, inherited)
        ├── git           (docker / git-sync)
        ├── hooks
        ├── pipeline
        └── environments
            └── <envName>
                ├── remote      (optional — enables SSH)
                ├── build       (optional override)
                ├── variables
                ├── register_path (binary only)
                └── deploy
                    ├── target_path
                    ├── strategy
                    └── ...
```

**Profile** = what you're deploying (frontend, API, infra).  
**Environment** = where it goes (production, staging, local).

Defaults when omitted: profile `default`, environment `production`.

---

## Sequences

Optional root-level `sequences` run several `profile/env` targets in order:

```yaml
sequences:
  release:
    - api/staging
    - api/production
    - web/production
```

```bash
pablo run sequence release
```

List order is execution order; Pablo stops on the first failure. Details: [Sequences guide](../guides/sequences.md) · [Configuration — Sequences](../reference/configuration.md#sequences).

---

## Inheritance

Profile-level settings flow into environments:

| Profile field | Becomes |
|---------------|---------|
| `variables` | Environment variables |
| `build` | Environment build (unless overridden) |
| `output_dir` | `deploy.source` (unless `deploy.source` is set) |

Environment `variables` merge into `deploy.variables`.

Details: [Configuration — Inheritance](../reference/configuration.md#inheritance).

---

## Multiple manifest files

Pablo accepts any YAML file via `-f`:

```bash
pablo run -f pablo-sepy.yaml -p cli-release -e production
```

Convention: `pablo*.yaml` or `pablo*.yml` (matched by the VS Code extension).

Common patterns:

| File | Use case |
|------|----------|
| `pablo.yaml` | Primary application deploy |
| `pablo-sepy.yaml` | Release / packaging orchestration |
| `pablo_local.yaml` | Developer overrides (gitignored) |

---

## Credentials block

Shared credentials live at the root and are referenced by name:

```yaml
credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_deploy

profiles:
  api:
    environments:
      production:
        remote:
          credential: prod-ssh
```

See [Credentials guide](../guides/credentials.md).

---

## Legacy single-profile format

Older manifests with a top-level `type` and `environments` (no `profiles` key) are auto-wrapped into `profiles.default`. Prefer the explicit `profiles` structure for new projects.

---

## Example layouts

### Monorepo (one manifest)

```
my-app/
├── pablo.yaml          # frontend + api profiles
├── frontend/
├── backend/
└── infra/
```

See [tests/agnostic/multi-profile](../../tests/agnostic/multi-profile/).

### Separate apps (multiple manifests)

```
my-app/
├── frontend/pablo.yaml
├── backend/pablo.yaml
└── php-app/pablo.yaml
```

See [tests/agnostic/separate-apps](../../tests/agnostic/separate-apps/).

---

## Related

- [Examples (easy → hard)](../examples/README.md)
- [Configuration reference](../reference/configuration.md)
- [CLI defaults](../reference/cli.md)
- [Capabilities](../reference/capabilities.md)
