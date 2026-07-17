# Troubleshooting

Diagnose common Pablo CLI and VS Code extension issues.

---

## Validation errors (`pablo check`)

### `path:line:col` error messages

Pablo prints semantic validation errors with file location:

```
pablo.yaml:24:11: environments.production.deploy.target_path is required
```

Fix the field at that location. Rules are shared between CLI and LSP.

### Profile or environment not found

```
profile "api" not found
```

Check spelling and that the profile exists under `profiles:` in your manifest. List with `pablo inspect`.

### Sequence step errors

```
sequence "release" step 1: target must be profile/environment
profile "missing" not found
```

Each sequence step must be `profile/env` and both keys must exist. Empty sequences are invalid. Validate with `pablo check`; list sequences with `pablo inspect`. Guide: [Sequences](guides/sequences.md).

### Run blocked despite passing `check`

`pablo run` re-validates the manifest. Ensure `-p` and `-e` match existing keys. For sequences, use `pablo run sequence <name>` without `-p` / `-e`.

---

## Build failures

| Symptom | Check |
|---------|-------|
| Command not found | Is the tool (npm, go, etc.) on PATH where Pablo runs? |
| Wrong directory | Set `build.path` relative to manifest location |
| Env vars missing | Add `build.variables` or `build.env_file` |

Build runs with your user privileges — no sandbox.

---

## Deploy failures

### Protected system path

```
refusing to deploy to protected path
```

Pablo blocks `backup`/`recreate` against system directories. Use a safer `target_path` or pass `--force` (dangerous).

### Empty artifact set

Check `output_dir` / `deploy.source` `include` globs match built files. Run build manually first to confirm output location.

If too many files match, remember that `*.exe` matches at every depth; use `/*.exe` or `./*.exe` for the artifact root only. See [Configuration — OutputDir](reference/configuration.md#outputdir).

### Health check timeout

`pipeline.health_check` retries HTTP GET for 30 seconds expecting status 200. Verify URL, server start time, and firewall rules on the target.

---

## SSH issues

| Symptom | Fix |
|---------|-----|
| Connection refused | Host reachable, port 22 open, SSH daemon running |
| Permission denied (publickey) | Correct `username`, `key` path, key permissions (`chmod 600`) |
| Permission denied (password) | Set `password` in credential or use key auth |
| Remote command fails | Test SSH manually: `ssh user@host 'ls /opt/app'` |
| Slow transfer | Ensure `deploy.remote` is `tar` (default), not `legacy` |

**Security:** host key verification is on by default — see [SSH guide](guides/ssh.md) and [SECURITY.md](../SECURITY.md).

---

## Docker issues

| Symptom | Fix |
|---------|-----|
| `docker: command not found` | Install Docker on target host |
| Compose file not found | `compose_file` path relative to cloned repo root |
| Permission denied (docker) | Add SSH user to `docker` group on Linux |
| Git pull fails while containers run | Default `stop_before_sync: true` stops the stack before sync; ensure it is not set to `false` |
| Private repo clone fails | Set `git.credential` with valid token |

---

## PATH registration

| Symptom | Fix |
|---------|-----|
| Binary not found after deploy | Open new shell (PATH changes need reload) |
| System scope fails | Run elevated / with sudo where required |
| Wrong binary on PATH | Check for duplicate installs; use `which pablo` |

Windows: PATH updated via PowerShell. macOS system scope uses `/etc/paths.d/`. Linux system scope uses `/etc/profile.d/pablo.sh`.

---

## VS Code extension

| Symptom | Fix |
|---------|-----|
| "Pablo not found" dialog | Install CLI or **Pablo: Select Executable** |
| No LSP features | Binary must be 1.3+ with `pablo lsp`; old PATH binary may lack LSP |
| Multiple binaries on PATH | Pick correct one in selector |
| LSP EPIPE / crash loop | Extension disables auto-restart; fix binary path, reload window |
| CodeLens missing | Reload file; ensure environment block is valid YAML |

**Debug:**

1. **View → Output → Pablo Language Server**
2. Set `"pablo.trace.server": "verbose"`
3. Confirm `pablo.path` points to current build

Full guide: [VS Code](guides/vscode.md)

---

## Visual Studio extension

| Symptom | Fix |
|---------|-----|
| No LSP / CodeLens | Pablo CLI 1.3+ with `pablo lsp`; **Tools → Pablo: Select Executable**; open a `pablo*.yaml` |
| Run cannot find manifest | Focus or reopen the manifest, or use the **Pablo** toolbar Manifest combo |
| Tool window: missing executable / inspect error | Select a valid CLI binary, then **Refresh** |
| Terminal path / quoting errors on Run | Update the extension; Run uses shell-aware quoting (`cmd /s /k` on cmd) |
| F5: no Pablo commands in Tools | Rebuild **Debug**, close both VS windows, F5 again (extension loads only in Experimental Instance) |

**Debug:** **View → Output → Pablo Language Server**

Full guide: [Visual Studio](guides/visual-studio.md)

---

## Exit codes

Pablo exits `1` on any error, `0` on success. No separate validation code. See [exit codes](reference/exit-codes.md).

---

## Getting help

Open a GitHub issue with:

- `pablo version` output
- Host OS / arch
- Target OS (if remote)
- Minimal `pablo.yaml` reproducer
- Full CLI output

Security issues: [SECURITY.md](../SECURITY.md) (not public issues).
