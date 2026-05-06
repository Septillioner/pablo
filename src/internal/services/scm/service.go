package scm

import (
	"fmt"
	"os"
	"os/exec"
	"pablo/internal/domain"
	"pablo/pkg/ui"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) CloneOrPull(config *domain.GitConfig, targetPath string) error {
	if config == nil {
		return fmt.Errorf("git configuration is missing")
	}

	// 0. Mark directory as safe (Fixes dubious ownership in root/sudo environments)
	ui.Log("*", "Marking target directory as safe in Git configuration")
	safeCmd := exec.Command("git", "config", "--global", "--add", "safe.directory", targetPath)
	_ = safeCmd.Run() // We ignore errors here as it might already be set

	// 1. Check if targetPath/.git exists
	gitDir := fmt.Sprintf("%s/.git", targetPath)
	if _, err := os.Stat(gitDir); err == nil {
		// Repo exists, do pull
		ui.Log(">", fmt.Sprintf("Repository exists, performing git pull in %s", targetPath))
		cmd := exec.Command("git", "pull", "origin", config.Branch)
		cmd.Dir = targetPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git pull failed: %w", err)
		}
	} else {
		// Repo doesn't exist, do clone
		ui.Log(">", fmt.Sprintf("Cloning %s to %s", config.Repo, targetPath))

		// Ensure parent directory exists
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}

		cmd := exec.Command("git", "clone", "-b", config.Branch, config.Repo, targetPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
	}

	return nil
}
