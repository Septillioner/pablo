package hooks

import (
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

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Inject environment variables
	cmd.Env = os.Environ()
	for k, v := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	return cmd.Run()
}
