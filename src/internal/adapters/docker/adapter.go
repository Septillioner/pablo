package docker

import (
	"os"
	"os/exec"
)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ComposeUp(composeFile string, build bool, targetPath string) error {
	args := []string{"compose", "-f", composeFile, "up", "-d"}
	if build {
		args = append(args, "--build")
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = targetPath // Run in target directory where .env might be
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (a *Adapter) ComposeDown(composeFile string, targetPath string) error {
	cmd := exec.Command("docker", "compose", "-f", composeFile, "down")
	cmd.Dir = targetPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
