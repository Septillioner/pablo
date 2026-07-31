# Blue-Green Deploy

`deploy.blue_green` writes into one of two configured slots, then runs your switch command. Pablo owns slot arbitration only — how traffic moves (symlink, systemd unit, reverse proxy, IIS site path) is entirely your command.

**See also:** [Configuration — Blue-Green](../reference/configuration.md#blue-green) · [Deploy strategies](deploy-strategies.md) · [Capabilities](../reference/capabilities.md)

---

## When to use it

Use blue-green when you need the live tree untouched while the next release is prepared, then a single cutover. Prefer a plain `strategy` (`overwrite`, `backup`, `recreate`, `rename-replace`) when a single deploy directory is enough.

---

## Flow

1. Run `detect_command` — stdout must exactly match one slot **key** (`key` if set, else `path`), or be empty (first deploy → `slots[0]`).
2. Choose the idle slot.
3. Run `pre_commands` with cwd = idle slot (slot env vars available).
4. Deploy artifacts into the idle slot (`strategy` applies there; default `recreate`).
5. Write `env_file` / templates / optional checksum inside the idle slot.
6. Run the resolved `switch_command` (slot-level overrides global).
7. `register_path` (binary) still uses `target_path`.
8. Run `post_commands` with cwd = idle slot.

---

## Slot `key` (optional)

When the target system names a slot differently than Pablo writes it (mapped drive vs local path, UNC vs drive letter, container mount), set `key` to the exact string `detect_command` returns for that slot. Empty `key` falls back to `path`. Matching is strict equality after trim — no basename or case folding.

Effective keys (`key` or `path`) must be unique across the two slots.

**Windows YAML tip:** use single backslashes in unquoted or single-quoted paths (`W:\slot\blue`). Double-backslash escapes apply only inside double-quoted YAML scalars; unquoted `W:\\slot` keeps two literal backslashes and will never match `detect_command` output.

---

## Nginx symlink example

```yaml
name: site
version: 1.0.0

credentials:
  prod-ssh:
    type: ssh
    username: deploy
    key: ~/.ssh/id_ed25519

profiles:
  frontend:
    type: static
    environments:
      production:
        remote:
          host: web.example.com
          credential: prod-ssh
        deploy:
          source:
            dir: ./dist
            include: ["**/*"]
          target_path: /var/www/app
          strategy: recreate
          blue_green:
            slots:
              - path: /var/www/app-blue
              - path: /var/www/app-green
            detect_command: cat /var/www/app/.active 2>/dev/null || true
            switch_command: >
              echo "$PABLO_TARGET_SLOT" > /var/www/app/.active &&
              ln -sfn "$PABLO_TARGET_SLOT" /var/www/app/current &&
              systemctl reload nginx
```

Point nginx `root` (or an alias) at `/var/www/app/current`. Pablo never creates `/var/www/app` — prepare the parent directory yourself before the first switch.

---

## Systemd port / path flip example

When each slot has its own unit (or the unit reads a root from the environment):

```yaml
blue_green:
  slots:
    - path: /opt/api/a
      switch_command: systemctl start api-a && systemctl stop api-b
    - path: /opt/api/b
      switch_command: systemctl start api-b && systemctl stop api-a
  detect_command: systemctl is-active api-a >/dev/null 2>&1 && echo /opt/api/a || (systemctl is-active api-b >/dev/null 2>&1 && echo /opt/api/b || true)
```

Slot-level `switch_command` overrides the global one. Every slot must still resolve a command (own or global).

---

## Windows IIS (mapped drive vs local path)

Pablo writes over a mapped share (`W:\...`) while IIS `appcmd` reports the server-local path (`C:\...`). Use `key` for the IIS physical path:

```yaml
deploy:
  source:
    dir: ./build/CWS.Core
    exclude: ["*.map"]
  target_path: W:\celka-api-bg
  strategy: recreate
  blue_green:
    slots:
      - path: W:\celka-api-bg\celka-api-blue
        key: C:\Celka Web Production Servers\celka-api-bg\celka-api-blue
        switch_command: >
          powershell.exe -NoProfile -ExecutionPolicy Bypass -File ".\send-cmd.ps1"
          "CELKA\Administrator" "webserver"
          "& '$env:windir\System32\inetsrv\appcmd.exe' set vdir \"cws-api/\" /physicalPath:'C:\Celka Web Production Servers\celka-api-bg\celka-api-blue'"
      - path: W:\celka-api-bg\celka-api-green
        key: C:\Celka Web Production Servers\celka-api-bg\celka-api-green
        switch_command: >
          powershell.exe -NoProfile -ExecutionPolicy Bypass -File ".\send-cmd.ps1"
          "CELKA\Administrator" "webserver"
          "& '$env:windir\System32\inetsrv\appcmd.exe' set vdir \"cws-api/\" /physicalPath:'C:\Celka Web Production Servers\celka-api-bg\celka-api-green'"
    detect_command: >
      powershell.exe -NoProfile -ExecutionPolicy Bypass -File ".\send-cmd.ps1"
      "CELKA\Administrator" "webserver"
      "& '$env:windir\System32\inetsrv\appcmd.exe' list vdir 'cws-api/' /text:physicalPath"
```

---

## Environment variables

| Variable | Meaning |
|---|---|
| `PABLO_TARGET_SLOT` | Slot path written this run (becomes active after switch) |
| `PABLO_PREVIOUS_SLOT` | Previously active slot path; empty on first deploy |

These are injected into command execution only — they do not appear in `env_file` or `{{VAR}}` substitution. Slot `key` is match-only and is not exported as an environment variable.

---

## Manual rollback

There is no `pablo rollback` command. Re-run your switch logic with the previous slot. Example after a bad cutover when `.active` already points at the new slot:

```bash
PREV=$(cat /var/www/app/.active)
# If you still know the other path:
echo /var/www/app-blue > /var/www/app/.active
ln -sfn /var/www/app-blue /var/www/app/current
systemctl reload nginx
```

Or set `PABLO_TARGET_SLOT` to the previous path and re-run the same `switch_command` by hand.

---

## Validation rules

- Only `static` and `binary`
- Exactly two `slots`, distinct paths, none equal to `target_path`
- Effective detect keys (`key` if set, else `path`) must be distinct
- `detect_command` required
- Each slot must resolve a `switch_command`
- `strategy: backup` is rejected with `blue_green`

---

## Uninstall

`pablo uninstall` removes `target_path` and both slot directories. It does not undo your switch mechanism (symlinks, systemd units, proxy config).
