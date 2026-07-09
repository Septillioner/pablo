# Credentials

Define reusable authentication in your manifest and reference by name.

**See also:** [Configuration — Credential](../reference/configuration.md#credential) · [SECURITY.md](../../SECURITY.md)

---

## Credential types

| Type | Use for | Required fields |
|------|---------|-----------------|
| `ssh` | Remote deploy, remote git over SSH | `username` + (`key` or `password`) |
| `token` | Private Git HTTPS repos | `value` |
| `basic` | HTTP basic auth (future integrations) | `username`, `password` |

---

## SSH credentials

**Key-based (recommended):**

```yaml
credentials:
  prod-server:
    type: ssh
    username: deploy
    key: ~/.ssh/id_deploy
    passphrase: ""    # optional
```

**Password auth:**

```yaml
credentials:
  staging:
    type: ssh
    username: deploy
    password: "${DEPLOY_PASSWORD}"   # prefer env injection, not plaintext
```

Reference in an environment:

```yaml
remote:
  method: ssh
  host: 10.0.0.5
  credential: prod-server
```

---

## Token credentials

For private Git repositories:

```yaml
credentials:
  github:
    type: token
    value: ghp_xxxxxxxxxxxx

profiles:
  app:
    type: docker
    git:
      repo: https://github.com/org/private.git
      credential: github
```

---

## Security practices

1. **Do not commit secrets.** Add `pablo_local.yaml` or use env vars for sensitive values.
2. **Use SSH keys** instead of passwords where possible.
3. **Restrict key permissions** — `chmod 600` on private keys.
4. **Rotate tokens** if a manifest was accidentally committed.
5. **Treat manifests as trusted input** — Pablo executes build and hook commands with your privileges.

Report vulnerabilities: [SECURITY.md](../../SECURITY.md).

---

## Multiple credentials

Name credentials by purpose:

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
    value: ghp_xxx
```

Different environments can reference different credentials within the same manifest.
