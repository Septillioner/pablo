# Pablo 🚀

**Pablo** is a professional, CLI-driven DevOps assistant designed for fast and automated deployments. It focuses on speed, security, and a sleek terminal experience.

## ✨ Philosophy
- **CLI First:** No GUI distractions. Fast, keyboard-driven management.
- **ASCII Aesthetics:** Professional karakter-based UI with a purple (Magenta) theme.
- **Lightweight:** Minimal footprint and sitem resources.
- **Extensible:** Easy to integrate into CI/CD pipelines.

## 🛠️ Visual Style
Pablo uses a dedicated ASCII-based UI system:
- `[+]` Success
- `[-]` Error
- `[!]` Warning
- `[*]` Information
- `[>]` Action/Wait

## 🚀 Getting Started

### Installation
1. Ensure you have **Go 1.25+** installed.
2. Clone the repository.
3. Run the build script:
   ```bash
   sh build.sh
   ```

### Development Environment
To use Pablo globally in your development terminal:
- **PowerShell:** Run `. .\pablo-dev.ps1` to add the build folder to your path.
- **New Session:** Run `.\pablo-shell.ps1` to open a dedicated Pablo shell.

## 📖 Commands

| Command | Description |
| :--- | :--- |
| `pablo init` | Initializes a sample `pablo_sample.yaml` in the current directory. |
| `pablo check` | Validates the syntax and paths of your manifest file. |
| `pablo run` | Executes the full deployment pipeline based on your manifest. |
| `pablo version` | Displays version and author information. |

## 🧩 Manifest Example (`pablo.yaml`)
```yaml
name: "my-app"
version: "1.2.0"
type: "static"

source:
  path: "."
  build_command: "npm run build"

artifacts:
  base_path: "./dist"
  include: ["*.js", "*.css", "index.html"]
  exclude: ["*.log"]

environments:
  production:
    target_path: "/var/www/html"
    strategy: "backup"
    variables:
      API_URL: "https://api.prod.com"
```

## 👤 Author
- **Ege İsmail Kösedag**
- **Email:** egeismailkosedag@gmail.com
- **Github:** [github.com/septillioner](https://github.com/septillioner)

---
*Built with ❤️ for DevOps engineers.*
