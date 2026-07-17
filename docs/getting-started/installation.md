# Installation

Install the Pablo CLI on your machine, then optionally add an editor extension for completion and validation.

---

## Requirements

| Component | When needed |
|-----------|-------------|
| **Go 1.25.5+** | Building from source only |
| **Git** | `git-sync` and `docker` deployment types |
| **Docker** | `docker` deployment type |
| **OpenSSH client** | Remote SSH deploys |

Supported hosts: Windows, macOS, and Linux.

---

## Option A — One-liner (recommended)

The installer downloads the latest release for your OS and architecture, verifies the SHA-256 checksum, and places the binary on a system path when permitted (otherwise a user path).

**Windows (PowerShell):**

```powershell
$s="$env:TEMP\pablo-install.ps1"; [Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; irm 'https://raw.githubusercontent.com/septillioner/pablo/master/scripts/install.ps1' -OutFile $s; powershell -NoProfile -ExecutionPolicy Bypass -File $s
```

Download the script to a temp file and run it in a clean PowerShell session. Avoid `irm ... | iex` — the pipe can pass null to `iex` on Windows PowerShell.

**Windows (cmd):**

```bat
curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/master/scripts/install.cmd -o install.cmd && install.cmd
```

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/master/scripts/install.sh | bash
```

**Pin a version:**

```powershell
# Windows
$env:PABLO_VERSION = "v1.4.0"
$s="$env:TEMP\pablo-install.ps1"; [Net.ServicePointManager]::SecurityProtocol=[Net.SecurityProtocolType]::Tls12; irm 'https://raw.githubusercontent.com/septillioner/pablo/master/scripts/install.ps1' -OutFile $s; powershell -NoProfile -ExecutionPolicy Bypass -File $s
```

```bash
# macOS / Linux
PABLO_VERSION=v1.4.0 curl -fsSL https://raw.githubusercontent.com/septillioner/pablo/master/scripts/install.sh | bash
```

**Install locations:**

| Platform | System (preferred) | User (fallback) |
|----------|-------------------|-----------------|
| Windows | `C:\Program Files\Pablo\pablo.exe` | `%LOCALAPPDATA%\Pablo\pablo.exe` |
| macOS / Linux | `/usr/local/bin/pablo` | `~/.local/bin/pablo` |

Verify with `pablo version`. To refresh an existing install later, run `pablo update` (or `pablo update --check` to only report a newer release). If another process holds the binary — for example `pablo lsp` from an editor — Pablo lists it and asks whether to close it before replacing the file.

---

## Option B — Pre-built binary

1. Open the [Releases page](https://github.com/septillioner/pablo/releases).
2. Download the asset for your OS and architecture:

   | Platform | Filename |
   |----------|----------|
   | macOS Intel | `pablo-darwin-amd64` |
   | macOS Apple Silicon | `pablo-darwin-arm64` |
   | Linux amd64 | `pablo-linux-amd64` |
   | Windows amd64 | `pablo-windows-amd64.exe` |
   | Windows arm64 | `pablo-windows-arm64.exe` |

3. Verify the SHA-256 checksum against `checksums.txt` in the release assets.
4. Move the binary onto your `PATH` (for example `/usr/local/bin/pablo` or `C:\Program Files\Pablo\pablo.exe`).
5. Run `pablo version`.

---

## Option C — Build from source

```bash
git clone https://github.com/septillioner/pablo.git
cd pablo

# Current platform → ./build/pablo[.exe]
./scripts/build.sh

# All platforms → ./build/pablo-{os}-{arch}[.exe]
./scripts/build.sh all
```

Run without installing:

```bash
cd src
go run main.go version
```

To install the built binary globally, use [Option A](#option-a--one-liner-recommended) or add `build/pablo` to your `PATH`.

---

## Editor extensions

For VS Code, install **Pablo** (`septillioner.pablo`) from the marketplace or a `.vsix` from [Releases](https://github.com/septillioner/pablo/releases). Ensure Pablo CLI 1.3+ with `pablo lsp` is on your PATH (or set `pablo.path`). Opening any `pablo.yaml` starts diagnostics and completion.

See [VS Code](../guides/vscode.md) and [Visual Studio](../guides/visual-studio.md) for binary resolution and troubleshooting.

---

## Next steps

- [Quick start](quick-start.md) — create and validate your first manifest
- [First deployment](first-deployment.md) — run a local static deploy
- [Examples](../examples/README.md) — sixteen copy-paste scenarios
