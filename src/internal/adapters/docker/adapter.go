package docker

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"pablo/pkg/ui"
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

func composePsQuietArgs(composeFile string) []string {
	return []string{"compose", "-f", composeFile, "ps", "-q"}
}

func (a *Adapter) ComposeUp(composeFile string, build bool, targetPath string) error {
	cmd := exec.Command("docker", composeUpArgs(composeFile, build)...)
	cmd.Dir = targetPath // Run in target directory where .env might be
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return ui.WithExternalOutput(cmd.Run)
}

func (a *Adapter) ComposeDown(composeFile string, targetPath string) error {
	cmd := exec.Command("docker", composeDownArgs(composeFile)...)
	cmd.Dir = targetPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return ui.WithExternalOutput(cmd.Run)
}

// ComposePsRunning reports whether any containers for the Compose project are present.
func (a *Adapter) ComposePsRunning(composeFile string, targetPath string) (bool, error) {
	cmd := exec.Command("docker", composePsQuietArgs(composeFile)...)
	cmd.Dir = targetPath
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return false, err
	}
	return strings.TrimSpace(stdout.String()) != "", nil
}
