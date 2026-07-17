# SSH Remote Deploy

When an environment includes a `remote` block, Pablo connects over SSH and runs the deploy on that host. Omit `remote` for a local deploy. All four profile types work remotely: `static`, `binary`, `docker`, and `git-sync`.

**See also:** [Credentials](credentials.md) · [Examples #4](../examples/README.md#4-ssh-static) · [SECURITY.md](../../SECURITY.md)

---

## Minimal remote static deploy

Define a named SSH credential, attach it to `remote`, and set an absolute `target_path` on the server:

```yaml
name: site
version: 0.1.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  web:
    type: static
    environments:
      production:
        remote:
          host: web.example.com
          credential: prod-ssh
        deploy:
          source:
            dir: ./src
            include: ["**/*"]
          target_path: /var/www/html
          strategy: backup
```

Add a profile-level `build` when the site needs compile or bundle before transfer. Full samples: [Examples #4](../examples/README.md#4-ssh-static) and [#5](../examples/README.md#5-ssh-rename-replace).

---

## Credentials

Define SSH credentials once under `credentials` and reference them from `remote.credential`.

| Auth method | Fields |
|-------------|--------|
| Key-based | `type: ssh`, `username`, `key`, optional `passphrase` |
| Password | `type: ssh`, `username`, `password` |

Paths like `~/.ssh/id_rsa` expand on the machine running Pablo (your laptop or CI runner), not on the remote host. Guide: [Credentials](credentials.md).

---

## File transfer modes

| Mode | Config | Behavior |
|------|--------|----------|
| **tar** (default) | `deploy.transfer: tar` or omit | Stream a tar archive over SSH — fast for many files |
| **legacy** | `deploy.transfer: legacy` | SCP files one by one — slower, useful for debugging |

```yaml
deploy:
  source:
    dir: ./dist
    include: ["**/*"]
  target_path: /var/www/html
  transfer: tar
  verify_checksum: true
```

Tar transfer streams over SSH stdin (no intermediate archive on disk). With `verify_checksum: true`, Pablo hashes local artifacts and runs `sha256sum -c` on the remote over stdin. Full sample: [Examples #14](../examples/README.md#14-transfer--checksum).

---

## Remote paths

`deploy.target_path` must be an absolute path on the remote host (POSIX paths for Linux targets). Pablo joins remote path segments with `pathutil.JoinRemote`. If you develop on Windows and deploy to Linux, use forward slashes in remote paths.

---

## Remote Docker

Docker profiles clone or pull the repo on the remote host and run `docker compose` there:

```yaml
name: stack
version: 0.1.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  app:
    type: docker
    git:
      repo: https://github.com/user/app.git
      branch: main
    environments:
      production:
        remote:
          host: docker-host.example.com
          credential: prod-ssh
        deploy:
          target_path: /opt/app
          docker:
            compose_file: docker-compose.yml
            build: true
```

Full sample: [Examples #7](../examples/README.md#7-docker-compose-over-ssh). Guide: [Docker](docker.md).

---

## Remote commands

`deploy.pre_commands` and `deploy.post_commands` run on the remote host when `remote` is set. Use them for restarts, migrations, or any shell step that must execute after files land.

---

## Host key verification

SSH host key verification is enabled by default. Pablo checks the remote key against the OpenSSH `known_hosts` file (`~/.ssh/known_hosts` on macOS/Linux, `%USERPROFILE%\.ssh\known_hosts` on Windows).

If the host is unknown, Pablo fails with the presented fingerprint and suggests adding it:

```bash
ssh-keyscan -H your.server.example >> ~/.ssh/known_hosts
```

Optional trust-on-first-use (default off):

```yaml
remote:
  host: web.example.com
  credential: prod-ssh
  trust_on_first_use: on
```

Opt out per environment (emits a warning; avoid on untrusted networks):

```yaml
remote:
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
| Slow deploy | Prefer `tar` over `legacy` |

More: [Troubleshooting](../troubleshooting.md).
