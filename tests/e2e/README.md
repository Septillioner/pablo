# Pablo E2E Tests

Docker-based integration tests: real deploy stories against an Ubuntu SSH target.

## Requirements

- Docker Desktop / Docker Engine
- Go 1.25.5+
- OpenSSH `ssh-keygen` on PATH

## Run

```bash
cd tests/e2e
go test -tags=integration -v -timeout 15m ./...
```

First run builds the target image (~1–2 min). SSH listens on `127.0.0.1:2222`.

## Scenarios (stories)

| Test | Dir | Story |
|------|-----|-------|
| `TestSSH_StaticSite` | `scenarios/static-site/` | Marketing site → `/var/www/static-site` (`overwrite`) |
| `TestSSH_StaticSiteHotfix` | `scenarios/static-site-hotfix/` | Live HTML/CSS swap (`rename-replace`) |
| `TestSSH_GoService` | `scenarios/go-service/` | Go binary → `/opt/go-service` + `post_commands` |
| `TestSSH_ComposeAPI` | `scenarios/compose-api/` | Git + Compose stack; redeploy while up |
| `TestSSH_PHPApp` | `scenarios/php-app/` | PHP app via `git-sync` → `/var/www/php-app` |
| `TestSSH_ReleaseSequence` | `scenarios/release-sequence/` | Staging then prod (`pablo run sequence`) |
| `TestSSH_SiteWithBackup` | `scenarios/site-with-backup/` | Prod deploy keeps previous tree (`backup`) |
| `TestSSH_CleanRedeploy` | `scenarios/clean-redeploy/` | Wipe stale files (`recreate`) |
| `TestSSH_LegacyTransfer` | `scenarios/legacy-transfer/` | Multi-file site over SCP (`transfer: legacy`) |
| `TestSSH_VerifiedTransfer` | `scenarios/verified-transfer/` | Post-transfer SHA-256 (`verify_checksum`) |

All manifests are Schema v2. Strategies covered: `overwrite`, `rename-replace`, `backup`, `recreate`.

## Architecture

- **Target:** `docker/docker-compose.yml` — Ubuntu 24.04, OpenSSH, git, docker CLI
- **Docker socket:** Host `/var/run/docker.sock` mount for remote Compose
- **Fixtures:** `fixtures/sample-docker-app`, `fixtures/sample-php-app` (bare repos created in entrypoint)
- **Keys:** `keys/` generated at test start (gitignored)

## Limits

- No bind-mounts inside Compose fixtures (socket mount resolves host paths).
- Container docker CLI uses `DOCKER_API_VERSION=1.43` wrapper.
- `PABLO_E2E_SKIP_DOCKER=1` skips container bring-up (dev only).

## Layout

```
tests/e2e/
├── docker/           # SSH target image + compose
├── fixtures/         # Bare-repo sources (docker, php)
├── scenarios/        # One story per directory
├── keys/             # Generated SSH keys
├── helpers.go
└── e2e_test.go
```
