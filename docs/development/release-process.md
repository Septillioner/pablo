# Release Process

How Pablo releases are cut and published.

Scripts resolve the repo root from their own location — run them from anywhere:

```bat
scripts\publish-release.bat 1.8.0
```

```bash
./scripts/publish-release.sh 1.8.0
```

---

## Versioning

- **Source of truth:** [`src/VERSION`](../../src/VERSION) (no `v` prefix)
- **Tags:** `v1.3.0` format on GitHub
- **Semver:** MAJOR.MINOR.PATCH per [semver.org](https://semver.org)

---

## Build matrix

`./scripts/build.sh all` produces:

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
- [ ] `CHANGELOG.md` has a `## [X.Y.Z]` section for the new version
- [ ] `release-notes.md` updated for `vX.Y.Z`
- [ ] [docs/](../README.md) accurate for any schema/CLI changes
- [ ] Manual fixture validation for affected deployment types

Version files (`src/VERSION`, extension manifests) are bumped by the release script.

---

## Maintainer publish scripts (two-step)

Releases are split on purpose. **Both scripts default to dry-run** (print the plan and exit). Real publish requires `--confirm` **and** typing `YES` at the prompt. Do not run `--confirm` unless you intend to ship.

### 1. CLI + GitHub Release

| Script | Platform |
|--------|----------|
| `scripts/publish-release.bat` | Windows (calls `publish-release.ps1`) |
| `scripts/publish-release.sh` | macOS / Linux |

```bat
scripts\publish-release.bat 1.8.0
scripts\publish-release.bat 1.8.0 --confirm
```

```bash
./scripts/publish-release.sh 1.8.0
./scripts/publish-release.sh 1.8.0 --confirm
```

What `--confirm` does (after `YES`):

1. Bump `src/VERSION` and extension version fields
2. `git commit` (`chore: release vX.Y.Z`)
3. Fresh cross-platform build into `build/`
4. Write `build/checksums.txt` (SHA-256, LF line endings)
5. Annotated tag `vX.Y.Z` (no force retag)
6. `git push` commit + tag (no `--force`)
7. `gh release create` with CLI binaries + `checksums.txt` + `release-notes.md`

**Tools:** `go`, `git`, `gh` (authenticated).

**Optional env:**

| Variable | Effect |
|----------|--------|
| `PABLO_RELEASE_ALLOW_DIRTY=1` | Skip clean working-tree check |
| `PABLO_RELEASE_SKIP_PUSH=1` | Build + tag locally; skip push and `gh release create` |
| `GH_TOKEN` / `GITHUB_TOKEN` | Used by `gh` if not already logged in |

### 2. Extensions (marketplaces)

| Script | Platform |
|--------|----------|
| `scripts/publish-extensions.bat` | Windows (VS Code + VS2026 VSIX) |
| `scripts/publish-extensions.sh` | macOS / Linux (VS Code VSIX only; VS2026 needs MSBuild) |

```bat
scripts\publish-extensions.bat 1.8.0
scripts\publish-extensions.bat 1.8.0 --confirm
```

```bash
./scripts/publish-extensions.sh 1.8.0
./scripts/publish-extensions.sh 1.8.0 --confirm
```

What `--confirm` does (after `YES`):

1. `npm ci` + `npm run vsix` in `extensions/vscode-pablo`
2. On Windows: `extensions/vs2026/build-vs2026.bat` → `build/pablo-vs2026-X.Y.Z.vsix`
3. `npm run publish:marketplace` (VS Code Marketplace / `vsce`)
4. `npm run publish:ovsx` (Open VSX)
5. Upload VSIX assets to the existing GitHub release `vX.Y.Z` when present

**Tokens (never commit these):**

| Variable | Purpose |
|----------|---------|
| `VSCE_PAT` | VS Code Marketplace publish (`vsce`) |
| `OVSX_PAT` | Open VSX publish (`ovsx`) |
| `GH_TOKEN` / `GITHUB_TOKEN` | Optional; `gh release upload` |

**Optional env:** `PABLO_EXT_SKIP_MARKETPLACE=1`, `PABLO_EXT_SKIP_OVSX=1`, `PABLO_EXT_SKIP_VS2026=1`, `PABLO_EXT_SKIP_GH_UPLOAD=1`.

Visual Studio Gallery upload for the VS2026 VSIX is **not** automated — distribute that package manually if needed.

---

## Manual release steps (summary)

If you prefer not to use the scripts:

1. Bump `src/VERSION` and update `CHANGELOG.md` / `release-notes.md`
2. `./scripts/build.sh all`
3. Generate `checksums.txt` in `build/`
4. Tag: `git tag -a vX.Y.Z -m "Pablo vX.Y.Z"` and push
5. Create GitHub Release with binaries + `checksums.txt` + `release-notes.md`
6. Publish extensions via `npm run publish:marketplace` / `npm run publish:ovsx`
7. Verify downloaded binary: `pablo version` + checksum

---

## Pablo self-release (`pablo-sepy.yaml`)

The repository may use a separate private manifest for packaging:

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

- **Hotfix:** branch from the release tag, patch, bump the patch version, re-release (do not force-push tags).
- **Rollback:** mark the bad release as pre-release on GitHub; ship a new patch.
