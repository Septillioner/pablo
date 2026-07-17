# Project Structure

How a Pablo manifest organizes work: what you deploy (profile), where it runs (environment), and how artifacts land (`deploy`). This page is about manifest concepts, not the Pablo source tree — for that see [Architecture](../development/architecture.md).

---

## Three nouns

| Noun | YAML location | Meaning |
|------|---------------|---------|
| **Profile** | `profiles.<name>` | What you deploy — type, build, git, shared variables |
| **Environment** | `profiles.<name>.environments.<name>` | Where it runs — local or SSH (`remote`), plus deploy settings |
| **Deploy** | `environments.<name>.deploy` | How artifacts land — source, strategy, commands, docker |

Everything deployable lives under `profiles`. Root-level fields are project metadata (`name`, `version`), shared `credentials`, and optional `sequences`.

Defaults when omitted: profile `default`, environment `production`.

---

## Shape of a manifest

```
pablo.yaml
├── credentials   (optional)
├── sequences     (optional — ordered profile/env lists)
└── profiles
    └── <profileName>
        ├── type          (static | binary | docker | git-sync)
        ├── variables     (inherited)
        ├── env_file      (inherited)
        ├── build         (optional, inherited)
        ├── git           (docker / git-sync)
        └── environments
            └── <envName>
                ├── remote      (optional — SSH when present)
                ├── build       (optional override)
                ├── variables
                ├── env_file
                ├── register_path (binary only)
                └── deploy
                    ├── source        (static / binary)
                    ├── target_path
                    ├── strategy
                    ├── transfer      (remote SSH)
                    ├── verify_checksum
                    ├── pre_commands
                    ├── post_commands
                    └── docker        (docker type)
```

---

## One manifest, many apps

A single file can hold several profiles. Each keeps its own `type`, build settings, and environments map:

```yaml
name: my-monorepo
version: 1.0.0

profiles:
  frontend:
    type: static
    environments:
      production:
        deploy:
          source:
            dir: ./frontend/dist
            include: ["**/*"]
          target_path: ./out/frontend
          strategy: overwrite

  api:
    type: binary
    build:
      command: go build -o api .
      path: ./api
    environments:
      production:
        deploy:
          source:
            dir: ./api
            include: ["api"]
          target_path: ./out/api
          strategy: overwrite
```

```bash
pablo run -p frontend -e production
pablo run -p api -e staging
```

Full sample: [Examples #9](../examples/README.md#9-multi-profile-monorepo).

---

## Sequences

Optional root-level `sequences` run several `profile/env` targets in list order and stop on the first failure:

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

Details: [Sequences guide](../guides/sequences.md) · [Examples #11](../examples/README.md#11-sequences).

---

## Inheritance

Only these profile fields cascade into each environment:

| Profile field | Behavior |
|---------------|----------|
| `variables` | Merged into environment variables (env wins on conflict) |
| `env_file` | Default env file name when the environment omits `env_file` |
| `build` | Copied when the environment has no `build`; partial merge when it does |

`deploy.source`, `deploy.target_path`, and `remote` are always set on the environment — they do not inherit. See [Examples #12](../examples/README.md#12-inheritance) and [Configuration — Inheritance](../reference/configuration.md#inheritance).

---

## Multiple manifest files

Point Pablo at any YAML file with `-f`:

```bash
pablo run -f pablo-sepy.yaml -p cli-release -e production
```

Editors match `pablo*.yaml` / `pablo*.yml`. Common patterns:

| File | Use case |
|------|----------|
| `pablo.yaml` | Primary application deploy |
| `pablo-sepy.yaml` | Release / packaging orchestration |
| `pablo_local.yaml` | Developer overrides (often gitignored) |

When apps ship on different cadences, prefer one small manifest per app — [Examples #10](../examples/README.md#10-separate-apps).

---

## Credentials

Shared credentials live at the root and are referenced by name from `remote.credential` or `git.credential`:

```yaml
credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  api:
    type: static
    environments:
      production:
        remote:
          host: api.example.com
          credential: prod-ssh
        deploy:
          source:
            dir: ./dist
            include: ["**/*"]
          target_path: /var/www/api
          strategy: backup
```

Guide: [Credentials](../guides/credentials.md) · [Examples #13](../examples/README.md#13-credentials).

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

Fixture: [tests/agnostic/multi-profile](../../tests/agnostic/multi-profile/).

### Separate apps (multiple manifests)

```
my-app/
├── frontend/pablo.yaml
├── backend/pablo.yaml
└── php-app/pablo.yaml
```

Fixture: [tests/agnostic/separate-apps](../../tests/agnostic/separate-apps/).

---

## Related

- [Examples](../examples/README.md)
- [Configuration reference](../reference/configuration.md)
- [CLI](../reference/cli.md)
- [Capabilities](../reference/capabilities.md)
