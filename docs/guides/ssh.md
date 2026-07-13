# SSH Remote Deploy

Deploy to remote Linux servers over SSH from your local machine.

**See also:** [Credentials](credentials.md) · [Configuration — Remote](../reference/configuration.md#remote) · [SECURITY.md](../../SECURITY.md)

---

## Overview

When an environment has a `remote` block, Pablo connects via SSH and runs deploy commands on the target host. Supported profile types: `static`, `binary`, `docker`, `git-sync`.

Simplest case — copy existing files (no build):

```yaml
credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_deploy

profiles:
  web:
    type: static
    output_dir:
      dir: ./src
      include: ["**/*"]
    environments:
      production:
        remote:
          method: ssh
          host: web.example.com
          credential: prod-ssh
        deploy:
          target_path: /var/www/html
          strategy: backup
```

With a build step (optional):

```yaml
profiles:
  web:
    type: static
    build:
      command: npm run build
      path: .
    output_dir:
      dir: ./dist
      include: ["**/*"]
    environments:
      production:
        remote:
          method: ssh
          host: web.example.com
          credential: prod-ssh
        deploy:
          target_path: /var/www/html
          strategy: backup
```

More progressive samples: [Examples](../examples/README.md).

---

## Credentials

Define SSH credentials once in `credentials` and reference by name in `remote.credential`.

| Auth method | Fields |
|-------------|--------|
| Key-based | `type: ssh`, `username`, `key`, optional `passphrase` |
| Password | `type: ssh`, `username`, `password` |

Paths like `~/.ssh/id_rsa` are expanded on the machine running Pablo (your laptop or CI runner), not on the remote host.

---

## File transfer modes

| Mode | Config | Behavior |
|------|--------|----------|
| **tar** (default) | `deploy.remote: tar` or omit | Stream a tar archive over SSH — fast for many files |
| **legacy** | `deploy.remote: legacy` | SCP files one by one — slower, useful for debugging |

```yaml
deploy:
  target_path: /var/www/html
  remote: tar    # default
```

---

## Remote paths

`deploy.target_path` must be an **absolute path on the remote host** (POSIX paths for Linux targets).

Pablo uses `pathutil.JoinRemote` when combining remote paths. If you develop on Windows and deploy to Linux, use forward slashes in remote paths.

---

## Remote Docker

Docker profiles can run `docker compose` on a remote host over SSH:

```yaml
profiles:
  app:
    type: docker
    git:
      repo: https://github.com/user/app.git
      branch: main
    environments:
      production:
        remote:
          method: ssh
          host: docker-host.example.com
          credential: prod-ssh
        deploy:
          target_path: /opt/app
          docker:
            compose_file: docker-compose.yml
            build: true
```

Pablo clones/pulls the repo on the remote host, writes env files, and runs `docker compose up`.

---

## Remote commands

`deploy.pre_commands` and `deploy.post_commands` run on the remote host when `remote` is set. Hooks (`hooks.pre` / `hooks.post`) run on the **local** machine (where you invoke `pablo run`).

---

## Host key verification

SSH host key verification is **enabled by default**. Pablo checks the remote key against the OpenSSH `known_hosts` file:

- macOS / Linux: `~/.ssh/known_hosts`
- Windows: `%USERPROFILE%\.ssh\known_hosts`

If the host is unknown, Pablo fails with the presented key fingerprint and suggests adding it:

```bash
ssh-keyscan -H your.server.example >> ~/.ssh/known_hosts
# or connect once: ssh user@your.server.example
```

Optional trust-on-first-use (default off):

```yaml
remote:
  method: ssh
  host: web.example.com
  credential: prod-ssh
  trust_on_first_use: on
```

Opt out per environment (emits a warning; not recommended on untrusted networks):

```yaml
remote:
  method: ssh
  host: web.example.com
  credential: prod-ssh
  host_key_verification: off
```

Full policy: [SECURITY.md](../../SECURITY.md).

---

## Troubleshooting

| Symptom | Check |
|---------|-------|
| Connection refused | Host, port (default 22), firewall on target |
| Permission denied | Username, key path, key permissions (`chmod 600`) |
| Host key not in known_hosts | Add with `ssh-keyscan` or interactive `ssh`; or set `trust_on_first_use: on` |
| Host key mismatch | Confirm the server was reinstalled; remove the stale known_hosts entry only if you trust the new key |
| Remote path errors | Use absolute POSIX paths on Linux |
| Slow deploy | Switch from `legacy` to `tar` (default) |

More: [Troubleshooting](../troubleshooting.md).
