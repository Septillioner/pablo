# Security Policy

## Reporting a vulnerability

If you discover a security vulnerability in Pablo, please report it privately. **Do not open a public GitHub issue.**

Preferred channels:

1. Open a [private security advisory](https://github.com/septillioner/pablo/security/advisories/new) on GitHub, or
2. Email the maintainer directly via the contact information listed on [github.com/septillioner](https://github.com/septillioner).

Please include:

- A description of the vulnerability and its potential impact.
- Steps to reproduce, ideally with a minimal `pablo.yaml`.
- The Pablo version (`pablo version`) and host OS / arch.
- Any suggested mitigations or patches.

You should receive an acknowledgement within a reasonable timeframe. Coordinated disclosure is appreciated; please give the maintainer a chance to ship a fix before publishing details.

## Supported versions

Pablo is pre-1.0 in spirit and currently follows the version embedded in `src/VERSION`. Only the latest released minor version receives security fixes. Older versions may be patched on a best-effort basis.

## Known security considerations

These are intentional or known limitations users should be aware of when operating Pablo in sensitive environments:

- **SSH host key verification is enabled by default.** Pablo checks the remote host key against the OpenSSH `known_hosts` file (`~/.ssh/known_hosts`, or `%USERPROFILE%\.ssh\known_hosts` on Windows). Unknown hosts fail with a fingerprint and instructions to add the key (`ssh-keyscan` or an interactive `ssh` login). Optional `remote.trust_on_first_use: on` records an unknown key on first connect (default off). Opt out per environment with `remote.host_key_verification: off` (emits a warning; not recommended on untrusted networks).
- **No sandboxing of build / deploy commands.** Pablo executes `build.command`, `deploy.pre_commands`, and `deploy.post_commands` with the privileges of the invoking user. Treat `pablo.yaml` as trusted input.
- **Template substitution does not escape values.** `{{VAR}}` replacement is a literal string substitution. Do not place untrusted user input into template variables that flow into shell commands or generated config.
- **Protected path detection is shallow.** Only top-level system directories are checked. The `--force` flag bypasses this guard entirely.
- **Backups may contain sensitive files.** `backup` strategy renames the existing target directory with a timestamp suffix. These backups are not encrypted and remain on the target until removed (`pablo uninstall --remove-backups`).
- **Credentials in manifests.** Avoid committing `pablo.yaml` files that embed plaintext passwords or private keys. Prefer SSH key files referenced by path and environment-injected secrets.

## Scope

In scope:

- The Pablo CLI and embedded language server (`pablo lsp`, `src/internal/lsp/`)
- The VS Code extension (`extensions/vscode-pablo/`)

Out of scope:

- Vulnerabilities in transitive dependencies that are already publicly tracked upstream (please report those to the upstream project).
- Issues that require an attacker to already have local code execution or write access to `pablo.yaml`.
