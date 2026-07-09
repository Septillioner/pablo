package docker

import (
	"os"
	"os/exec"
)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func composeUpArgs(composeFile string, build bool) []string {
	args := []string{"compose", "-f", composeFile, "up", "-d"}
	if build {
		args = append(args, "--build")
	}
	return args
}

func composeDownArgs(composeFile string) []string {
	return []string{"compose", "-f", composeFile, "down"}
}

func (a *Adapter) ComposeUp(composeFile string, build bool, targetPath string) error {
	cmd := exec.Command("docker", composeUpArgs(composeFile, build)...)
	cmd.Dir = targetPath // Run in target directory where .env might be
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (a *Adapter) ComposeDown(composeFile string, targetPath string) error {
	cmd := exec.Command("docker", composeDownArgs(composeFile)...)
	cmd.Dir = targetPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
