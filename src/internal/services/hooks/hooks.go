package hooks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"pablo/pkg/ui"
)

// Prefixed to Windows PowerShell -Command so native tool stdout (e.g. appcmd)
// is decoded/emitted as UTF-8 instead of the OEM code page (Turkish CP857 →
// mojibake like "de§iŸtirildi" for "değiştirildi" in UTF-8 terminals).
const windowsUTF8Preamble = "$OutputEncoding = [Console]::InputEncoding = [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false); "

func Execute(command string, workingDir string, envVars map[string]string) error {
	if command == "" {
		return nil
	}

	if ui.Verbose() && workingDir != "" {
		ui.Log("*", fmt.Sprintf("Hook cwd: %s", workingDir))
	}

	cmd := buildCommand(command, workingDir, envVars)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Capture runs a command and returns trimmed stdout. Stderr is passed through
// to the process stderr so diagnostics stay visible without polluting the result.
func Capture(command string, workingDir string, envVars map[string]string) (string, error) {
	if command == "" {
		return "", nil
	}

	if ui.Verbose() {
		if workingDir != "" {
			ui.Log("*", fmt.Sprintf("Hook cwd: %s", workingDir))
		}
		ui.Log(">", command)
	}

	cmd := buildCommand(command, workingDir, envVars)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return string(out), err
	}
	return string(bytes.TrimSpace(out)), nil
}

func buildCommand(command, workingDir string, envVars map[string]string) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-NoProfile", "-Command", windowsUTF8Preamble+command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	cmd.Env = os.Environ()
	for k, v := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	return cmd
}
