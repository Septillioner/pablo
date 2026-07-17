# Credentials

Define reusable authentication once at the root of `pablo.yaml`, then reference credentials by name from `remote.credential` or `git.credential`. Pablo does not invent a default SSH key.

**See also:** [Configuration — Credential](../reference/configuration.md#credential) · [Examples #13](../examples/README.md#13-credentials) · [SECURITY.md](../../SECURITY.md)

---

## Credential types

| Type | Use for | Required fields |
|------|---------|-----------------|
| `ssh` | Remote deploy, remote git over SSH | `username` + (`key` or `password`) |
| `token` | Private Git HTTPS repos | `value` |
| `basic` | HTTP basic auth | `username`, `password` |

---

## SSH credentials

Key-based auth (recommended):

```yaml
credentials:
  prod-server:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519
    passphrase: ""
```

Password auth (prefer env injection, not plaintext in git):

```yaml
credentials:
  staging:
    type: ssh
    username: deploy
    password: "${DEPLOY_PASSWORD}"
```

Attach the name to an environment:

```yaml
remote:
  host: 10.0.0.5
  credential: prod-server
```

---

## Token credentials

For private Git repositories over HTTPS:

```yaml
name: stack
version: 0.1.0

credentials:
  github:
    type: token
    value: ghp_xxxxxxxxxxxx

profiles:
  app:
    type: docker
    git:
      repo: https://github.com/org/private.git
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

---

## Basic auth

```yaml
credentials:
  registry:
    type: basic
    username: dockeruser
    password: "${REGISTRY_PASSWORD}"
```

A complete side-by-side sample of `ssh`, `token`, and `basic` is in [Examples #13](../examples/README.md#13-credentials). The multi-profile fixture exercises all three: [tests/agnostic/multi-profile](../../tests/agnostic/multi-profile/).

---

## Security practices

1. Do not commit secrets. Prefer `pablo_local.yaml` (gitignored) or environment variables for sensitive values.
2. Prefer SSH keys over passwords where possible.
3. Restrict key permissions — `chmod 600` on private keys.
4. Rotate tokens if a manifest was accidentally committed.
5. Treat manifests as trusted input — Pablo executes build and deploy commands with your privileges.

Report vulnerabilities: [SECURITY.md](../../SECURITY.md).

---

## Multiple credentials

Name credentials by purpose so environments can pick different ones in the same file:

```yaml
credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/prod
  staging-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/staging
  github-ci:
    type: token
    value: "${GITHUB_TOKEN}"
```
