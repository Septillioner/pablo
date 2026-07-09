# Docker Deployments

Run Docker Compose workloads locally or on a remote host over SSH.

**See also:** [Configuration — Docker](../reference/configuration.md#docker) · [SSH guide](ssh.md)

---

## Overview

The `docker` profile type:

1. Clones or pulls a Git repository
2. Writes environment variables to an `.env` file (if configured)
3. Runs `docker compose` with your compose file

```yaml
profiles:
  stack:
    type: docker
    git:
      repo: https://github.com/user/my-stack.git
      branch: main
      credential: github    # optional, for private repos
    environments:
      local:
        deploy:
          target_path: /tmp/my-stack    # clone directory
          docker:
            compose_file: docker-compose.yml
            build: true
            command: up -d --build
```

---

## Requirements

- Docker CLI and Docker Compose v2 on the target machine
- Git (for clone/pull)
- For remote deploy: SSH access (see [SSH guide](ssh.md))

---

## Local vs remote

| | Local | Remote SSH |
|---|-------|------------|
| `remote` block | Omitted | Required |
| Clone location | `deploy.target_path` on your machine | `deploy.target_path` on remote host |
| Compose runs | Locally | Via SSH on remote |

---

## Compose configuration

| Field | Default | Description |
|-------|---------|-------------|
| `compose_file` | *(required)* | Path relative to repo root after clone |
| `build` | `false` | Pass `--build` to compose |
| `command` | `up -d` | Arguments after `docker compose` |

Example with build:

```yaml
docker:
  compose_file: docker-compose.yml
  build: true
  command: up -d --build
```

---

## Environment variables

Set variables at the environment or deploy level. Pablo generates an env file in the clone directory when `env_file` is configured:

```yaml
environments:
  production:
    variables:
      APP_PORT: "8080"
      DB_URL: "postgres://..."
    deploy:
      env_file: .env
      target_path: /opt/my-stack
      docker:
        compose_file: docker-compose.yml
```

Template substitution (`{{VAR}}`) applies to config files in the deployed tree.

---

## Private repositories

Reference a token credential for HTTPS Git URLs:

```yaml
credentials:
  github:
    type: token
    value: ghp_xxxxxxxx

profiles:
  stack:
    type: docker
    git:
      repo: https://github.com/org/private-repo.git
      credential: github
```

Prefer environment-injected tokens in CI rather than committing secrets.

---

## Post-deploy commands

Use `deploy.post_commands` for tasks Compose does not handle (migrations, cache warm-up):

```yaml
deploy:
  post_commands:
    - docker compose exec -T api ./migrate up
```

On remote hosts, these run over SSH after compose up.

---

## Limitations

- Pablo wraps `docker compose`; it does not manage individual container lifecycle beyond compose.
- `deploy.service` (systemd/PM2) is not implemented — use compose `restart` policies or `post_commands`.
- Remote host must have Docker daemon running and your SSH user must have permission to run Docker.

---

## Example fixture

Repository E2E scenario: [tests/e2e/scenarios/](../../tests/e2e/).
