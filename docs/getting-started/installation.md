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

## Option A — Pre-built binary (recommended)

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

## Option B — Build from source

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

---

## Option C — Self-deploy

Pablo can build and register itself in your `PATH` using its own manifest ([pablo.yaml](../pablo.yaml)):

```bash
# macOS / Linux
./publish-self.sh

# Windows (PowerShell, elevated)
./publish-self.ps1
```

After self-deploy, `pablo` is available globally.

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
