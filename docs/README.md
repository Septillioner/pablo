# Pablo Documentation

Pablo is a CLI that builds, filters, and deploys artifacts from a single `pablo.yaml` manifest — locally or over SSH.

**Version:** see `pablo version` · **Manifest:** `pablo.yaml` · **Defaults:** profile `default`, environment `production`

---

## Getting started

| Page | Description |
|------|-------------|
| [Installation](getting-started/installation.md) | Download the binary, build from source, or install the editor extension |
| [Quick start](getting-started/quick-start.md) | Create a manifest, validate it, and run your first deploy |
| [First deployment](getting-started/first-deployment.md) | Step-by-step local static walkthrough |
| [Project structure](getting-started/project-structure.md) | Profiles, environments, sequences, and multi-file layouts |
| [Examples](examples/README.md) | Sixteen scenarios from local copy through SSH, Docker, git-sync, sequences, and Windows |

---

## Guides

| Page | Description |
|------|-------------|
| [Sequences](guides/sequences.md) | Ordered multi-target runs with `pablo run sequence` |
| [SSH remote deploy](guides/ssh.md) | Credentials, tar vs legacy transfer, remote Docker |
| [Docker](guides/docker.md) | Local and remote Docker Compose deployments |
| [Credentials](guides/credentials.md) | SSH, token, and basic auth references |
| [Deploy strategies](guides/deploy-strategies.md) | Overwrite, backup, recreate, rename-replace |
| [Blue-green](guides/blue-green.md) | Slot detect, idle-slot write, user switch command |
| [Git sync](guides/git-sync.md) | Pull-based deploys for interpreted apps |
| [Binary and PATH](guides/binary-and-path.md) | Binary type, PATH registration, uninstall |
| [VS Code extension](guides/vscode.md) | Editor setup, LSP, Run command |
| [Visual Studio extension](guides/visual-studio.md) | VS 2022/2026 VSIX, LSP, tool window |

---

## Reference

| Page | Description |
|------|-------------|
| [CLI](reference/cli.md) | Commands and flags |
| [Configuration](reference/configuration.md) | Complete `pablo.yaml` field reference |
| [Capabilities](reference/capabilities.md) | Types, strategies, pipeline phases, limitations |
| [Exit codes](reference/exit-codes.md) | Process exit behavior |
| [API](reference/api.md) | `inspect --json` and LSP surface |

---

## Development

| Page | Description |
|------|-------------|
| [Architecture](development/architecture.md) | Components, pipeline flow, services and adapters |
| [Testing](development/testing.md) | Unit, E2E, and manual fixture testing |
| [Contributing](development/contributing.md) | Contribution workflow (canonical: [CONTRIBUTING.md](../CONTRIBUTING.md)) |
| [Release process](development/release-process.md) | Versioning and publishing |

---

## More

| Page | Description |
|------|-------------|
| [Roadmap](roadmap.md) | Shipped features and planned work |
| [FAQ](faq.md) | Common questions |
| [Troubleshooting](troubleshooting.md) | Validation, SSH, Docker, and extension issues |

**Also see:** [SECURITY.md](../SECURITY.md) · [CHANGELOG.md](../CHANGELOG.md) · [LICENSE](../LICENSE)
