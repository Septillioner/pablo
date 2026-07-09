# Pablo E2E Tests

Docker tabanli integration testleri: Ubuntu SSH hedefi uzerinde remote deploy senaryolarini dogrular.

## Gereksinimler

- Docker Desktop / Docker Engine (WSL2 backend onerilir — Windows)
- Go 1.25.5+
- OpenSSH `ssh-keygen` (PATH uzerinde)

## Calistirma

```powershell
cd tests/e2e
go test -tags=integration -v -timeout 10m ./...
```

Ilk calistirmada Docker image build edilir (~1-2 dk). Testler sirasinda `127.0.0.1:2222` uzerinde SSH hedefi ayaga kalkar.

## Senaryolar

| Test | Aciklama |
|------|----------|
| `TestSSH_StaticDeploy` | SSH ile static artifact deploy (`/tmp/pablo-e2e-static`) |
| `TestSSH_DockerRemoteDeploy` | SSH uzerinden git clone + uzak `docker compose up` |

## Mimari

- **Hedef:** `docker/docker-compose.yml` — Ubuntu 24.04, OpenSSH, git, docker CLI
- **Docker socket:** Host `/var/run/docker.sock` mount — uzak `docker compose` host daemon uzerinde calisir
- **Anahtarlar:** `keys/` dizini test basinda uretilir (gitignore)

## Bilinen kisitlar

- Socket mount nedeniyle compose icindeki bind-mount yollari **host** dosya sistemine gore cozulur; fixture'larda bind-mount kullanmayin.
- E2E container'daki docker CLI, host API surumu ile uyum icin `DOCKER_API_VERSION=1.43` wrapper kullanir.
- `PABLO_E2E_SKIP_DOCKER=1` ortam degiskeni ile container baslatmadan test calistirilabilir (yalnizca gelistirme).

## Dizin yapisi

```
tests/e2e/
├── docker/           # Ubuntu SSH hedef image + compose
├── fixtures/         # Ornek docker uygulamasi (bare repo kaynagi)
├── scenarios/        # pablo.yaml senaryolari
├── keys/             # Uretilen SSH anahtarlari (gitignore)
├── helpers.go
└── e2e_test.go
```
