# FAQ

Common questions about Pablo.

---

## General

### What is Pablo?

A CLI that automates optional build → filter → deploy from a YAML manifest (`pablo.yaml`). It supports local and remote (SSH) targets, multiple profiles, and four deployment types: `static`, `binary`, `docker`, and `git-sync`. For `static`, you can omit `build` and only copy files.

### Does Pablo replace CI/CD?

No. Pablo is a deployment helper you run locally or in CI to push artifacts to servers. It does not manage PR checks or provision infrastructure.

### What license is Pablo under?

The repository [LICENSE](../LICENSE) is Apache 2.0. The VS Code extension `package.json` also lists Apache-2.0.

---

## Configuration

### Can I use multiple manifest files?

Yes. Pass `-f`:

```bash
pablo run -f pablo-sepy.yaml -p cli-release -e production
```

Editor extensions recognize `pablo*.yaml` and `pablo*.yml`.

### What are the defaults?

| Setting | Default |
|---------|---------|
| Manifest | `pablo.yaml` |
| Profile | `default` |
| Environment | `production` |

### Where do I set artifact paths?

On each environment under `deploy.source` (`dir`, `include`, `exclude`). Static and binary types require an explicit `deploy.source` per environment.

### Can I deploy static files without a build step?

Yes. For `type: static`, omit `build`. Pablo copies filtered files from `deploy.source` to `target_path`. See [Examples #1](examples/README.md#1-local-static-no-build) and [First deployment](getting-started/first-deployment.md).

### Can I run multiple environments in order?

Yes. Define a root-level `sequences` map of `profile/env` targets, then:

```bash
pablo run sequence release
```

Steps run in list order; the first failure aborts the rest. See [Sequences](guides/sequences.md) · [Examples #11](examples/README.md#11-sequences).

### Why is my `.env` / `env_file` empty or missing?

Pablo **writes** dotenv files from YAML maps; it does not load an existing `.env` as input. Put values under environment `variables`. Use `build.env_file` for a pre-build file under `build.path`, and deploy `env_file` for a post-deploy file under `target_path`. Optional `build.variables` only overlay build-time keys. An empty resolved map skips writing. See [Variables and env files](reference/configuration.md#variables-and-env-files).

---

## Deployment

### Does Pablo manage nginx, firewall, or Linux users?

No. Pablo deploys files, binaries, git repos, and Docker Compose stacks. Reverse proxies, firewall rules, and OS account provisioning are out of scope — use `deploy.post_commands` or external tooling.

### Is systemd / PM2 restart supported?

Use `deploy.post_commands`:

```yaml
post_commands:
  - systemctl restart myapp
```

### Can I deploy from Windows to Linux?

Yes. Use POSIX absolute paths in `deploy.target_path` for Linux targets. Pablo joins remote paths via `pathutil`.

---

## SSH

### Is SSH host key verification enabled?

Yes, by default. Pablo checks the remote host key against OpenSSH `known_hosts`. Unknown hosts fail with a fingerprint and add instructions. Optional `remote.trust_on_first_use: on` records the key on first connect. Opt out with `remote.host_key_verification: off` (not recommended). See [SECURITY.md](../SECURITY.md) and the [SSH guide](guides/ssh.md).

### tar vs legacy transfer?

`tar` (default) streams an archive — faster for many files. `legacy` uses SCP file-by-file — set `deploy.transfer: legacy`. Useful for debugging transfer issues. See [Examples #14](examples/README.md#14-transfer--checksum).

---

## PATH and binaries

### Multiple `pablo` binaries on PATH?

The VS Code extension prompts you to choose. Use **Pablo: Select Executable** or set `pablo.path`.

### Does uninstall work on remote servers?

`pablo uninstall` targets local deploy paths only. Remove remote deployments manually or via SSH commands.

---

## Editor

### Autocomplete works but validation doesn't (or vice versa)?

Snippets (`pablo-tpl-*`) come from the extension. Schema validation and completion require `pablo lsp` (CLI 1.3+). Check **Pablo Language Server** output.

### Does the extension bundle Pablo?

It can use a bundled binary or resolve from `pablo.path` / PATH. See [VS Code guide](guides/vscode.md).

---

## More help

- [Troubleshooting](troubleshooting.md)
- [Examples](examples/README.md)
- [Roadmap](roadmap.md)
- [GitHub Issues](https://github.com/septillioner/pablo/issues)
