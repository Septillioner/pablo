# TEST_SPEC — Pablo Test Katalogu

Her satir: **hangi sistem**, **hangi katmanda**, **ne dogrulanir**. Strateji icin [TEST_PLAN.md](TEST_PLAN.md).

**Durum:** `done` | `planned`

---

## Config ve manifest

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-CFG-01 | YAML yukleme | unit | `config/loader_test.go` | Minimal manifest, BaseDir, kalitim (variables, build, output_dir) | done |
| U-CFG-02 | YAML hatalar | unit | `config/loader_test.go` | Eksik dosya, gecersiz YAML | done |
| M-CFG-01 | Manifest ornekleri | manuel | `agnostic/inheritance-test/` | Elle `pablo check` | done |

---

## Artifact filter ve path

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-FILTER-01 | Glob eslestirme | unit | `filter/filter_test.go` | Match: glob, `/` deseni, ayirici normalizasyonu | done |
| U-FILTER-02 | Dosya listeleme | unit | `filter/filter_test.go` | GetFiles: include/exclude, bos dizin, hata | done |
| U-PATH-01 | Remote path | unit | `pathutil/pathutil_test.go` | JoinRemote, DirRemote POSIX cikti | done |

---

## Template

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-TPL-01 | Uzanti filtresi | unit | `template/template_test.go` | isConfigExt | done |
| U-TPL-02 | Degisken degistirme | unit | `template/template_test.go` | replaceVariables, ProcessFiles | done |

---

## Deployer

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-DEP-01 | Korumali path | unit | `deployer/service_test.go` | isProtectedPath (Unix/Windows) | done |
| U-DEP-02 | Stratejiler | unit | `deployer/service_test.go` | overwrite, backup, recreate | done |
| U-DEP-03 | blue-green | unit | `deployer/service_test.go` | Henuz implemente degil → hata | done |
| U-DEP-04 | Dosya kopyalama | unit | `deployer/service_test.go` | copyFile mode | done |
| U-DEP-05 | Remote deploy | unit | `deployer/service_test.go` | DeployRemote mock SSH, tar/scp | done |

---

## Health ve hooks

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-HTH-01 | HTTP health | unit | `health/health_test.go` | Bos URL, 200, 5xx, retry, timeout | done |
| U-HOOK-01 | Hook calistirma | unit | `hooks/hooks_test.go` | Bos komut, env, workingDir, hata | done |

---

## System PATH

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-SYS-01 | PATH ekleme | unit | `system/path_test.go` | Scope routing, Unix idempotency | done |
| U-SYS-02 | PATH kaldirma | unit | `system/path_test.go` | Unix user/system, Windows | done |
| M-SYS-01 | Windows servis | manuel | `windows/nssm-service/` | NSSM post_commands | done |

---

## SSH adapter

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-SSH-01 | Path genisletme | unit | `ssh/ssh_test.go` | expandPath `~/` | done |
| U-SSH-02 | Connect hatalar | unit | `ssh/ssh_test.go` | Desteklenmeyen tip, eksik key/password, okunamayan key | done |
| U-SSH-03 | Tar stream | unit | `ssh/ssh_test.go` | addToTar goreli yol ve icerik | done |
| U-SSH-04 | Remote komut | unit | — | ExecuteCommand / CreateBackup mock | planned |
| E-SSH-01 | Static remote deploy | e2e | `TestSSH_StaticDeploy` | SSH tar deploy, dosya varligi | done |
| E-SSH-02 | Docker remote deploy | e2e | `TestSSH_DockerRemoteDeploy` | git clone + compose, container ayakta | done |

---

## Pipeline helpers

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-PIP-01 | resolvePath | unit | `pipeline/helpers_test.go` | Goreli/mutlak yol, baseDir | done |
| U-PIP-02 | resolveVariables | unit | `pipeline/helpers_test.go` | env.Deploy.Variables kopyasi | done |
| U-PIP-03 | resolveArtifacts | unit | `pipeline/helpers_test.go` | output_dir vs deploy.source onceligi | done |
| U-PIP-04 | writeEnvFile | unit | `pipeline/helpers_test.go` | KEY=VALUE format, bos map | done |
| U-PIP-05 | runCommands | unit | `pipeline/helpers_test.go` | Bos liste, lokal echo, remote SSH hatasi | done |

---

## SCM (git)

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-SCM-01 | CloneOrPull | unit | `scm/service_test.go` | nil config, gecersiz URL, temp repo clone/pull | done |
| E-SCM-01 | Remote git | e2e | `ssh-docker-remote` | file:// bare repo clone uzakta | done |

---

## Docker adapter

| ID | Sistem | Katman | Konum | Dogruladigi davranis | Durum |
|----|--------|--------|-------|----------------------|-------|
| U-DOCK-01 | Compose arg | unit | `docker/adapter_test.go` | composeUpArgs, composeDownArgs | done |
| U-DOCK-02 | Gercek daemon | unit | — | integration tag ile docker calistirma | planned |
| E-DOCK-01 | Remote compose | e2e | `TestSSH_DockerRemoteDeploy` | Uzak `docker compose up` | done |

---

## Manuel fixture indeksi

| Dizin | Amac |
|-------|------|
| `agnostic/local-deploy/` | Lokal static deploy |
| `agnostic/multi-profile/` | Coklu profil tek YAML |
| `agnostic/separate-apps/` | Profil basina ayri manifest |
| `agnostic/inheritance-test/` | Kalitim kurallari |
| `agnostic/self-deploy/` | Pablo kendi kendine deploy |
| `windows/nssm-service/` | Windows binary + NSSM |

---

## Guncelleme

Yeni test eklendiginde bu tabloya satir ekleyin veya `planned` → `done` yapin.
