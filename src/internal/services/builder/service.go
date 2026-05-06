package builder

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Build(command, outputDir string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty build command")
	}

	head := parts[0]
	args := parts[1:]

	cmd := exec.Command(head, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}
	}

	return cmd.Run()
}
