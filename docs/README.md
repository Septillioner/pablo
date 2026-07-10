# Pablo Documentation

Pablo is a CLI deployment helper that automates building, filtering, and deploying artifacts across local and remote (SSH) environments from a single YAML manifest.

**Version:** see `pablo version` · **Manifest:** `pablo.yaml` · **Defaults:** profile `default`, environment `production`

---

## Getting started

| Page | Description |
|------|-------------|
| [Installation](getting-started/installation.md) | Download binary, build from source, self-deploy, VS Code extension |
| [Quick start](getting-started/quick-start.md) | Create a manifest, validate, and run your first deploy |
| [First deployment](getting-started/first-deployment.md) | Step-by-step local static deploy walkthrough |
| [Project structure](getting-started/project-structure.md) | Manifest concepts, profiles, environments, multiple files |
| [Examples](examples/README.md) | Progressive samples — local copy → build → SSH → Docker → multi-profile |

---

## Guides

| Page | Description |
|------|-------------|
| [SSH remote deploy](guides/ssh.md) | Credentials, tar vs legacy transfer, remote Docker |
| [Docker](guides/docker.md) | Local and remote Docker Compose deployments |
| [Credentials](guides/credentials.md) | SSH, token, and basic auth references |
| [Deploy strategies](guides/deploy-strategies.md) | Overwrite, backup, recreate, protected paths |
| [Git sync](guides/git-sync.md) | Pull-based deployments for interpreted apps |
| [Binary and PATH](guides/binary-and-path.md) | Binary type, PATH registration, uninstall cleanup |
| [VS Code extension](guides/vscode.md) | Editor setup, LSP, Run command, troubleshooting |
| [Visual Studio extension](guides/visual-studio.md) | VS 2022/2026 VSIX, LSP, tool window, toolbar, troubleshooting |

---

## Reference

| Page | Description |
|------|-------------|
| [CLI](reference/cli.md) | All commands and flags |
| [Configuration](reference/configuration.md) | Complete `pablo.yaml` field reference |
| [Capabilities](reference/capabilities.md) | Deployment types, strategies, pipeline, limitations |
| [Exit codes](reference/exit-codes.md) | Process exit behavior |
| [API](reference/api.md) | `inspect --json` and LSP protocol surface |

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
