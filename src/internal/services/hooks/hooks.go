package hooks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Execute(command string, workingDir string, envVars map[string]string) error {
	if command == "" {
		return nil
	}

	if workingDir != "" {
		fmt.Printf("Executing hook in %s: %s\n", workingDir, command)
	} else {
		fmt.Printf("Executing hook: %s\n", command)
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
		cmd = exec.Command("powershell", "-Command", command)
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
