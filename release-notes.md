## Pablo v2.1.1

Patch release: 
fix: resolve StepRail and Spinner redraw conflict in Windows PowerShell

- Updated the StepRail implementation to prevent racing conditions with Spinner/ProgressBar redraws, ensuring smoother output in Windows PowerShell.
- Enhanced documentation to clarify the behavior of the sticky StepRail and its interaction with spinners and progress bars.


### Fixed

- **Step rail vs spinner** — Sticky StepRail pulse no longer races Spinner/ProgressBar `\r` redraws (common garble in Windows PowerShell). Live-line chrome is serialized; rail pulse pauses while a spinner or incomplete progress bar owns the line, and phase updates repaint the spinner after the rail redraw.


### Downloads

| Platform | File |
|----------|------|
| macOS (Intel) | pablo-darwin-amd64 |
| macOS (Apple Silicon) | pablo-darwin-arm64 |
| Linux (amd64) | pablo-linux-amd64 |
| Windows (amd64) | pablo-windows-amd64.exe |
| Windows (arm64) | pablo-windows-arm64.exe |

Verify downloads with checksums.txt (SHA-256).
