# TEST_PLAN — Pablo Test Stratejisi

Bu belge **nasıl** test edildiğini anlatır. Hangi testin neyi doğruladığı için bkz. [TEST_SPEC.md](TEST_SPEC.md). Backlog için bkz. [../docs/roadmap.md](../docs/roadmap.md).

---

## Katmanlar

| Katman | Konum | Amaç | Gereksinim |
|--------|-------|------|------------|
| **Unit** | `src/**/*_test.go` | Paket mantığı; ağ/daemon yok | Go 1.25.5+ |
| **E2E** | `tests/e2e/` | Gerçek SSH + uzak docker (Ubuntu container) | Docker, `ssh-keygen` |
| **Manuel fixture** | `tests/agnostic/`, `tests/windows/` | YAML senaryoları; elle `pablo check` / `run` | Derlenmiş `pablo` veya `go run` |

```mermaid
flowchart LR
    unit[Unit src]
    e2e[E2E Docker SSH]
    manual[Manuel fixtures]
    unit -->|hizli regresyon| ci[CI hedefi]
    e2e -->|remote akis| ci
    manual -->|senaryo ornekleri| dev[Gelistirici]
```

---

## Komutlar

### Unit (her degisiklikten sonra)

```bash
cd src
go test ./...
```

### E2E (remote SSH / docker)

```bash
cd tests/e2e
go test -tags=integration -v -timeout 15m ./...
```

### Manuel fixture

```bash
# Manifest dogrulama
go run ./src/main.go check -f tests/agnostic/local-deploy/pablo.yaml

# Deploy (dikkat: gercek dosya kopyalar)
cd tests/agnostic/local-deploy
go run ../../../src/main.go run -e production
```

---

## Kapsam matrisi (sistem → katman)

| Sistem | Unit | E2E | Manuel fixture |
|--------|------|-----|----------------|
| Config / kalitim | x | | x |
| Artifact filter | x | | |
| Remote path (pathutil) | x | | |
| Template `{{VAR}}` | x | | |
| Deployer / stratejiler | x | x (overwrite, backup, recreate, rename-replace) | |
| Health check HTTP | x | | |
| Hooks / post_commands | x | x (Go service post_commands) | |
| PATH (system adapter) | x | | x (windows) |
| SSH adapter | x | x | |
| Pipeline helpers | x | | |
| SCM (git) | x | x (compose + php git-sync) | x |
| Docker adapter | x | x (compose-api + redeploy) | x |
| Types: static / binary / docker / git-sync | | x | x |
| Sequences | | x | x |
| `transfer: legacy` / `verify_checksum` | | x | |
| Tam pipeline `Run` | | x | x |

E2E story index: [e2e/README.md](e2e/README.md).

---

## Kurallar

1. Yeni unit test: `src/internal/.../foo_test.go` — test edilen paketle ayni dizin.
2. Ag gerektiren unit test yazma; remote davranis E2E'de.
3. Yeni test eklendiginde [TEST_SPEC.md](TEST_SPEC.md) guncellenir.
4. `docs/roadmap.md` backlog tamamlaninca SPEC'te `done` isaretlenir.
5. SCM unit testleri `git` CLI kullanir; yoksa test `Skip`.

---

## CI hedefi (henuz yok)

- PR: `cd src && go test ./...`
- Opsiyonel job: `cd tests/e2e && go test -tags=integration ./...` (Docker runner)
- Coverage: `go test -coverprofile=coverage.out ./...`

---

## Dizin haritasi

```
tests/
├── TEST_PLAN.md      # bu dosya
├── TEST_SPEC.md      # test katalogu
├── README.md         # kisa indeks
├── e2e/              # integration (Go + Docker)
├── agnostic/         # manuel YAML ornekleri
└── windows/          # Windows ozel fixture
```
