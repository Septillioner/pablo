# Exit Codes

Pablo uses a simple exit code scheme today.

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Any failure — validation error, missing profile/environment, pipeline error, I/O error |

All error paths in the CLI call `os.Exit(1)`. There is no separate validation exit code.

**Examples of exit code `1`:**

- `pablo check` finds schema violations
- `pablo run` references a profile or environment that does not exist
- Build command fails
- SSH connection or remote command fails
- Deploy to a protected system path without `--force`

**Scripting tip:** rely on exit code only; parse stderr/stdout for error details. For structured manifest data, use `pablo inspect --json` (exits `0` on success).

`pablo update --check` exits `1` when a newer release is available (even though the check itself succeeded).
