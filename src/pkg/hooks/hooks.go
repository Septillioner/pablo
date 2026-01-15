package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Execute(command string) error {
	if command == "" {
		return nil
	}

	fmt.Printf("Executing hook: %s\n", command)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
