# Git Sync

Deploy interpreted applications by pulling source and running post-deploy commands.

**See also:** [Configuration](../reference/configuration.md) · [Docker guide](docker.md) (also uses git)

---

## Overview

The `git-sync` type:

1. Clones or pulls a Git repository to `deploy.target_path`
2. Writes environment variables (optional `.env` file)
3. Runs `deploy.post_commands` on the target

No build step — use `post_commands` to install dependencies and restart processes.

```yaml
profiles:
  api:
    type: git-sync
    git:
      repo: https://github.com/user/my-api.git
      branch: main
      credential: github
    environments:
      production:
        remote:
          method: ssh
          host: api.example.com
          credential: prod-ssh
        deploy:
          target_path: /opt/my-api
          strategy: backup
          post_commands:
            - composer install --no-dev
            - php artisan migrate --force
            - systemctl restart my-api
        variables:
          APP_ENV: production
        env_file: .env
```

---

## Local git-sync

Omit `remote` to run on your machine:

```yaml
environments:
  local:
    deploy:
      target_path: ./runtime/my-api
      post_commands:
        - npm install
        - pm2 restart my-api
```

---

## Remote git-sync

With `remote` set, clone/pull and `post_commands` execute on the remote host over SSH.

---

## Git configuration

| Field | Required | Description |
|-------|----------|-------------|
| `git.repo` | Yes | Repository URL (HTTPS or SSH) |
| `git.branch` | No | Branch (default: `main`) |
| `git.credential` | No | Token credential for private HTTPS repos |

For SSH Git URLs on the remote host, ensure the remote user has deploy keys configured — Pablo's SSH credential is for Pablo's connection to the host, not necessarily for Git on that host.

---

## When to use git-sync vs docker

| | git-sync | docker |
|---|----------|--------|
| Runtime | Native on host | Containers |
| Dependencies | `post_commands` | Compose file |
| Isolation | Low | High |
| Best for | PHP, Node without containers, legacy apps | Microservices, reproducible stacks |

---

## Service management

`deploy.service` (systemd/PM2) is validated in the schema but **not executed** at runtime. Use explicit commands in `post_commands`:

```yaml
post_commands:
  - systemctl daemon-reload
  - systemctl restart my-api
```

Service management is planned — see [roadmap](../roadmap.md).

---

## Example fixtures

- [tests/agnostic/separate-apps/php-app](../../tests/agnostic/separate-apps/php-app/)
- [tests/agnostic/multi-profile](../../tests/agnostic/multi-profile/) (git-sync profile)
