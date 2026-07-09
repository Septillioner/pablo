package scm

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"pablo/pkg/domain"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}
}

func initTempRepo(t *testing.T) string {
	t.Helper()
	requireGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@pablo.local")
	runGit(t, dir, "config", "user.name", "Pablo Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "init")
	runGit(t, dir, "branch", "-M", "main")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func TestCloneOrPullNilConfig(t *testing.T) {
	s := New()
	err := s.CloneOrPull(nil, t.TempDir())
	if err == nil {
		t.Fatal("expected error for nil config")
	}
	if !strings.Contains(err.Error(), "git configuration is missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloneOrPullInvalidURL(t *testing.T) {
	requireGit(t)
	s := New()
	target := filepath.Join(t.TempDir(), "clone-target")
	err := s.CloneOrPull(&domain.GitConfig{
		Repo:   "https://invalid.invalid.example/repo.git",
		Branch: "main",
	}, target)
	if err == nil {
		t.Fatal("expected clone error for invalid URL")
	}
	if !strings.Contains(err.Error(), "git clone failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloneOrPullCloneAndPull(t *testing.T) {
	repoDir := initTempRepo(t)
	s := New()
	target := filepath.Join(t.TempDir(), "work")
	cfg := &domain.GitConfig{Repo: repoDir, Branch: "main"}

	if err := s.CloneOrPull(cfg, target); err != nil {
		t.Fatalf("clone: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".git")); err != nil {
		t.Fatalf("missing .git after clone: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "README.md")); err != nil {
		t.Fatalf("missing README after clone: %v", err)
	}

	if err := s.CloneOrPull(cfg, target); err != nil {
		t.Fatalf("pull: %v", err)
	}
}
