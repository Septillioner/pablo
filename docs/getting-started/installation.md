# Installation

How to install the Pablo CLI and optional VS Code extension.

---

## Requirements

| Component | When needed |
|-----------|-------------|
| **Go 1.25.5+** | Building from source only |
| **Git** | `git-sync` and `docker` deployment types |
| **Docker** | `docker` deployment type |
| **OpenSSH client** | Remote SSH deploys |

Supported host platforms: Windows, macOS, Linux.

---

## Option A — One-liner (recommended)

Downloads the latest release binary for your OS and architecture, verifies the SHA-256 checksum, and installs to a system path when permitted (otherwise a user path).

**Windows (PowerShell):**

```powershell
$s="$env:TEMP\pablo-install.ps1"; [Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; irm 'https://raw.githubusercontent.com/septillioner/pablo/master/install.ps1' -OutFile $s; powershell -NoProfile -ExecutionPolicy Bypass -File $s
```

Downloads the installer to a temp file and runs it in a clean PowerShell session. Avoid `irm ... | iex` — the pipe can pass null to `iex` on Windows PowerShell.

**Windows (cmd):**

```bat
curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/master/install.cmd -o install.cmd && install.cmd
```

Run from **PowerShell**, not cmd.

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/master/install.sh | bash
```

**Pin a version:**

```powershell
# Windows
$env:PABLO_VERSION = "v1.4.0"
$s="$env:TEMP\pablo-install.ps1"; [Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; irm 'https://raw.githubusercontent.com/septillioner/pablo/master/install.ps1' -OutFile $s; powershell -NoProfile -ExecutionPolicy Bypass -File $s
```

```bash
# macOS / Linux
PABLO_VERSION=v1.4.0 curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/master/install.sh | bash
```

**Install locations:**

| Platform | System (preferred) | User (fallback) |
|----------|-------------------|-----------------|
| Windows | `C:\Program Files\Pablo\pablo.exe` | `%LOCALAPPDATA%\Pablo\pablo.exe` |
| macOS / Linux | `/usr/local/bin/pablo` | `~/.local/bin/pablo` |

Verify:

```bash
pablo version
```

**Update an existing install:**

```bash
pablo update
```

Checks GitHub Releases for the latest CLI binary for your OS, verifies the checksum, and replaces the running `pablo` executable. Use `pablo update --check` to see if a newer version exists without downloading.

If another process is using the same binary (for example `pablo lsp` from VS Code or Visual Studio), Pablo lists it and asks whether to close it before continuing the update.

---

## Option B — Pre-built binary

1. Open the [Releases page](https://github.com/septillioner/pablo/releases).
2. Download the file for your OS and architecture:

   | Platform | Filename |
   |----------|----------|
   | macOS Intel | `pablo-darwin-amd64` |
   | macOS Apple Silicon | `pablo-darwin-arm64` |
   | Linux amd64 | `pablo-linux-amd64` |
   | Windows amd64 | `pablo-windows-amd64.exe` |
   | Windows arm64 | `pablo-windows-arm64.exe` |

3. Verify the SHA-256 checksum against `checksums.txt` in the release assets.
4. Move the binary to a directory on your `PATH`:
   - macOS/Linux: `/usr/local/bin/pablo`
   - Windows: `C:\Program Files\Pablo\pablo.exe`
5. Verify:

```bash
pablo version
```

---

## Option C — Build from source

```bash
git clone https://github.com/septillioner/pablo.git
cd pablo

# Current platform → ./build/pablo[.exe]
./build.sh

# All platforms → ./build/pablo-{os}-{arch}[.exe]
./build.sh all
```

Run without installing:

```bash
cd src
go run main.go version
```

To install the built binary globally, use [Option A](installation.md#option-a--one-liner-recommended) or add `build/pablo` to your `PATH`.

---

## VS Code extension

1. Install from the marketplace: **Pablo** (`septillioner.pablo`), or install a `.vsix` from [Releases](https://github.com/septillioner/pablo/releases).
2. Ensure Pablo CLI **1.3+** with `pablo lsp` is on your PATH or set `pablo.path` in settings.
3. Open any `pablo.yaml` file — diagnostics and completion start automatically.

See [VS Code guide](../guides/vscode.md) for binary resolution and troubleshooting.

---

## Next steps

- [Quick start](quick-start.md) — create and validate your first manifest
- [First deployment](first-deployment.md) — run a local static deploy
