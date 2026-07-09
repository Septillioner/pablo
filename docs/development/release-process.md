# Release Process

Summary of how Pablo releases are cut. **Canonical document:** [RELEASING.md](../../RELEASING.md)

---

## Versioning

- **Source of truth:** [`src/VERSION`](../../src/VERSION) (no `v` prefix)
- **Tags:** `v1.3.0` format on GitHub
- **Semver:** MAJOR.MINOR.PATCH per [semver.org](https://semver.org)

---

## Build matrix

`./build.sh all` produces:

| OS | Arch | Filename |
|----|------|----------|
| darwin | amd64 | `pablo-darwin-amd64` |
| darwin | arm64 | `pablo-darwin-arm64` |
| linux | amd64 | `pablo-linux-amd64` |
| windows | amd64 | `pablo-windows-amd64.exe` |
| windows | arm64 | `pablo-windows-arm64.exe` |

Output: `build/` (gitignored).

---

## Pre-release checklist

- [ ] Clean working tree on `main`
- [ ] `src/VERSION` bumped
- [ ] [CHANGELOG.md](../../CHANGELOG.md) updated
- [ ] [docs/](../README.md) accurate for any schema/CLI changes
- [ ] Manual fixture validation for affected deployment types

---

## Release steps (summary)

1. Bump `src/VERSION` and update `CHANGELOG.md`
2. `./build.sh all`
3. Generate `checksums.txt` in `build/`
4. Tag: `git tag -a vX.Y.Z -m "Pablo vX.Y.Z"` and push
5. Create GitHub Release with binaries + `checksums.txt` + `release-notes.md`
6. Verify downloaded binary: `pablo version` + checksum

---

## Pablo self-release (`pablo-sepy.yaml`)

The repository uses a separate manifest for packaging:

| Profile | Purpose |
|---------|---------|
| `cli-release` | Cross-platform CLI binaries → `dist/releases/` |
| `extension` | VSIX package → `dist/releases/extension/` |
| `cli-install` | Local install helpers |

```bash
pablo run -f pablo-sepy.yaml -p cli-release -e production
```

---

## Hotfix and rollback

- **Hotfix:** branch from release tag, patch, bump patch version, re-release.
- **Rollback:** mark bad release as pre-release on GitHub; ship a new patch.

Details: [RELEASING.md](../../RELEASING.md)
