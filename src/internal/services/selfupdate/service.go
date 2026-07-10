package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
)

type CheckResult struct {
	CurrentVersion string
	LatestVersion  string
	ReleaseTag     string
	UpToDate       bool
}

type Options struct {
	CurrentVersion string
	PinnedVersion  string
}

type Service struct {
	github *githubClient
}

func New() *Service {
	return &Service{github: newGitHubClient()}
}

func (s *Service) Check(opts Options) (*CheckResult, error) {
	assets, err := s.github.resolveAssets(opts.PinnedVersion)
	if err != nil {
		return nil, err
	}

	latest := normalizeVersion(assets.Tag)
	current := normalizeVersion(opts.CurrentVersion)

	return &CheckResult{
		CurrentVersion: current,
		LatestVersion:  latest,
		ReleaseTag:     assets.Tag,
		UpToDate:       latest == current,
	}, nil
}

func (s *Service) Update(opts Options) (*CheckResult, error) {
	check, err := s.Check(opts)
	if err != nil {
		return nil, err
	}
	if check.UpToDate {
		return check, nil
	}

	assets, err := s.github.resolveAssets(opts.PinnedVersion)
	if err != nil {
		return nil, err
	}

	tempPath, err := s.github.downloadVerifiedBinary(assets)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempPath)

	targetPath, err := currentExecutablePath()
	if err != nil {
		return nil, err
	}

	if err := ensureExecutableReplaceable(targetPath); err != nil {
		return nil, err
	}

	if err := replaceExecutable(targetPath, tempPath); err != nil {
		return nil, wrapReplaceError(targetPath, err)
	}

	return &CheckResult{
		CurrentVersion: check.LatestVersion,
		LatestVersion:  check.LatestVersion,
		ReleaseTag:     assets.Tag,
		UpToDate:       true,
	}, nil
}

func currentExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve current executable: %w", err)
	}

	resolved, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", fmt.Errorf("resolve executable symlinks: %w", err)
	}

	return filepath.Abs(resolved)
}

func PinnedVersionFromEnv() string {
	return os.Getenv("PABLO_VERSION")
}
