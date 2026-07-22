# Git Sync

Deploy interpreted applications by pulling source into `deploy.target_path` and running post-deploy commands. There is no Pablo build step — install dependencies and restart processes with `post_commands`.

**See also:** [Configuration](../reference/configuration.md) · [Docker](docker.md) · [Examples #8](../examples/README.md#8-git-sync--post-commands)

---

## How it works

1. Clone or pull a Git repository to `deploy.target_path`.
2. Write a dotenv file from environment `variables` when `env_file` is set and the map is non-empty (Pablo does not load an existing `.env` as input).
3. Run `deploy.post_commands` on the target (local shell or remote over SSH).

```yaml
name: api
version: 0.1.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519
  github:
    type: token
    value: "${GITHUB_TOKEN}"

profiles:
  default:
    type: git-sync
    git:
      repo: https://github.com/user/my-api.git
      branch: main
      credential: github
    environments:
      production:
        remote:
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

Git-sync forbids `deploy.source`, `deploy.docker`, and `register_path`.

---

## Local git-sync

Omit `remote` to run on your machine:

```yaml
name: api
version: 0.1.0

profiles:
  default:
    type: git-sync
    git:
      repo: https://github.com/user/my-api.git
      branch: main
    environments:
      local:
        deploy:
          target_path: ./runtime/my-api
          strategy: overwrite
          post_commands:
            - npm install
            - pm2 restart my-api
```

---

## Remote git-sync

With `remote` set, clone/pull and `post_commands` execute on the remote host over SSH. See [SSH](ssh.md).

---

## Git configuration

| Field | Required | Description |
|-------|----------|-------------|
| `git.repo` | Yes | Repository URL (HTTPS or SSH) |
| `git.branch` | No | Branch (default: `main`) |
| `git.credential` | No | Token credential for private HTTPS repos |

For SSH Git URLs on the remote host, ensure the remote user has deploy keys configured. Pablo’s SSH credential authenticates Pablo’s connection to the host, not necessarily Git on that host.

---

## When to use git-sync vs docker

| | git-sync | docker |
|---|----------|--------|
| Runtime | Native on host | Containers |
| Dependencies | `post_commands` | Compose file |
| Isolation | Low | High |
| Best for | PHP, Node without containers | Microservices, reproducible stacks |

---

## Service restart

Put restart commands in `post_commands`:

```yaml
post_commands:
  - systemctl daemon-reload
  - systemctl restart my-api
```

---

## Example fixtures

- [Examples #8](../examples/README.md#8-git-sync--post-commands)
- [tests/agnostic/separate-apps/php-app](../../tests/agnostic/separate-apps/php-app/)
- [tests/agnostic/multi-profile](../../tests/agnostic/multi-profile/) (git-sync profile)
