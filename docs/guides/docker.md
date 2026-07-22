# Docker Deployments

Run Docker Compose workloads locally or on a remote host over SSH. A `docker` profile clones or pulls a Git repository into `deploy.target_path`, writes env files when configured, and runs `docker compose`.

**See also:** [Configuration — Docker](../reference/configuration.md#docker) · [SSH](ssh.md) · [Examples #6](../examples/README.md#6-docker-compose-local) · [Examples #7](../examples/README.md#7-docker-compose-over-ssh)

---

## How it works

1. If a Compose stack is already running and `stop_before_sync` is enabled (default), Pablo runs `docker compose down` (without `-v`).
2. It clones or pulls the Git repository into `deploy.target_path`.
3. It writes environment variables to an env file when `env_file` / `variables` are configured.
4. It runs `docker compose up` with your compose file.

```yaml
name: stack
version: 0.1.0

profiles:
  default:
    type: docker
    git:
      repo: https://github.com/user/my-stack.git
      branch: main
    environments:
      local:
        deploy:
          target_path: ./runtime
          docker:
            compose_file: docker-compose.yml
            build: true
```

Docker profiles require `git.repo`, `deploy.target_path`, and `deploy.docker.compose_file`. They forbid `deploy.source` and `register_path`.

---

## Requirements

- Docker CLI and Docker Compose v2 on the target machine
- Git for clone/pull
- For remote deploy: SSH access ([SSH guide](ssh.md))

---

## Local vs remote

| | Local | Remote SSH |
|---|-------|------------|
| `remote` block | Omitted | Required |
| Clone location | `deploy.target_path` on your machine | `deploy.target_path` on the remote host |
| Compose runs | Locally | Via SSH on the remote |

---

## Compose configuration

| Field | Default | Description |
|-------|---------|-------------|
| `compose_file` | *(required)* | Path relative to repo root after clone |
| `build` | `false` | Pass `--build` to compose |
| `stop_before_sync` | `true` | Stop a running stack before git sync |

```yaml
deploy:
  target_path: /opt/my-stack
  docker:
    compose_file: docker-compose.yml
    build: true
    stop_before_sync: true
```

On redeploy, Pablo detects running containers with `docker compose ps -q` and stops them before `git pull` so bind mounts and dirty trees do not break sync. Volumes are kept. Default `true` means a short downtime on redeploy; set `stop_before_sync: false` to keep the old pull-while-running behavior.

---

## Environment variables

Pablo **writes** an env file in the clone directory from environment `variables` when `env_file` is set and the map is non-empty. It does not read an existing `.env` as input. For compile-time files under a local `build.path` (e.g. Vite `.env.production` on a `static` profile), put values in environment `variables` and set `build.env_file` — see [Configuration — Variables and env files](../reference/configuration.md#variables-and-env-files).

```yaml
name: stack
version: 0.1.0

profiles:
  default:
    type: docker
    git:
      repo: https://github.com/user/my-stack.git
      branch: main
    environments:
      production:
        variables:
          APP_PORT: "8080"
          DB_URL: "postgres://db/app"
        env_file: .env
        deploy:
          target_path: /opt/my-stack
          docker:
            compose_file: docker-compose.yml
            build: true
```

Template substitution (`{{VAR}}`) applies to config files in the deployed tree.

---

## Private repositories

Reference a token credential for HTTPS Git URLs:

```yaml
name: stack
version: 0.1.0

credentials:
  github:
    type: token
    value: ghp_xxxxxxxx

profiles:
  default:
    type: docker
    git:
      repo: https://github.com/org/private-repo.git
      branch: main
      credential: github
    environments:
      production:
        deploy:
          target_path: ./runtime
          docker:
            compose_file: docker-compose.yml
            build: true
```

Prefer environment-injected tokens in CI rather than committing secrets. See [Credentials](credentials.md).

---

## Post-deploy commands

Use `deploy.post_commands` for tasks Compose does not handle (migrations, cache warm-up). On remote hosts these run over SSH after compose up:

```yaml
deploy:
  target_path: /opt/my-stack
  docker:
    compose_file: docker-compose.yml
  post_commands:
    - docker compose exec -T api ./migrate up
```

---

## Limitations

- Pablo wraps `docker compose`; it does not manage individual container lifecycle beyond compose.
- Use compose `restart` policies or `post_commands` for service restarts.
- The remote host must have a Docker daemon, and your SSH user must be allowed to run Docker.

---

## Example fixtures

- Local-shaped sample: [Examples #6](../examples/README.md#6-docker-compose-local)
- Remote E2E: [tests/e2e/scenarios/compose-api](../../tests/e2e/scenarios/compose-api/)
- Separate-apps backend: [tests/agnostic/separate-apps/backend](../../tests/agnostic/separate-apps/backend/)
