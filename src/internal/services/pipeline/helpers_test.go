package pipeline

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"pablo/internal/adapters/docker"
	"pablo/pkg/config"
	"pablo/pkg/domain"
	"pablo/internal/services/builder"
	"pablo/internal/services/deployer"
	"pablo/internal/services/scm"
)

func newTestPipeline() *Service {
	return New(config.NewLoader(), deployer.New(), builder.New(), scm.New(), docker.New())
}

func TestResolvePath(t *testing.T) {
	s := newTestPipeline()
	base := filepath.Join("C:", "proj")

	var absPath string
	if runtime.GOOS == "windows" {
		absPath = `C:\var\www`
	} else {
		absPath = "/var/www"
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{"empty", "", ""},
		{"absolute", absPath, absPath},
		{"relative", "dist", filepath.Join(base, "dist")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.resolvePath(base, tt.path); got != tt.want {
				t.Fatalf("resolvePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestResolveVariables(t *testing.T) {
	s := newTestPipeline()
	env := domain.Environment{
		EnvConfig: domain.EnvConfig{
			Variables: map[string]string{
				"APP_ENV": "production",
				"PORT":    "8080",
			},
		},
	}

	vars := s.resolveVariables(env)
	if len(vars) != 2 {
		t.Fatalf("len(vars) = %d, want 2", len(vars))
	}
	if vars["APP_ENV"] != "production" || vars["PORT"] != "8080" {
		t.Fatalf("unexpected vars: %v", vars)
	}

	empty := s.resolveVariables(domain.Environment{})
	if len(empty) != 0 {
		t.Fatalf("expected empty map, got %v", empty)
	}
}

func TestResolveArtifacts(t *testing.T) {
	s := newTestPipeline()
	baseDir := t.TempDir()
	env := domain.Environment{
		Deploy: domain.DeployConfig{
			Source: &domain.ArtifactsConfig{
				Dir:     "build",
				Include: []string{"app"},
				Exclude: []string{"tmp"},
			},
		},
	}

	artifactBase, include, exclude := s.resolveArtifacts(env, baseDir)
	wantBase := filepath.Join(baseDir, "build")
	if artifactBase != wantBase {
		t.Fatalf("artifactBase = %q, want %q", artifactBase, wantBase)
	}
	if len(include) != 1 || include[0] != "app" {
		t.Fatalf("include = %v, want [app]", include)
	}
	if len(exclude) != 1 || exclude[0] != "tmp" {
		t.Fatalf("exclude = %v, want [tmp]", exclude)
	}
}

func TestWriteEnvFile(t *testing.T) {
	s := newTestPipeline()
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", ".env")

	if err := s.writeEnvFile(path, map[string]string{"A": "1", "B": "two"}); err != nil {
		t.Fatalf("writeEnvFile: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "A=1") || !strings.Contains(content, "B=two") {
		t.Fatalf("unexpected content: %q", content)
	}

	emptyPath := filepath.Join(dir, "empty.env")
	if err := s.writeEnvFile(emptyPath, map[string]string{}); err != nil {
		t.Fatalf("writeEnvFile empty: %v", err)
	}
	emptyData, _ := os.ReadFile(emptyPath)
	if len(emptyData) != 0 {
		t.Fatalf("expected empty file, got %q", emptyData)
	}
}

func TestRunCommandsEmpty(t *testing.T) {
	s := newTestPipeline()
	if err := s.runCommands(nil, domain.Environment{}, false, &domain.Config{}, nil, ""); err != nil {
		t.Fatalf("empty commands: %v", err)
	}
}

func TestRunCommandsLocal(t *testing.T) {
	s := newTestPipeline()
	cmd := "echo ok"
	if runtime.GOOS == "windows" {
		cmd = "Write-Output ok"
	}
	if err := s.runCommands([]domain.DeployCommand{{Command: cmd}}, domain.Environment{}, false, &domain.Config{}, nil, ""); err != nil {
		t.Fatalf("local command: %v", err)
	}
}

func TestRunCommandsRemoteMissingHost(t *testing.T) {
	s := newTestPipeline()
	env := domain.Environment{
		Remote: &domain.RemoteConfig{
			Host:       "",
			Credential: "missing",
		},
		Deploy: domain.DeployConfig{
			TargetPath: "/tmp/target",
		},
	}
	err := s.runCommands([]domain.DeployCommand{{Command: "echo fail"}}, env, true, &domain.Config{}, nil, "")
	if err == nil {
		t.Fatal("expected error for remote SSH without valid host")
	}
}
