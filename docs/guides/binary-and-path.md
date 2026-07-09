# Binary Deployments and PATH

Build, deploy, and register compiled executables so they are callable from the shell.

**See also:** [Configuration — RegisterPath](../reference/configuration.md#registerpath) · [Deploy strategies](deploy-strategies.md)

---

## Overview

The `binary` profile type:

1. Runs `build.command`
2. Filters artifacts via `output_dir` or `deploy.source`
3. Copies the binary to `deploy.target_path`
4. Registers the directory in the system PATH (optional)
5. Runs `post_commands` (e.g. systemd restart)

```yaml
profiles:
  cli:
    type: binary
    build:
      command: go build -o pablo .
      path: ./src
    environments:
      production:
        deploy:
          source:
            dir: ./src
            include: ["pablo"]
          target_path: /usr/local/bin
          strategy: overwrite
        register_path:
          scope: system
```

---

## Artifact selection

Specify which files to deploy:

```yaml
# Profile-level (inherited)
output_dir:
  dir: ./build
  include: ["myapp", "myapp.exe"]
  exclude: ["*.tmp"]

# Or per-environment override
deploy:
  source:
    dir: ./build
    include: ["myapp"]
  target_path: /opt/myapp/bin
```

---

## PATH registration

After deploy, Pablo can add `target_path` to the PATH so the binary is invocable by name.

```yaml
register_path:
  scope: user     # default — current user's PATH
  # scope: system  # requires elevated privileges on some platforms
```

| Platform | User scope | System scope |
|----------|------------|--------------|
| Windows | User `PATH` via PowerShell | Machine `PATH` |
| macOS | Shell profile | `/etc/paths.d/` |
| Linux | `~/.bashrc` or similar | `/etc/profile.d/pablo.sh` |

`register_path` applies to **binary** profiles only.

---

## Remote binary deploy

Combine with SSH for remote installation:

```yaml
environments:
  production:
    remote:
      method: ssh
      host: server.example.com
      credential: prod-ssh
    deploy:
      target_path: /usr/local/bin
      strategy: overwrite
    register_path:
      scope: system
```

Remote PATH registration updates shell config on the **remote** host.

---

## Uninstall

`pablo uninstall` removes deployed files and cleans PATH entries (local targets):

```bash
pablo uninstall -p cli -e production
pablo uninstall -p cli -e production --remove-backups
```

Windows PATH cleanup is supported. Remote uninstall is not automated — remove files on the server manually.

---

## Service restart

`deploy.service` (systemd/PM2) is **not implemented** at runtime. Use `post_commands`:

```yaml
deploy:
  post_commands:
    - systemctl daemon-reload
    - systemctl restart myapp
```

---

## Self-deploy example

Pablo deploys itself via [pablo.yaml](../../pablo.yaml):

```bash
./publish-self.sh    # macOS/Linux
./publish-self.ps1   # Windows (elevated)
```

Fixture: [tests/agnostic/self-deploy](../../tests/agnostic/self-deploy/).
