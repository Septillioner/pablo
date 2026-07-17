# Binary Deployments and PATH

Build a compiled executable, filter it with `deploy.source`, copy it to `target_path`, optionally register that directory on PATH, then run any `post_commands` you need (for example a service restart).

**See also:** [Configuration — RegisterPath](../reference/configuration.md#registerpath) · [Examples #3](../examples/README.md#3-local-binary--path) · [Deploy strategies](deploy-strategies.md)

---

## How it works

1. Run `build.command` (required on the profile or environment).
2. Filter artifacts via `deploy.source`.
3. Copy the binary to `deploy.target_path`.
4. Optionally register the directory on PATH.
5. Run `post_commands`.

```yaml
name: cli
version: 0.1.0

profiles:
  default:
    type: binary
    build:
      command: go build -o mycli .
      path: .
    environments:
      production:
        deploy:
          source:
            dir: .
            include: ["mycli"]
          target_path: ./bin
          strategy: overwrite
        register_path:
          scope: user
```

Binary profiles forbid `git` and `deploy.docker`.

---

## Artifact selection

Specify which files to deploy on each environment — sources do not inherit from the profile:

```yaml
deploy:
  source:
    dir: ./build
    include: ["myapp", "myapp.exe"]
    exclude: ["*.tmp"]
  target_path: /opt/myapp/bin
```

---

## PATH registration

After deploy, Pablo can add `target_path` to PATH so the binary is invocable by name. `register_path` applies to binary profiles only.

```yaml
register_path:
  scope: user
```

| Platform | User scope | System scope |
|----------|------------|--------------|
| Windows | User `PATH` via PowerShell | Machine `PATH` |
| macOS | Shell profile | `/etc/paths.d/` |
| Linux | `~/.bashrc` or similar | `/etc/profile.d/pablo.sh` |

`scope: system` may require elevated privileges.

---

## Remote binary deploy

Combine with SSH for remote installation. PATH registration then updates shell config on the remote host:

```yaml
name: cli
version: 0.1.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  default:
    type: binary
    build:
      command: go build -o myapp .
      path: .
    environments:
      production:
        remote:
          host: server.example.com
          credential: prod-ssh
        deploy:
          source:
            dir: .
            include: ["myapp"]
          target_path: /usr/local/bin
          strategy: overwrite
          post_commands:
            - systemctl restart myapp
        register_path:
          scope: system
```

Fixture-shaped sample: [tests/agnostic/separate-apps/go-binary](../../tests/agnostic/separate-apps/go-binary/).

---

## Uninstall

`pablo uninstall` removes deployed files and cleans PATH entries for local targets:

```bash
pablo uninstall -p default -e production
pablo uninstall -p default -e production --remove-backups
```

Windows PATH cleanup is supported. Remote uninstall is not automated — remove files on the server manually.

---

## Service restart

Use `post_commands` for systemd, NSSM, or any process manager. Windows NSSM sample: [Examples #15](../examples/README.md#15-windows-binary--service).

```yaml
deploy:
  post_commands:
    - systemctl daemon-reload
    - systemctl restart myapp
```

---

## Self-deploy

Pablo can deploy itself using the repo-root [pablo.yaml](../../pablo.yaml). Install the CLI first ([installation](../getting-started/installation.md)), then:

```bash
pablo run -f pablo.yaml -e linux-local
```

Also try `windows-local` or `macos-local`. Fixture: [tests/agnostic/self-deploy](../../tests/agnostic/self-deploy/).
