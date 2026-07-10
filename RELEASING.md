# Releasing Pablo

This document describes the manual release process. The single source of truth for the version is [`src/VERSION`](src/VERSION).

## Versioning

Pablo follows [Semantic Versioning](https://semver.org/):

- **MAJOR** — incompatible API or manifest schema changes
- **MINOR** — backward-compatible features
- **PATCH** — backward-compatible bug fixes

Tag names use a leading `v` (e.g. `v1.0.46`). The version inside `src/VERSION` does **not** include the `v` prefix.

## Release matrix

The `build.sh all` script produces these artifacts:

| OS      | Arch  | Output filename             |
|---------|-------|-----------------------------|
| darwin  | amd64 | `pablo-darwin-amd64`        |
| darwin  | arm64 | `pablo-darwin-arm64`        |
| linux   | amd64 | `pablo-linux-amd64`         |
| windows | amd64 | `pablo-windows-amd64.exe`   |
| windows | arm64 | `pablo-windows-arm64.exe`   |

Artifacts are written to `build/`, which is gitignored.

## Pre-release checklist

- [ ] Working tree is clean (`git status`)
- [ ] On the `main` branch and up to date with the remote
- [ ] `src/VERSION` updated to the new version (no `v` prefix, no trailing whitespace)
- [ ] [`CHANGELOG.md`](CHANGELOG.md) updated with a new entry for this version
- [ ] [`README.md`](README.md) install / usage instructions still accurate
- [ ] [`docs/reference/configuration.md`](docs/reference/configuration.md) reflects any schema changes
- [ ] Manually validated against fixtures under `tests/` for the affected deployment types
- [ ] No new direct dependencies without justification in CHANGELOG

## Release steps

1. **Bump the version**

   ```bash
   echo "1.4.7" > src/VERSION   # use the new version
   git add src/VERSION CHANGELOG.md
   git commit -m "chore: release v1.4.7"
   ```

2. **Build all platform artifacts**

   ```bash
   ./build.sh all
   ls build/
   ```

3. **Generate checksums**

   macOS / Linux:

   ```bash
   cd build
   shasum -a 256 pablo-* > checksums.txt
   cat checksums.txt
   cd ..
   ```

   Windows (PowerShell):

   ```powershell
   cd build
   Get-ChildItem pablo-* | ForEach-Object {
     $hash = (Get-FileHash $_ -Algorithm SHA256).Hash.ToLower()
     "$hash  $($_.Name)"
   } | Out-File -Encoding ascii checksums.txt
   cd ..
   ```

4. **Create and push the tag**

   ```bash
   git tag -a v1.4.7 -m "Pablo v1.4.7"
   git push origin main
   git push origin v1.4.7
   ```

5. **Publish the GitHub Release**

   Using the GitHub CLI:

   ```bash
   gh release create v1.4.7 \
     build/pablo-darwin-amd64 \
     build/pablo-darwin-arm64 \
     build/pablo-linux-amd64 \
     build/pablo-windows-amd64.exe \
     build/pablo-windows-arm64.exe \
     build/checksums.txt \
     --title "Pablo v1.4.7" \
     --notes-file release-notes.md
   ```

   Or, via the GitHub web UI:
   - Open **Releases → Draft a new release**
   - Choose the tag `v1.4.7`
   - Title: `Pablo v1.4.7`
   - Paste the contents of `release-notes.md` into the description
   - Attach all five binaries plus `checksums.txt`
   - Publish

6. **Post-release verification**

   - Download one binary from the release page and run `pablo version`
   - Verify the SHA-256 against `checksums.txt`
   - Confirm README "Releases" link resolves

## Hotfix procedure

For a critical bug in the latest release:

1. Branch from the release tag: `git switch -c hotfix/v1.4.7-x v1.4.7`
2. Apply the minimal fix and bump `src/VERSION` (e.g. `1.4.7` → `1.4.8`)
3. Update `CHANGELOG.md`
4. Follow the standard release steps above for the new patch version
5. Merge the hotfix branch back into `main`

### Windows release scripts (local tooling)

Maintainer-only batch/PowerShell scripts live at the repo root (gitignored). They wrap build, checksums, tag, and `gh release create`.

| Script | When to use |
|--------|-------------|
| `release-new-version.bat X.Y.Z` | Normal release — you write `CHANGELOG.md` and `release-notes.md` first |
| `release-hotfix.bat "fix message"` | Code hotfix — auto patch bump (`1.5.6` → `1.5.7`), auto `CHANGELOG` + `release-notes`, then full release |
| `release-hotfix.bat 1.5.7 "fix message"` | Same as above with an explicit target version |
| `release-force.bat X.Y.Z` | **Artifact-only** rebuild — same tag overwritten (wrong `checksums.txt`, bad assets). Requires `src/VERSION` = `X.Y.Z`. Type `YES` to confirm. Does **not** bump semver. |

**Do not use** non-semver tags like `1.5.6a` — VS Code marketplace, `pablo update`, and extension assembly versions expect `MAJOR.MINOR.PATCH`.

`release-force.bat` deletes the local/remote git tag and GitHub Release for `vX.Y.Z`, rebuilds CLI + VSIXes, and republishes. Use only when release **files** were wrong, not when you need to ship new code (use `release-hotfix.bat` instead).

`release-hotfix.bat` commits generated docs (`chore: prepare vX.Y.Z hotfix`) then calls `release-new-version.bat`.

Environment flags used internally by the scripts:

| Variable | Effect |
|----------|--------|
| `PABLO_RELEASE_SKIP_DOCS=1` | Skip `CHANGELOG` / `release-notes` validation |
| `PABLO_RELEASE_ALLOW_RETAG=1` | Skip tag-exists checks |
| `PABLO_RELEASE_SKIP_BUMP=1` | Skip `src/VERSION` + extension bump commit |
| `PABLO_RELEASE_AUTO_CONFIRM=1` | Skip interactive `y/N` prompt |
| `PABLO_RELEASE_ALLOW_DIRTY=1` | Skip clean working-tree check |

## Rollback

GitHub Releases cannot be edited destructively without losing download counts. To rollback:

1. Mark the bad release as **pre-release** or **draft** on the GitHub UI
2. Cut a new patch release with the fix
3. Update the README install instructions only if the latest-release link is wrong
