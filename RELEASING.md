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
- [ ] [`schema/schema.md`](schema/schema.md) reflects any schema changes
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

## Rollback

GitHub Releases cannot be edited destructively without losing download counts. To rollback:

1. Mark the bad release as **pre-release** or **draft** on the GitHub UI
2. Cut a new patch release with the fix
3. Update the README install instructions only if the latest-release link is wrong
